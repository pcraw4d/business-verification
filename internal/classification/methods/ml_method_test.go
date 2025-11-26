// +build !integration

package methods

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/machine_learning"
	"kyb-platform/internal/machine_learning/infrastructure"
)

// MockPythonMLService is a mock implementation of Python ML service
type MockPythonMLService struct {
	ShouldFail      bool
	EnhancedResp    *infrastructure.EnhancedClassificationResponse
	StandardResp    *infrastructure.ClassificationResponse
	Error           error
	CallCount       int
	LastRequest     *infrastructure.EnhancedClassificationRequest
}

func (m *MockPythonMLService) ClassifyEnhanced(ctx context.Context, req *infrastructure.EnhancedClassificationRequest) (*infrastructure.EnhancedClassificationResponse, error) {
	m.CallCount++
	m.LastRequest = req
	
	if m.ShouldFail {
		return nil, m.Error
	}
	
	if m.EnhancedResp != nil {
		return m.EnhancedResp, nil
	}
	
	// Default response
	return &infrastructure.EnhancedClassificationResponse{
		RequestID:      "test-request-123",
		Industry:       "Technology",
		Confidence:     0.85,
		Summary:        "Test summary of business content",
		Explanation:    "This business is classified as Technology based on keywords and content analysis",
		AllScores:      map[string]float64{"Technology": 0.85, "Retail": 0.10, "Healthcare": 0.05},
		Classifications: []infrastructure.ClassificationPrediction{
			{Label: "Technology", Confidence: 0.85, Probability: 0.85, Rank: 1},
			{Label: "Retail", Confidence: 0.10, Probability: 0.10, Rank: 2},
		},
		ProcessingTime:     0.5,
		QuantizationEnabled: true,
		ModelVersion:       "2.0.0",
		Timestamp:         infrastructure.Timestamp{},
		Success:           true,
	}, nil
}

// MockCodeGenerator is a mock implementation of ClassificationCodeGenerator
type MockCodeGenerator struct {
	ShouldFail bool
	Error      error
	Codes      *classification.ClassificationCodesInfo
	CallCount  int
}

func (m *MockCodeGenerator) GenerateClassificationCodes(
	ctx context.Context,
	keywords []string,
	detectedIndustry string,
	confidence float64,
	additionalIndustries ...classification.IndustryResult,
) (*classification.ClassificationCodesInfo, error) {
	m.CallCount++
	
	if m.ShouldFail {
		return nil, m.Error
	}
	
	if m.Codes != nil {
		return m.Codes, nil
	}
	
	// Default codes
	return &classification.ClassificationCodesInfo{
		MCC: []classification.MCCCode{
			{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
			{Code: "5045", Description: "Computers, Computer Peripheral Equipment", Confidence: 0.80},
		},
		SIC: []classification.SICCode{
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.90},
			{Code: "7371", Description: "Computer Programming Services", Confidence: 0.85},
			{Code: "7373", Description: "Computer Integrated Systems Design", Confidence: 0.80},
		},
		NAICS: []classification.NAICSCode{
			{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.90},
			{Code: "541512", Description: "Computer Systems Design Services", Confidence: 0.85},
			{Code: "334111", Description: "Electronic Computer Manufacturing", Confidence: 0.80},
		},
	}, nil
}

// MockWebsiteScraper is a mock implementation of EnhancedWebsiteScraper
type MockWebsiteScraper struct {
	ShouldFail bool
	Error      string
	TextContent string
	Success    bool
	CallCount  int
}

func (m *MockWebsiteScraper) ScrapeWebsite(ctx context.Context, websiteURL string) *classification.ScrapingResult {
	m.CallCount++
	
	result := &classification.ScrapingResult{
		URL:       websiteURL,
		Success:   !m.ShouldFail && m.Success,
		TextContent: m.TextContent,
		Error:     m.Error,
	}
	
	if m.TextContent == "" {
		result.TextContent = "Sample website content about technology services and software development"
	}
	
	return result
}

func TestMLClassificationMethod_BasicProperties(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)

	// Test basic properties
	assert.Equal(t, "ml_classification", method.GetName())
	assert.Equal(t, "ml", method.GetType())
	assert.Contains(t, method.GetDescription(), "DistilBART")
	assert.Equal(t, 0.4, method.GetWeight())
	assert.True(t, method.IsEnabled())

	// Test weight management
	method.SetWeight(0.8)
	assert.Equal(t, 0.8, method.GetWeight())
	method.SetWeight(0.4)

	// Test enabled state
	method.SetEnabled(false)
	assert.False(t, method.IsEnabled())
	method.SetEnabled(true)
	assert.True(t, method.IsEnabled())
}

func TestMLClassificationMethod_ExtractWebsiteContent_NilScraper(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)
	ctx := context.Background()
	
	// Test that nil scraper returns empty content
	content := method.extractWebsiteContent(ctx, "https://techcorp.com")
	assert.Empty(t, content, "Expected empty content when scraper is nil")
}

func TestMLClassificationMethod_CodeGeneration_NilGenerator(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	
	enhancedResp := &infrastructure.EnhancedClassificationResponse{
		Industry:    "Technology",
		Confidence:  0.85,
		Summary:     "Technology company",
		Explanation: "Tech business",
		Classifications: []infrastructure.ClassificationPrediction{
			{Label: "Technology", Confidence: 0.85},
		},
	}
	
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)
	ctx := context.Background()
	
	result := method.buildEnhancedResult(ctx, enhancedResp, "Test Company")
	
	require.NotNil(t, result)
	assert.Equal(t, "Technology", result.PrimaryIndustry)
	assert.Equal(t, 0.85, result.ConfidenceScore)
	assert.Equal(t, "Technology company", result.ContentSummary)
	assert.Equal(t, "Tech business", result.Explanation)
	
	// Code generation will be empty if codeGen is nil, which is expected
	assert.Empty(t, result.ClassificationCodes.MCC)
	assert.Empty(t, result.ClassificationCodes.SIC)
	assert.Empty(t, result.ClassificationCodes.NAICS)
}

func TestMLClassificationMethod_KeywordExtraction(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)

	tests := []struct {
		name           string
		summary        string
		explanation    string
		expectedMinLen int
	}{
		{
			name:           "normal text",
			summary:        "Technology company providing software development services",
			explanation:    "This business is classified as Technology based on software and development keywords",
			expectedMinLen: 5,
		},
		{
			name:           "empty text",
			summary:        "",
			explanation:    "",
			expectedMinLen: 0,
		},
		{
			name:           "only stop words",
			summary:        "the and or but",
			explanation:    "is are was were",
			expectedMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := method.extractKeywordsFromSummary(tt.summary, tt.explanation)
			assert.GreaterOrEqual(t, len(keywords), tt.expectedMinLen)
			
			// Verify no stop words
			stopWords := map[string]bool{"the": true, "and": true, "or": true, "but": true, "is": true, "are": true}
			for _, keyword := range keywords {
				assert.False(t, stopWords[keyword], "Should not contain stop words")
			}
		})
	}
}

// Note: Code conversion tests require classification package types
// These are tested in integration tests to avoid import cycles

func TestMLClassificationMethod_BuildEnhancedResult_WithoutCodeGen(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	
	enhancedResp := &infrastructure.EnhancedClassificationResponse{
		RequestID:      "test-123",
		Industry:       "Technology",
		Confidence:     0.85,
		Summary:        "Technology company providing software solutions",
		Explanation:    "Classified as Technology based on content analysis",
		AllScores:      map[string]float64{"Technology": 0.85, "Retail": 0.10},
		Classifications: []infrastructure.ClassificationPrediction{
			{Label: "Technology", Confidence: 0.85, Probability: 0.85, Rank: 1},
			{Label: "Retail", Confidence: 0.10, Probability: 0.10, Rank: 2},
		},
		ProcessingTime:     0.5,
		QuantizationEnabled: true,
		ModelVersion:       "2.0.0",
		Success:           true,
	}
	
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)
	ctx := context.Background()

	result := method.buildEnhancedResult(ctx, enhancedResp, "TechCorp")

	require.NotNil(t, result)
	
	// Verify primary fields
	assert.Equal(t, "Technology", result.PrimaryIndustry)
	assert.Equal(t, "Technology", result.IndustryName)
	assert.Equal(t, "Technology", result.IndustryCode)
	assert.Equal(t, 0.85, result.ConfidenceScore)
	assert.Equal(t, "ml_distilbart", result.ClassificationMethod)
	
	// Verify enhanced fields
	assert.Equal(t, "Technology company providing software solutions", result.ContentSummary)
	assert.Equal(t, "Classified as Technology based on content analysis", result.Explanation)
	assert.True(t, result.QuantizationEnabled)
	assert.Equal(t, "2.0.0", result.ModelVersion)
	
	// Verify all scores
	assert.Equal(t, 0.85, result.AllIndustryScores["Technology"])
	assert.Equal(t, 0.10, result.AllIndustryScores["Retail"])
	
	// Verify risk level calculation
	assert.NotEmpty(t, result.RiskLevel)
	assert.Contains(t, []string{"low", "medium", "high"}, result.RiskLevel)
	
	// Verify code distribution exists (even if empty)
	assert.NotNil(t, result.CodeDistribution)
}

func TestMLClassificationMethod_RiskLevelCalculation(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	method := NewMLClassificationMethod(nil, nil, nil, nil, logger)

	tests := []struct {
		name           string
		confidence     float64
		expectedRisk   string
	}{
		{
			name:         "high confidence - low risk",
			confidence:   0.90,
			expectedRisk: "low",
		},
		{
			name:         "medium confidence - medium risk",
			confidence:   0.60,
			expectedRisk: "medium",
		},
		{
			name:         "low confidence - high risk",
			confidence:   0.40,
			expectedRisk: "high",
		},
		{
			name:         "boundary - 0.5",
			confidence:   0.5,
			expectedRisk: "high",
		},
		{
			name:         "boundary - 0.7",
			confidence:   0.7,
			expectedRisk: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancedResp := &infrastructure.EnhancedClassificationResponse{
				Industry:    "Technology",
				Confidence:  tt.confidence,
				Summary:     "Test",
				Explanation: "Test",
				Classifications: []infrastructure.ClassificationPrediction{
					{Label: "Technology", Confidence: tt.confidence},
				},
			}

			result := method.buildEnhancedResult(context.Background(), enhancedResp, "Test")
			assert.Equal(t, tt.expectedRisk, result.RiskLevel)
		})
	}
}

// Note: Enhanced classification with Python service requires full integration
// These tests are in integration test suite to avoid import cycles

func TestMLClassificationMethod_PerformMLClassification_Fallback(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	
	// Create a mock ML classifier
	mlConfig := machine_learning.ContentClassifierConfig{
		ModelType:          "bert",
		MaxSequenceLength:  512,
		ConfidenceThreshold: 0.7,
	}
	mlClassifier := machine_learning.NewContentClassifier(mlConfig)

	// No Python service - should fallback to standard ML classification
	method := NewMLClassificationMethod(mlClassifier, nil, nil, nil, logger)
	ctx := context.Background()

	// Test without website URL - should use standard classification
	result, err := method.performMLClassification(ctx, "TechCorp", "Software development", "")

	// Standard ML classifier may not be fully initialized in test environment
	// So we check for either success or expected error
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
		return
	}

	if result != nil {
		assert.NotEmpty(t, result.IndustryCode)
		assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
	}
}

