package test

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMultiMethodClassification tests the multi-method classification system
func (suite *FeatureFunctionalityTestSuite) testMultiMethodClassification(t *testing.T) {
	tests := []struct {
		name            string
		businessName    string
		description     string
		websiteURL      string
		expectedResult  bool
		expectedMethods int
	}{
		{
			name:            "Technology Company Classification",
			businessName:    "TechCorp Solutions",
			description:     "Software development and IT consulting services",
			websiteURL:      "https://techcorp.com",
			expectedResult:  true,
			expectedMethods: 3, // keyword, ML, similarity
		},
		{
			name:            "Retail Business Classification",
			businessName:    "Fashion Store",
			description:     "Clothing and accessories retail store",
			websiteURL:      "https://fashionstore.com",
			expectedResult:  true,
			expectedMethods: 3,
		},
		{
			name:            "Financial Services Classification",
			businessName:    "Investment Bank",
			description:     "Investment banking and financial advisory services",
			websiteURL:      "https://investmentbank.com",
			expectedResult:  true,
			expectedMethods: 3,
		},
		{
			name:            "Healthcare Services Classification",
			businessName:    "Medical Clinic",
			description:     "Primary healthcare and medical services",
			websiteURL:      "https://medicalclinic.com",
			expectedResult:  true,
			expectedMethods: 3,
		},
		{
			name:            "Manufacturing Company Classification",
			businessName:    "Industrial Manufacturing",
			description:     "Heavy machinery and industrial equipment manufacturing",
			websiteURL:      "https://industrialmfg.com",
			expectedResult:  true,
			expectedMethods: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create multi-method classifier
			mmc := suite.createMultiMethodClassifier()

			// Perform classification
			result, err := mmc.ClassifyWithMultipleMethods(ctx, tt.businessName, tt.description, tt.websiteURL)

			if tt.expectedResult {
				require.NoError(t, err, "Classification should succeed")
				require.NotNil(t, result, "Result should not be nil")

				// Verify result structure
				assert.NotEmpty(t, result.RequestID, "Request ID should be present")
				assert.NotEmpty(t, result.BusinessName, "Business name should be present")
				assert.NotEmpty(t, result.ProcessingTime, "Processing time should be present")

				// Verify method results
				assert.Len(t, result.MethodResults, tt.expectedMethods, "Should have expected number of methods")

				// Verify each method has results
				for _, methodResult := range result.MethodResults {
					assert.NotEmpty(t, methodResult.MethodName, "Method name should be present")
					assert.NotEmpty(t, methodResult.MethodType, "Method type should be present")
					assert.True(t, methodResult.ProcessingTime > 0, "Processing time should be positive")

					if methodResult.Success {
						assert.NotNil(t, methodResult.Result, "Result should not be nil for successful methods")
						assert.True(t, methodResult.Confidence >= 0 && methodResult.Confidence <= 1, "Confidence should be between 0 and 1")
					}
				}

				// Verify ensemble result
				if result.EnsembleResult != nil {
					assert.NotEmpty(t, result.EnsembleResult.IndustryName, "Ensemble industry name should be present")
					assert.True(t, result.EnsembleResult.ConfidenceScore >= 0 && result.EnsembleResult.ConfidenceScore <= 1, "Ensemble confidence should be between 0 and 1")
					assert.NotEmpty(t, result.EnsembleResult.ClassificationCodes, "Classification codes should be present")
				}

				// Verify quality metrics
				assert.NotNil(t, result.QualityMetrics, "Quality metrics should be present")
				assert.True(t, result.QualityMetrics.OverallConfidence >= 0 && result.QualityMetrics.OverallConfidence <= 1, "Overall confidence should be between 0 and 1")
				assert.True(t, result.QualityMetrics.MethodAgreement >= 0 && result.QualityMetrics.MethodAgreement <= 1, "Method agreement should be between 0 and 1")

			} else {
				assert.Error(t, err, "Classification should fail")
			}
		})
	}
}

// testKeywordBasedClassification tests keyword-based classification
func (suite *FeatureFunctionalityTestSuite) testKeywordBasedClassification(t *testing.T) {
	tests := []struct {
		name               string
		keywords           []string
		expectedIndustry   string
		expectedConfidence float64
	}{
		{
			name:               "Technology Keywords",
			keywords:           []string{"software", "development", "programming", "IT", "technology"},
			expectedIndustry:   "Technology",
			expectedConfidence: 0.8,
		},
		{
			name:               "Retail Keywords",
			keywords:           []string{"retail", "store", "clothing", "fashion", "shopping"},
			expectedIndustry:   "Retail",
			expectedConfidence: 0.8,
		},
		{
			name:               "Financial Keywords",
			keywords:           []string{"banking", "finance", "investment", "financial", "money"},
			expectedIndustry:   "Financial Services",
			expectedConfidence: 0.8,
		},
		{
			name:               "Healthcare Keywords",
			keywords:           []string{"medical", "healthcare", "clinic", "hospital", "health"},
			expectedIndustry:   "Healthcare",
			expectedConfidence: 0.8,
		},
		{
			name:               "Manufacturing Keywords",
			keywords:           []string{"manufacturing", "production", "factory", "industrial", "machinery"},
			expectedIndustry:   "Manufacturing",
			expectedConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Create keyword classifier
			classifier := suite.createKeywordClassifier()

			// Perform classification
			result, err := classifier.ClassifyBusinessByKeywords(ctx, tt.keywords)

			require.NoError(t, err, "Keyword classification should succeed")
			require.NotNil(t, result, "Result should not be nil")

			// Verify result
			assert.NotNil(t, result.Industry, "Industry should not be nil")
			assert.Contains(t, result.Industry.Name, tt.expectedIndustry, "Industry should match expected")
			assert.True(t, result.Confidence >= tt.expectedConfidence, "Confidence should meet expected threshold")
			assert.NotEmpty(t, result.Keywords, "Keywords should be present")
			assert.NotEmpty(t, result.Reasoning, "Reasoning should be present")
		})
	}
}

// testMLBasedClassification tests ML-based classification
func (suite *FeatureFunctionalityTestSuite) testMLBasedClassification(t *testing.T) {
	tests := []struct {
		name               string
		businessName       string
		description        string
		expectedResult     bool
		expectedConfidence float64
	}{
		{
			name:               "Technology Company ML Classification",
			businessName:       "AI Solutions Inc",
			description:        "Artificial intelligence and machine learning solutions for enterprise clients",
			expectedResult:     true,
			expectedConfidence: 0.85,
		},
		{
			name:               "E-commerce Business ML Classification",
			businessName:       "Online Marketplace",
			description:        "Digital marketplace connecting buyers and sellers worldwide",
			expectedResult:     true,
			expectedConfidence: 0.85,
		},
		{
			name:               "Consulting Services ML Classification",
			businessName:       "Business Consulting Group",
			description:        "Strategic business consulting and advisory services",
			expectedResult:     true,
			expectedConfidence: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Create ML classifier
			mlClassifier := suite.createMLClassifier()

			// Perform ML classification
			result, err := mlClassifier.ClassifyContent(ctx, tt.description, tt.businessName)

			if tt.expectedResult {
				require.NoError(t, err, "ML classification should succeed")
				require.NotNil(t, result, "Result should not be nil")

				// Verify result structure
				assert.NotEmpty(t, result.Classifications, "Classifications should be present")
				assert.True(t, len(result.Classifications) > 0, "Should have at least one classification")

				// Verify best classification
				bestClassification := result.Classifications[0]
				assert.NotEmpty(t, bestClassification.Label, "Label should be present")
				assert.True(t, bestClassification.Confidence >= tt.expectedConfidence, "Confidence should meet expected threshold")
				assert.True(t, bestClassification.Confidence >= 0 && bestClassification.Confidence <= 1, "Confidence should be between 0 and 1")

			} else {
				assert.Error(t, err, "ML classification should fail")
			}
		})
	}
}

// testEnsembleClassification tests ensemble classification
func (suite *FeatureFunctionalityTestSuite) testEnsembleClassification(t *testing.T) {
	tests := []struct {
		name               string
		businessName       string
		description        string
		websiteURL         string
		expectedResult     bool
		expectedConfidence float64
	}{
		{
			name:               "Complex Business Ensemble Classification",
			businessName:       "FinTech Innovations",
			description:        "Financial technology solutions combining AI, blockchain, and traditional banking",
			websiteURL:         "https://fintechinnovations.com",
			expectedResult:     true,
			expectedConfidence: 0.9,
		},
		{
			name:               "Multi-Industry Business Ensemble Classification",
			businessName:       "Global Solutions Corp",
			description:        "Diversified business providing technology, consulting, and manufacturing services",
			websiteURL:         "https://globalsolutions.com",
			expectedResult:     true,
			expectedConfidence: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create ensemble classifier
			ensembleClassifier := suite.createEnsembleClassifier()

			// Perform ensemble classification
			result, err := ensembleClassifier.ClassifyWithEnsemble(ctx, tt.businessName, tt.description, tt.websiteURL)

			if tt.expectedResult {
				require.NoError(t, err, "Ensemble classification should succeed")
				require.NotNil(t, result, "Result should not be nil")

				// Verify result structure
				assert.NotEmpty(t, result.MethodName, "Method name should be present")
				assert.Equal(t, "ensemble", result.MethodType, "Method type should be ensemble")
				assert.True(t, result.Success, "Ensemble classification should succeed")
				assert.True(t, result.Confidence >= tt.expectedConfidence, "Confidence should meet expected threshold")

				// Verify result content
				require.NotNil(t, result.Result, "Result should not be nil")
				assert.NotEmpty(t, result.Result.IndustryName, "Industry name should be present")
				assert.NotEmpty(t, result.Result.ClassificationCodes, "Classification codes should be present")

			} else {
				assert.Error(t, err, "Ensemble classification should fail")
			}
		})
	}
}

// testConfidenceScoring tests confidence scoring system
func (suite *FeatureFunctionalityTestSuite) testConfidenceScoring(t *testing.T) {
	tests := []struct {
		name                  string
		businessName          string
		description           string
		expectedMinConfidence float64
		expectedMaxConfidence float64
	}{
		{
			name:                  "High Confidence Business",
			businessName:          "Microsoft Corporation",
			description:           "Technology company specializing in software development and cloud computing",
			expectedMinConfidence: 0.9,
			expectedMaxConfidence: 1.0,
		},
		{
			name:                  "Medium Confidence Business",
			businessName:          "Local Services LLC",
			description:           "Various business services and consulting",
			expectedMinConfidence: 0.6,
			expectedMaxConfidence: 0.9,
		},
		{
			name:                  "Low Confidence Business",
			businessName:          "ABC Company",
			description:           "Business operations",
			expectedMinConfidence: 0.3,
			expectedMaxConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// Create confidence scorer
			confidenceScorer := suite.createConfidenceScorer()

			// Perform confidence scoring
			confidence, err := confidenceScorer.CalculateConfidence(ctx, tt.businessName, tt.description)

			require.NoError(t, err, "Confidence scoring should succeed")
			assert.True(t, confidence >= tt.expectedMinConfidence && confidence <= tt.expectedMaxConfidence,
				"Confidence should be within expected range: got %.2f, expected %.2f-%.2f",
				confidence, tt.expectedMinConfidence, tt.expectedMaxConfidence)
		})
	}
}

// Helper methods for creating mock classifiers
func (suite *FeatureFunctionalityTestSuite) createMultiMethodClassifier() *classification.MultiMethodClassifier {
	// Implementation would create a real multi-method classifier with mock dependencies
	// For testing purposes, this would use the actual implementation with test data
	return nil // Placeholder - would be implemented with actual classifier
}

func (suite *FeatureFunctionalityTestSuite) createKeywordClassifier() *classification.SupabaseKeywordRepository {
	// Implementation would create a real keyword classifier with test data
	return nil // Placeholder - would be implemented with actual classifier
}

func (suite *FeatureFunctionalityTestSuite) createMLClassifier() *classification.MLIntegrationManager {
	// Implementation would create a real ML classifier with mock models
	return nil // Placeholder - would be implemented with actual classifier
}

func (suite *FeatureFunctionalityTestSuite) createEnsembleClassifier() *classification.MLIntegrationManager {
	// Implementation would create a real ensemble classifier
	return nil // Placeholder - would be implemented with actual classifier
}

func (suite *FeatureFunctionalityTestSuite) createConfidenceScorer() *classification.ConfidenceScorer {
	// Implementation would create a real confidence scorer
	return nil // Placeholder - would be implemented with actual scorer
}
