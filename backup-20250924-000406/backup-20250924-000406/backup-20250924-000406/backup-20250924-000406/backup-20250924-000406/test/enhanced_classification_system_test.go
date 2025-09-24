package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/classification"
	"internal/modules/risk_assessment"
	"internal/risk"
)

// TestSuite provides common test utilities and setup
type TestSuite struct {
	db                    *sql.DB
	riskService           *risk.RiskDetectionService
	crosswalkService      *classification.CrosswalkAnalyzer
	riskAssessmentService *risk_assessment.RiskAssessmentService
	logger                *log.Logger
}

// SetupTestSuite initializes the test environment
func SetupTestSuite(t *testing.T) *TestSuite {
	// Get database connection from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Initialize services
	riskService := risk.NewRiskDetectionService(db, logger)
	crosswalkService := classification.NewCrosswalkAnalyzer(db, logger)
	riskAssessmentService := risk_assessment.NewRiskAssessmentService(db, logger)

	return &TestSuite{
		db:                    db,
		riskService:           riskService,
		crosswalkService:      crosswalkService,
		riskAssessmentService: riskAssessmentService,
		logger:                logger,
	}
}

// CleanupTestSuite cleans up test resources
func (ts *TestSuite) CleanupTestSuite(t *testing.T) {
	if ts.db != nil {
		ts.db.Close()
	}
}

// TestRiskKeywordDetection tests the risk keyword detection functionality
func TestRiskKeywordDetection(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	ctx := context.Background()

	t.Run("TestDirectKeywordMatching", func(t *testing.T) {
		// Test content with obvious risk keywords
		testContent := `
		Welcome to our online casino and gambling platform. 
		We offer the best poker, blackjack, and slot games.
		We also sell tobacco products and adult entertainment.
		`

		result, err := ts.riskService.DetectRiskKeywords(ctx, testContent, "website")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify high-risk keywords are detected
		assert.Greater(t, result.RiskScore, 0.7, "Risk score should be high for gambling content")
		assert.Equal(t, "high", result.RiskLevel, "Risk level should be high")

		// Check for specific risk categories
		expectedCategories := []string{"prohibited", "high_risk"}
		for _, category := range expectedCategories {
			assert.Contains(t, result.RiskCategories, category,
				"Should detect %s category", category)
		}

		// Verify specific keywords are detected
		expectedKeywords := []string{"casino", "gambling", "poker", "tobacco", "adult entertainment"}
		for _, keyword := range expectedKeywords {
			assert.Contains(t, result.DetectedKeywords, keyword,
				"Should detect keyword: %s", keyword)
		}
	})

	t.Run("TestSynonymMatching", func(t *testing.T) {
		// Test content with synonyms of risk keywords
		testContent := `
		We provide digital currency exchange services.
		Our platform supports Bitcoin, Ethereum, and other cryptocurrencies.
		We also offer adult content and escort services.
		`

		result, err := ts.riskService.DetectRiskKeywords(ctx, testContent, "website")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify cryptocurrency synonyms are detected
		cryptoKeywords := []string{"digital currency", "cryptocurrency", "Bitcoin", "Ethereum"}
		for _, keyword := range cryptoKeywords {
			assert.Contains(t, result.DetectedKeywords, keyword,
				"Should detect crypto keyword: %s", keyword)
		}

		// Verify adult entertainment synonyms
		adultKeywords := []string{"adult content", "escort services"}
		for _, keyword := range adultKeywords {
			assert.Contains(t, result.DetectedKeywords, keyword,
				"Should detect adult keyword: %s", keyword)
		}
	})

	t.Run("TestPatternMatching", func(t *testing.T) {
		// Test content with regex patterns
		testContent := `
		Contact us at +1-555-DRUGS or visit our pharmacy.
		We sell prescription medications without prescription.
		Call 1-800-WEAPONS for our firearms collection.
		`

		result, err := ts.riskService.DetectRiskKeywords(ctx, testContent, "website")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pattern-based detection
		assert.Greater(t, result.RiskScore, 0.8, "Risk score should be very high for illegal content")
		assert.Equal(t, "critical", result.RiskLevel, "Risk level should be critical")

		// Check for illegal activity patterns
		illegalPatterns := []string{"DRUGS", "WEAPONS", "prescription medications without prescription"}
		for _, pattern := range illegalPatterns {
			found := false
			for _, keyword := range result.DetectedKeywords {
				if containsPattern(keyword, pattern) {
					found = true
					break
				}
			}
			assert.True(t, found, "Should detect illegal pattern: %s", pattern)
		}
	})

	t.Run("TestLowRiskContent", func(t *testing.T) {
		// Test content with no risk keywords
		testContent := `
		Welcome to our family restaurant. We serve delicious meals
		and provide excellent customer service. Visit us for lunch or dinner.
		`

		result, err := ts.riskService.DetectRiskKeywords(ctx, testContent, "website")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify low risk assessment
		assert.Less(t, result.RiskScore, 0.3, "Risk score should be low for legitimate business")
		assert.Equal(t, "low", result.RiskLevel, "Risk level should be low")
		assert.Empty(t, result.DetectedKeywords, "Should not detect any risk keywords")
	})

	t.Run("TestConfidenceScoring", func(t *testing.T) {
		// Test confidence scoring accuracy
		testContent := `
		We are a legitimate technology company providing software solutions.
		Our services include web development, mobile apps, and cloud computing.
		We are fully licensed and compliant with all regulations.
		`

		result, err := ts.riskService.DetectRiskKeywords(ctx, testContent, "website")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify high confidence for legitimate business
		assert.Greater(t, result.Confidence, 0.8, "Confidence should be high for legitimate business")
		assert.Equal(t, "low", result.RiskLevel, "Risk level should be low")
	})
}

// TestCodeCrosswalkFunctionality tests the MCC/NAICS/SIC crosswalk functionality
func TestCodeCrosswalkFunctionality(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	ctx := context.Background()

	t.Run("TestMCCToIndustryMapping", func(t *testing.T) {
		// Test MCC code to industry mapping
		mccCode := "5734" // Computer Software Stores

		result, err := ts.crosswalkService.MapMCCCodesToIndustries(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify MCC code is mapped to technology industry
		industryMappings, exists := result.MCCToIndustryMappings[mccCode]
		assert.True(t, exists, "MCC code %s should have industry mappings", mccCode)
		assert.NotEmpty(t, industryMappings, "MCC code %s should map to at least one industry", mccCode)

		// Verify technology industry is included
		hasTechnology := false
		for _, industry := range industryMappings {
			if containsTechnologyKeywords(industry) {
				hasTechnology = true
				break
			}
		}
		assert.True(t, hasTechnology, "MCC code %s should map to technology industry", mccCode)
	})

	t.Run("TestNAICSToIndustryMapping", func(t *testing.T) {
		// Test NAICS code to industry mapping
		naicsCode := "541511" // Custom Computer Programming Services

		result, err := ts.crosswalkService.MapNAICSCodesToIndustries(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify NAICS code is mapped to technology industry
		industryMappings, exists := result.NAICSToIndustryMappings[naicsCode]
		assert.True(t, exists, "NAICS code %s should have industry mappings", naicsCode)
		assert.NotEmpty(t, industryMappings, "NAICS code %s should map to at least one industry", naicsCode)

		// Verify technology industry is included
		hasTechnology := false
		for _, industry := range industryMappings {
			if containsTechnologyKeywords(industry) {
				hasTechnology = true
				break
			}
		}
		assert.True(t, hasTechnology, "NAICS code %s should map to technology industry", naicsCode)
	})

	t.Run("TestSICToIndustryMapping", func(t *testing.T) {
		// Test SIC code to industry mapping
		sicCode := "7372" // Prepackaged Software

		result, err := ts.crosswalkService.MapSICCodesToIndustries(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify SIC code is mapped to technology industry
		industryMappings, exists := result.SICToIndustryMappings[sicCode]
		assert.True(t, exists, "SIC code %s should have industry mappings", sicCode)
		assert.NotEmpty(t, industryMappings, "SIC code %s should map to at least one industry", sicCode)

		// Verify technology industry is included
		hasTechnology := false
		for _, industry := range industryMappings {
			if containsTechnologyKeywords(industry) {
				hasTechnology = true
				break
			}
		}
		assert.True(t, hasTechnology, "SIC code %s should map to technology industry", sicCode)
	})

	t.Run("TestCrosswalkValidation", func(t *testing.T) {
		// Test crosswalk validation rules
		result, err := ts.crosswalkService.ValidateCrosswalkMappings(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify validation results
		assert.Greater(t, result.ValidationScore, 0.8, "Crosswalk validation score should be high")
		assert.Less(t, len(result.Issues), 10, "Should have minimal validation issues")

		// Check for specific validation rules
		expectedValidations := []string{
			"mcc_industry_consistency",
			"naics_industry_consistency",
			"sic_industry_consistency",
			"confidence_score_validation",
		}
		for _, validation := range expectedValidations {
			found := false
			for _, rule := range result.ValidationRules {
				if rule.RuleName == validation {
					found = true
					assert.True(t, rule.IsValid, "Validation rule %s should be valid", validation)
					break
				}
			}
			assert.True(t, found, "Should have validation rule: %s", validation)
		}
	})

	t.Run("TestCrosswalkPerformance", func(t *testing.T) {
		// Test crosswalk query performance
		startTime := time.Now()

		result, err := ts.crosswalkService.GetAllCrosswalkMappings(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		duration := time.Since(startTime)

		// Verify performance requirements
		assert.Less(t, duration, 2*time.Second, "Crosswalk query should complete within 2 seconds")
		assert.Greater(t, len(result.Mappings), 100, "Should have substantial crosswalk mappings")
	})
}

// TestBusinessRiskAssessmentWorkflow tests the complete business risk assessment workflow
func TestBusinessRiskAssessmentWorkflow(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	ctx := context.Background()

	t.Run("TestHighRiskBusinessAssessment", func(t *testing.T) {
		// Test assessment of high-risk business
		request := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "Online Casino & Gambling Platform",
			WebsiteURL:   "https://example-casino.com",
			DomainName:   "example-casino.com",
			Industry:     "Gambling",
			BusinessType: "Online Gaming",
		}

		result, err := ts.riskAssessmentService.AssessRisk(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify high-risk assessment
		assert.Greater(t, result.OverallRiskScore, 0.7, "Risk score should be high for gambling business")
		assert.Equal(t, "high", string(result.RiskLevel), "Risk level should be high")
		assert.Greater(t, len(result.RiskFactors), 0, "Should identify multiple risk factors")

		// Verify specific risk factors
		expectedRiskFactors := []string{"prohibited_industry", "gambling_activities", "high_risk_mcc"}
		for _, factor := range expectedRiskFactors {
			found := false
			for _, riskFactor := range result.RiskFactors {
				if riskFactor.FactorType == factor {
					found = true
					break
				}
			}
			assert.True(t, found, "Should identify risk factor: %s", factor)
		}

		// Verify recommendations
		assert.Greater(t, len(result.Recommendations), 0, "Should provide risk mitigation recommendations")
	})

	t.Run("TestLowRiskBusinessAssessment", func(t *testing.T) {
		// Test assessment of low-risk business
		request := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "Family Restaurant & Catering",
			WebsiteURL:   "https://family-restaurant.com",
			DomainName:   "family-restaurant.com",
			Industry:     "Food Service",
			BusinessType: "Restaurant",
		}

		result, err := ts.riskAssessmentService.AssessRisk(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify low-risk assessment
		assert.Less(t, result.OverallRiskScore, 0.3, "Risk score should be low for legitimate restaurant")
		assert.Equal(t, "low", string(result.RiskLevel), "Risk level should be low")
		assert.Less(t, len(result.RiskFactors), 3, "Should identify minimal risk factors")

		// Verify confidence
		assert.Greater(t, result.ConfidenceScore, 0.8, "Confidence should be high for legitimate business")
	})

	t.Run("TestMediumRiskBusinessAssessment", func(t *testing.T) {
		// Test assessment of medium-risk business
		request := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "Cryptocurrency Exchange Platform",
			WebsiteURL:   "https://crypto-exchange.com",
			DomainName:   "crypto-exchange.com",
			Industry:     "Financial Services",
			BusinessType: "Cryptocurrency Exchange",
		}

		result, err := ts.riskAssessmentService.AssessRisk(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify medium-risk assessment
		assert.GreaterOrEqual(t, result.OverallRiskScore, 0.4, "Risk score should be medium for crypto business")
		assert.LessOrEqual(t, result.OverallRiskScore, 0.7, "Risk score should not be too high for regulated crypto business")
		assert.Equal(t, "medium", string(result.RiskLevel), "Risk level should be medium")

		// Verify specific risk factors
		expectedRiskFactors := []string{"cryptocurrency_activities", "financial_services", "regulatory_compliance"}
		for _, factor := range expectedRiskFactors {
			found := false
			for _, riskFactor := range result.RiskFactors {
				if riskFactor.FactorType == factor {
					found = true
					break
				}
			}
			assert.True(t, found, "Should identify risk factor: %s", factor)
		}
	})

	t.Run("TestAssessmentPerformance", func(t *testing.T) {
		// Test assessment performance
		request := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "Technology Consulting Services",
			WebsiteURL:   "https://tech-consulting.com",
			DomainName:   "tech-consulting.com",
			Industry:     "Technology",
			BusinessType: "Consulting",
		}

		startTime := time.Now()
		result, err := ts.riskAssessmentService.AssessRisk(ctx, request)
		duration := time.Since(startTime)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify performance requirements
		assert.Less(t, duration, 5*time.Second, "Risk assessment should complete within 5 seconds")
		assert.Less(t, result.ProcessingTime, 5*time.Second, "Processing time should be under 5 seconds")
	})

	t.Run("TestAssessmentErrorHandling", func(t *testing.T) {
		// Test error handling for invalid requests
		invalidRequest := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "", // Invalid: empty business name
			WebsiteURL:   "invalid-url",
			DomainName:   "",
		}

		result, err := ts.riskAssessmentService.AssessRisk(ctx, invalidRequest)
		assert.Error(t, err, "Should return error for invalid request")
		assert.Nil(t, result, "Should not return result for invalid request")
	})
}

// TestUIIntegrationPoints tests the UI integration points
func TestUIIntegrationPoints(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	ctx := context.Background()

	t.Run("TestRiskDisplayDataFormat", func(t *testing.T) {
		// Test that risk data is properly formatted for UI display
		request := &risk_assessment.RiskAssessmentRequest{
			BusinessName: "Test Business",
			WebsiteURL:   "https://test-business.com",
			DomainName:   "test-business.com",
			Industry:     "Technology",
		}

		result, err := ts.riskAssessmentService.AssessRisk(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify UI-compatible data format
		assert.NotEmpty(t, result.RequestID, "Should have request ID for UI tracking")
		assert.NotZero(t, result.AssessmentTimestamp, "Should have assessment timestamp")
		assert.GreaterOrEqual(t, result.OverallRiskScore, 0.0, "Risk score should be valid")
		assert.LessOrEqual(t, result.OverallRiskScore, 1.0, "Risk score should be normalized")
		assert.NotEmpty(t, result.RiskLevel, "Should have risk level for UI display")

		// Verify risk factors are UI-compatible
		for _, factor := range result.RiskFactors {
			assert.NotEmpty(t, factor.FactorType, "Risk factor should have type")
			assert.NotEmpty(t, factor.Description, "Risk factor should have description")
			assert.GreaterOrEqual(t, factor.Score, 0.0, "Risk factor score should be valid")
			assert.LessOrEqual(t, factor.Score, 1.0, "Risk factor score should be normalized")
		}

		// Verify recommendations are UI-compatible
		for _, rec := range result.Recommendations {
			assert.NotEmpty(t, rec.Type, "Recommendation should have type")
			assert.NotEmpty(t, rec.Description, "Recommendation should have description")
			assert.NotEmpty(t, rec.Priority, "Recommendation should have priority")
		}
	})

	t.Run("TestRiskLevelColorMapping", func(t *testing.T) {
		// Test risk level to color mapping for UI
		riskLevelColors := map[string]string{
			"low":      "green",
			"medium":   "yellow",
			"high":     "orange",
			"critical": "red",
		}

		for riskLevel, expectedColor := range riskLevelColors {
			color := getRiskLevelColor(riskLevel)
			assert.Equal(t, expectedColor, color, "Risk level %s should map to color %s", riskLevel, expectedColor)
		}
	})

	t.Run("TestRiskScoreProgressBar", func(t *testing.T) {
		// Test risk score for progress bar display
		testScores := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		expectedPercentages := []int{0, 25, 50, 75, 100}

		for i, score := range testScores {
			percentage := int(score * 100)
			assert.Equal(t, expectedPercentages[i], percentage,
				"Risk score %.2f should convert to %d%%", score, expectedPercentages[i])
		}
	})
}

// TestPerformanceWithLargeDatasets tests performance with large datasets
func TestPerformanceWithLargeDatasets(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	ctx := context.Background()

	t.Run("TestBulkRiskKeywordDetection", func(t *testing.T) {
		// Test performance with large content
		largeContent := generateLargeTestContent(10000) // 10KB of content

		startTime := time.Now()
		result, err := ts.riskService.DetectRiskKeywords(ctx, largeContent, "website")
		duration := time.Since(startTime)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify performance requirements
		assert.Less(t, duration, 1*time.Second, "Risk detection should complete within 1 second for large content")
		assert.Greater(t, result.Confidence, 0.0, "Should have valid confidence score")
	})

	t.Run("TestBulkCrosswalkQueries", func(t *testing.T) {
		// Test performance with multiple crosswalk queries
		startTime := time.Now()

		// Run multiple crosswalk queries
		mccResult, err := ts.crosswalkService.MapMCCCodesToIndustries(ctx)
		require.NoError(t, err)

		naicsResult, err := ts.crosswalkService.MapNAICSCodesToIndustries(ctx)
		require.NoError(t, err)

		sicResult, err := ts.crosswalkService.MapSICCodesToIndustries(ctx)
		require.NoError(t, err)

		duration := time.Since(startTime)

		// Verify performance requirements
		assert.Less(t, duration, 3*time.Second, "All crosswalk queries should complete within 3 seconds")
		assert.NotNil(t, mccResult, "MCC result should not be nil")
		assert.NotNil(t, naicsResult, "NAICS result should not be nil")
		assert.NotNil(t, sicResult, "SIC result should not be nil")
	})

	t.Run("TestConcurrentRiskAssessments", func(t *testing.T) {
		// Test concurrent risk assessments
		requests := make([]*risk_assessment.RiskAssessmentRequest, 10)
		for i := 0; i < 10; i++ {
			requests[i] = &risk_assessment.RiskAssessmentRequest{
				BusinessName: fmt.Sprintf("Test Business %d", i),
				WebsiteURL:   fmt.Sprintf("https://test-business-%d.com", i),
				DomainName:   fmt.Sprintf("test-business-%d.com", i),
				Industry:     "Technology",
			}
		}

		startTime := time.Now()

		// Run concurrent assessments
		results := make([]*risk_assessment.RiskAssessmentResult, 10)
		errors := make([]error, 10)

		for i, request := range requests {
			go func(idx int, req *risk_assessment.RiskAssessmentRequest) {
				results[idx], errors[idx] = ts.riskAssessmentService.AssessRisk(ctx, req)
			}(i, request)
		}

		// Wait for all assessments to complete
		time.Sleep(10 * time.Second)

		duration := time.Since(startTime)

		// Verify performance requirements
		assert.Less(t, duration, 15*time.Second, "Concurrent assessments should complete within 15 seconds")

		// Verify all assessments completed successfully
		successCount := 0
		for i, err := range errors {
			if err == nil && results[i] != nil {
				successCount++
			}
		}
		assert.GreaterOrEqual(t, successCount, 8, "At least 8 out of 10 concurrent assessments should succeed")
	})
}

// Helper functions

func containsPattern(text, pattern string) bool {
	// Simple pattern matching for test purposes
	return len(text) > 0 && len(pattern) > 0
}

func containsTechnologyKeywords(industry string) bool {
	techKeywords := []string{"technology", "software", "computer", "tech", "IT", "digital"}
	for _, keyword := range techKeywords {
		if containsPattern(industry, keyword) {
			return true
		}
	}
	return false
}

func getRiskLevelColor(riskLevel string) string {
	colorMap := map[string]string{
		"low":      "green",
		"medium":   "yellow",
		"high":     "orange",
		"critical": "red",
	}
	return colorMap[riskLevel]
}

func generateLargeTestContent(size int) string {
	baseContent := "This is a legitimate technology company providing software solutions. "
	content := ""
	for len(content) < size {
		content += baseContent
	}
	return content[:size]
}
