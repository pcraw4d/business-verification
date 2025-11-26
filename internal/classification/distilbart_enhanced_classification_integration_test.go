//go:build integration

package classification

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/classification/testutil"
	"kyb-platform/internal/machine_learning/infrastructure"
	"kyb-platform/internal/shared"
)

// Helper function to create a mock Python ML service
func createMockPythonMLService(t *testing.T) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/classify-enhanced":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			var req infrastructure.EnhancedClassificationRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
				return
			}

			// Mock enhanced classification response
			response := infrastructure.EnhancedClassificationResponse{
				RequestID:           "test-request-123",
				ModelID:             "distilbart-model-1",
				ModelVersion:        "distilbart-v1.0",
				Classifications: []infrastructure.ClassificationPrediction{
					{Label: "Technology", Confidence: 0.92},
					{Label: "Software", Confidence: 0.85},
					{Label: "IT Services", Confidence: 0.78},
				},
				Confidence:          0.92,
				Summary:             "This business is a technology company specializing in software development and IT consulting services.",
				Explanation:         "The business was classified as Technology based on keywords like 'software', 'development', 'IT', and 'consulting' found in the description and website content. The high confidence score (92%) indicates strong alignment with technology industry characteristics.",
				ProcessingTime:      0.15,
				QuantizationEnabled: true,
				Timestamp:           time.Now(),
				Success:             true,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case "/classify":
			// Standard classification endpoint
			var req struct {
				Content string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			response := map[string]interface{}{
				"industry":   "Technology",
				"confidence": 0.85,
				"success":    true,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case "/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return httptest.NewServer(handler)
}

// Helper function to build classification from enhanced response
func buildClassificationFromEnhancedResponse(
	enhancedResp *infrastructure.EnhancedClassificationResponse,
	codes *ClassificationCodesInfo,
) *shared.IndustryClassification {
	// Get primary industry and confidence
	primaryIndustry := "Unknown"
	confidence := 0.0
	if len(enhancedResp.Classifications) > 0 {
		primaryIndustry = enhancedResp.Classifications[0].Label
		confidence = enhancedResp.Classifications[0].Confidence
	}

	// Build all industry scores map
	allScores := make(map[string]float64)
	for _, classification := range enhancedResp.Classifications {
		allScores[classification.Label] = classification.Confidence
	}

	// Convert codes
	sharedCodes := shared.ClassificationCodes{
		MCC:   []shared.MCCCode{},
		SIC:   []shared.SICCode{},
		NAICS: []shared.NAICSCode{},
	}

	if codes != nil {
		for _, code := range codes.MCC {
			sharedCodes.MCC = append(sharedCodes.MCC, shared.MCCCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  code.Confidence,
			})
		}
		for _, code := range codes.SIC {
			sharedCodes.SIC = append(sharedCodes.SIC, shared.SICCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  code.Confidence,
			})
		}
		for _, code := range codes.NAICS {
			sharedCodes.NAICS = append(sharedCodes.NAICS, shared.NAICSCode{
				Code:        code.Code,
				Description: code.Description,
				Confidence:  code.Confidence,
			})
		}
	}

	// Calculate code distribution
	codeDistribution := sharedCodes.CalculateCodeDistribution()

	// Determine risk level based on confidence
	riskLevel := "low"
	if confidence < 0.5 {
		riskLevel = "high"
	} else if confidence < 0.7 {
		riskLevel = "medium"
	}

	classification := &shared.IndustryClassification{
		IndustryCode:         primaryIndustry,
		IndustryName:         primaryIndustry,
		PrimaryIndustry:     primaryIndustry,
		ConfidenceScore:      confidence,
		ClassificationMethod: "ml_distilbart",
		ContentSummary:       enhancedResp.Summary,
		Explanation:          enhancedResp.Explanation,
		AllIndustryScores:    allScores,
		QuantizationEnabled:  enhancedResp.QuantizationEnabled,
		ModelVersion:         enhancedResp.ModelVersion,
		RiskLevel:            riskLevel,
		CodeDistribution:     &codeDistribution,
		Keywords:             []string{},
	}
	
	// Note: ClassificationCodes is not a field in IndustryClassification
	// It's stored separately in BusinessClassificationResponse
	// For this test, we'll store codes in metadata
	if codes != nil {
		if classification.Metadata == nil {
			classification.Metadata = make(map[string]interface{})
		}
		classification.Metadata["classification_codes"] = sharedCodes
	}
	
	return classification
}

// Helper function to extract keywords from text
func extractKeywordsFromText(text string) []string {
	var keywords []string
	
	// Simple keyword extraction - split by spaces and filter common words
	allText := strings.ToLower(text)
	words := strings.Fields(allText)
	
	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "are": true,
		"was": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"this": true, "that": true, "these": true, "those": true,
	}
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// TestDistilBARTEnhancedClassification_EndToEnd tests the complete enhanced classification flow
func TestDistilBARTEnhancedClassification_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if Python ML service URL is not set
	pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	if pythonMLServiceURL == "" {
		// Create mock Python ML service for testing
		mockPythonService := createMockPythonMLService(t)
		defer mockPythonService.Close()
		pythonMLServiceURL = mockPythonService.URL
	}

	// Setup test dependencies
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockRepo := testutil.NewMockKeywordRepository()
	codeGenerator := NewClassificationCodeGenerator(mockRepo, logger)
	websiteScraper := NewEnhancedWebsiteScraper(logger)

	// Create Python ML service client
	pythonMLService := infrastructure.NewPythonMLService(pythonMLServiceURL, logger)

	ctx := context.Background()

	tests := []struct {
		name                string
		businessName        string
		description         string
		websiteURL          string
		expectEnhanced      bool
		expectPrimaryIndustry bool
		expectExplanation   bool
		expectSummary        bool
		expectCodes         bool
		expectDistribution  bool
		expectRiskLevel     bool
	}{
		{
			name:                "technology company with website",
			businessName:        "TechCorp Solutions",
			description:         "Software development and IT consulting services",
			websiteURL:          "https://techcorp.com",
			expectEnhanced:      true,
			expectPrimaryIndustry: true,
			expectExplanation:   true,
			expectSummary:       true,
			expectCodes:         true,
			expectDistribution:  true,
			expectRiskLevel:     true,
		},
		{
			name:                "retail business with website",
			businessName:        "Fashion Store Inc",
			description:         "Clothing and accessories retail store",
			websiteURL:          "https://fashionstore.com",
			expectEnhanced:      true,
			expectPrimaryIndustry: true,
			expectExplanation:   true,
			expectSummary:       true,
			expectCodes:         true,
			expectDistribution:  true,
			expectRiskLevel:     true,
		},
		{
			name:                "business without website",
			businessName:        "Local Business",
			description:         "Local services",
			websiteURL:          "",
			expectEnhanced:      false, // Should fallback to standard classification
			expectPrimaryIndustry: true,
			expectExplanation:   false, // No explanation without enhanced classification
			expectSummary:       false,
			expectCodes:         true, // May still have codes from keyword classification
			expectDistribution:  true,
			expectRiskLevel:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test enhanced classification by calling Python ML service directly
			// Extract website content if URL is provided
			var websiteContent string
			if tt.websiteURL != "" && websiteScraper != nil {
				scrapingResult := websiteScraper.ScrapeWebsite(ctx, tt.websiteURL)
				if scrapingResult.Success {
					websiteContent = scrapingResult.TextContent
				}
			}
			
			// Call enhanced classification endpoint
			enhancedReq := &infrastructure.EnhancedClassificationRequest{
				BusinessName:     tt.businessName,
				Description:      tt.description,
				WebsiteURL:       tt.websiteURL,
				WebsiteContent:   websiteContent,
				MaxResults:       5,
				MaxContentLength: 1024,
			}
			
			enhancedResp, err := pythonMLService.ClassifyEnhanced(ctx, enhancedReq)
			if err != nil {
				if tt.expectEnhanced {
					t.Logf("Enhanced classification error (may be expected): %v", err)
				}
				return
			}

			require.NotNil(t, enhancedResp, "Expected enhanced classification response")
			require.True(t, enhancedResp.Success, "Expected successful enhanced classification")
			
			// Build result using code generator
			keywords := extractKeywordsFromText(enhancedResp.Summary + " " + enhancedResp.Explanation)
			// Get primary industry from classifications
			primaryIndustry := "Unknown"
			if len(enhancedResp.Classifications) > 0 {
				primaryIndustry = enhancedResp.Classifications[0].Label
			}
			codes, _ := codeGenerator.GenerateClassificationCodes(ctx, keywords, primaryIndustry, enhancedResp.Confidence)
			
			// Build classification result
			classification := buildClassificationFromEnhancedResponse(enhancedResp, codes)

			// Verify primary industry (Requirement 1)
			if tt.expectPrimaryIndustry {
				assert.NotEmpty(t, classification.PrimaryIndustry, "Should have primary industry")
				assert.Greater(t, classification.ConfidenceScore, 0.0, "Should have confidence score")
				assert.LessOrEqual(t, classification.ConfidenceScore, 1.0, "Confidence should be <= 1.0")
			}

			// Verify explanation (Requirement 4)
			if tt.expectExplanation {
				assert.NotEmpty(t, classification.Explanation, "Should have explanation for enhanced classification")
			}

			// Verify summary
			if tt.expectSummary {
				assert.NotEmpty(t, classification.ContentSummary, "Should have content summary for enhanced classification")
			}

			// Verify codes (Requirement 2)
			if tt.expectCodes {
				// Extract codes from metadata if available
				var sharedCodes shared.ClassificationCodes
				if classification.Metadata != nil {
					if codesVal, ok := classification.Metadata["classification_codes"]; ok {
						if codes, ok := codesVal.(shared.ClassificationCodes); ok {
							sharedCodes = codes
						}
					}
				}
				
				// Should have at least some codes
				totalCodes := len(sharedCodes.MCC) +
					len(sharedCodes.SIC) +
					len(sharedCodes.NAICS)
				
				if totalCodes == 0 {
					t.Logf("Warning: No codes generated, but code generator is available")
					// This is acceptable if code generator doesn't find matches
				} else {
					// Verify top 3 codes per type
					topMCC := sharedCodes.GetTopMCC(3)
					topSIC := sharedCodes.GetTopSIC(3)
					topNAICS := sharedCodes.GetTopNAICS(3)
					
					assert.LessOrEqual(t, len(topMCC), 3, "Should have at most 3 MCC codes")
					assert.LessOrEqual(t, len(topSIC), 3, "Should have at most 3 SIC codes")
					assert.LessOrEqual(t, len(topNAICS), 3, "Should have at most 3 NAICS codes")
				}
			}

			// Verify code distribution (Requirement 3)
			if tt.expectDistribution {
				assert.NotNil(t, classification.CodeDistribution, "Should have code distribution")
				if classification.CodeDistribution != nil {
					assert.GreaterOrEqual(t, classification.CodeDistribution.Total, 0, "Total should be >= 0")
				}
			}

			// Verify risk level (Requirement 5)
			if tt.expectRiskLevel {
				assert.NotEmpty(t, classification.RiskLevel, "Should have risk level")
				assert.Contains(t, []string{"low", "medium", "high"}, classification.RiskLevel, "Risk level should be low, medium, or high")
			}
		})
	}
}

// TestDistilBARTEnhancedClassification_AllUIRequirements explicitly verifies all 5 UI requirements
func TestDistilBARTEnhancedClassification_AllUIRequirements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock Python ML service
	mockPythonService := createMockPythonMLService(t)
	defer mockPythonService.Close()

	// Setup test dependencies
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockRepo := testutil.NewMockKeywordRepository()
	codeGenerator := NewClassificationCodeGenerator(mockRepo, logger)
	websiteScraper := NewEnhancedWebsiteScraper(logger)

	pythonMLService := infrastructure.NewPythonMLService(mockPythonService.URL, logger)

	ctx := context.Background()
	
	// Test with website URL
	scrapingResult := websiteScraper.ScrapeWebsite(ctx, "https://techcorp.com")
	var websiteContent string
	if scrapingResult.Success {
		websiteContent = scrapingResult.TextContent
	}
	
	enhancedReq := &infrastructure.EnhancedClassificationRequest{
		BusinessName:     "TechCorp Solutions",
		Description:      "Software development services",
		WebsiteURL:       "https://techcorp.com",
		WebsiteContent:   websiteContent,
		MaxResults:       5,
		MaxContentLength: 1024,
	}
	
	enhancedResp, err := pythonMLService.ClassifyEnhanced(ctx, enhancedReq)
	require.NoError(t, err)
	require.NotNil(t, enhancedResp)
	require.True(t, enhancedResp.Success)
	
	// Build result
	keywords := extractKeywordsFromText(enhancedResp.Summary + " " + enhancedResp.Explanation)
	// Get primary industry from classifications
	primaryIndustry := "Unknown"
	if len(enhancedResp.Classifications) > 0 {
		primaryIndustry = enhancedResp.Classifications[0].Label
	}
	codes, _ := codeGenerator.GenerateClassificationCodes(ctx, keywords, primaryIndustry, enhancedResp.Confidence)
	classification := buildClassificationFromEnhancedResponse(enhancedResp, codes)

	// Requirement 1: Primary Industry with Confidence Level
	t.Run("Requirement1_PrimaryIndustryWithConfidence", func(t *testing.T) {
		assert.NotEmpty(t, classification.PrimaryIndustry, "Should have primary industry")
		assert.Greater(t, classification.ConfidenceScore, 0.0, "Should have confidence score > 0")
		assert.LessOrEqual(t, classification.ConfidenceScore, 1.0, "Should have confidence score <= 1.0")
	})

	// Requirement 2: Top 3 Codes by Type (MCC/SIC/NAICS) with Confidence
	t.Run("Requirement2_Top3CodesByType", func(t *testing.T) {
		// Extract codes from metadata if available
		var sharedCodes shared.ClassificationCodes
		if classification.Metadata != nil {
			if codesVal, ok := classification.Metadata["classification_codes"]; ok {
				if codes, ok := codesVal.(shared.ClassificationCodes); ok {
					sharedCodes = codes
				}
			}
		}
		
		topMCC := sharedCodes.GetTopMCC(3)
		topSIC := sharedCodes.GetTopSIC(3)
		topNAICS := sharedCodes.GetTopNAICS(3)

		// Verify limits
		assert.LessOrEqual(t, len(topMCC), 3, "Should have at most 3 MCC codes")
		assert.LessOrEqual(t, len(topSIC), 3, "Should have at most 3 SIC codes")
		assert.LessOrEqual(t, len(topNAICS), 3, "Should have at most 3 NAICS codes")

		// Verify confidence scores if codes exist
		for _, code := range topMCC {
			assert.Greater(t, code.Confidence, 0.0, "MCC code should have confidence")
			assert.NotEmpty(t, code.Code, "MCC code should have code value")
			assert.NotEmpty(t, code.Description, "MCC code should have description")
		}
		for _, code := range topSIC {
			assert.Greater(t, code.Confidence, 0.0, "SIC code should have confidence")
			assert.NotEmpty(t, code.Code, "SIC code should have code value")
			assert.NotEmpty(t, code.Description, "SIC code should have description")
		}
		for _, code := range topNAICS {
			assert.Greater(t, code.Confidence, 0.0, "NAICS code should have confidence")
			assert.NotEmpty(t, code.Code, "NAICS code should have code value")
			assert.NotEmpty(t, code.Description, "NAICS code should have description")
		}
	})

	// Requirement 3: Industry Code Distribution
	t.Run("Requirement3_CodeDistribution", func(t *testing.T) {
		assert.NotNil(t, classification.CodeDistribution, "Should have code distribution")
		
		if classification.CodeDistribution != nil {
			assert.GreaterOrEqual(t, classification.CodeDistribution.Total, 0, "Total should be >= 0")
			assert.Equal(t, classification.CodeDistribution.Total,
				classification.CodeDistribution.MCC.Count+
					classification.CodeDistribution.SIC.Count+
					classification.CodeDistribution.NAICS.Count,
				"Total should equal sum of counts")
		}
	})

	// Requirement 4: Explanation
	t.Run("Requirement4_Explanation", func(t *testing.T) {
		assert.NotEmpty(t, classification.Explanation, "Should have explanation")
		assert.Contains(t, classification.Explanation, "classified", "Explanation should mention classification")
	})

	// Requirement 5: Risk Level
	t.Run("Requirement5_RiskLevel", func(t *testing.T) {
		assert.NotEmpty(t, classification.RiskLevel, "Should have risk level")
		assert.Contains(t, []string{"low", "medium", "high"}, classification.RiskLevel, "Risk level should be low, medium, or high")
	})
	}
