package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/modules/risk_assessment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testComprehensiveRiskAssessment tests comprehensive risk assessment
func (suite *FeatureFunctionalityTestSuite) testComprehensiveRiskAssessment(t *testing.T) {
	tests := []struct {
		name              string
		businessName      string
		websiteURL        string
		domainName        string
		industry          string
		expectedResult    bool
		expectedRiskLevel string
	}{
		{
			name:              "Low Risk Technology Company",
			businessName:      "Microsoft Corporation",
			websiteURL:        "https://microsoft.com",
			domainName:        "microsoft.com",
			industry:          "Technology",
			expectedResult:    true,
			expectedRiskLevel: "low",
		},
		{
			name:              "Medium Risk E-commerce Business",
			businessName:      "Online Marketplace",
			websiteURL:        "https://onlinemarketplace.com",
			domainName:        "onlinemarketplace.com",
			industry:          "E-commerce",
			expectedResult:    true,
			expectedRiskLevel: "medium",
		},
		{
			name:              "High Risk Cryptocurrency Business",
			businessName:      "Crypto Exchange",
			websiteURL:        "https://cryptoexchange.com",
			domainName:        "cryptoexchange.com",
			industry:          "Cryptocurrency",
			expectedResult:    true,
			expectedRiskLevel: "high",
		},
		{
			name:              "Critical Risk Adult Entertainment",
			businessName:      "Adult Entertainment Site",
			websiteURL:        "https://adultsite.com",
			domainName:        "adultsite.com",
			industry:          "Adult Entertainment",
			expectedResult:    true,
			expectedRiskLevel: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Create risk assessment request
			req := &risk_assessment.RiskAssessmentRequest{
				BusinessName: tt.businessName,
				WebsiteURL:   tt.websiteURL,
				DomainName:   tt.domainName,
				Industry:     tt.industry,
				RequestID:    generateTestRequestID(),
			}

			// Perform risk assessment
			result, err := suite.riskAssessmentService.AssessRisk(ctx, req)

			if tt.expectedResult {
				require.NoError(t, err, "Risk assessment should succeed")
				require.NotNil(t, result, "Result should not be nil")

				// Verify result structure
				assert.NotEmpty(t, result.RequestID, "Request ID should be present")
				assert.Equal(t, tt.businessName, result.BusinessName, "Business name should match")
				assert.Equal(t, tt.WebsiteURL, result.WebsiteURL, "Website URL should match")
				assert.NotZero(t, result.AssessmentTimestamp, "Assessment timestamp should be present")
				assert.True(t, result.ProcessingTime > 0, "Processing time should be positive")

				// Verify risk level
				assert.Equal(t, tt.expectedRiskLevel, string(result.RiskLevel), "Risk level should match expected")
				assert.True(t, result.OverallRiskScore >= 0 && result.OverallRiskScore <= 1, "Risk score should be between 0 and 1")
				assert.True(t, result.ConfidenceScore >= 0 && result.ConfidenceScore <= 1, "Confidence score should be between 0 and 1")

				// Verify risk factors
				assert.NotNil(t, result.RiskFactors, "Risk factors should be present")
				assert.NotNil(t, result.Recommendations, "Recommendations should be present")

			} else {
				assert.Error(t, err, "Risk assessment should fail")
			}
		})
	}
}

// testSecurityAnalysis tests security analysis functionality
func (suite *FeatureFunctionalityTestSuite) testSecurityAnalysis(t *testing.T) {
	tests := []struct {
		name                  string
		websiteURL            string
		expectedResult        bool
		expectedSecurityScore float64
	}{
		{
			name:                  "Secure Website Analysis",
			websiteURL:            "https://microsoft.com",
			expectedResult:        true,
			expectedSecurityScore: 0.8,
		},
		{
			name:                  "Insecure Website Analysis",
			websiteURL:            "http://insecure-site.com",
			expectedResult:        true,
			expectedSecurityScore: 0.3,
		},
		{
			name:                  "Unknown Website Analysis",
			websiteURL:            "https://unknown-site.com",
			expectedResult:        true,
			expectedSecurityScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create security analysis request
			req := &risk_assessment.RiskAssessmentRequest{
				BusinessName: "Test Business",
				WebsiteURL:   tt.websiteURL,
				DomainName:   extractDomainFromURL(tt.websiteURL),
			}

			// Perform security analysis
			securityResult, err := suite.riskAssessmentService.securityAnalyzer.AnalyzeSecurity(ctx, req)

			if tt.expectedResult {
				require.NoError(t, err, "Security analysis should succeed")
				require.NotNil(t, securityResult, "Security result should not be nil")

				// Verify security analysis structure
				assert.NotZero(t, securityResult.AnalysisTimestamp, "Analysis timestamp should be present")
				assert.True(t, securityResult.ProcessingTime > 0, "Processing time should be positive")
				assert.True(t, securityResult.SecurityScore >= 0 && securityResult.SecurityScore <= 1, "Security score should be between 0 and 1")

				// Verify security checks
				assert.NotNil(t, securityResult.SSLStatus, "SSL status should be present")
				assert.NotNil(t, securityResult.SecurityHeaders, "Security headers should be present")
				assert.NotNil(t, securityResult.Vulnerabilities, "Vulnerabilities should be present")

				// Verify expected security score
				assert.True(t, securityResult.SecurityScore >= tt.expectedSecurityScore,
					"Security score should meet expected threshold: got %.2f, expected >= %.2f",
					securityResult.SecurityScore, tt.expectedSecurityScore)

			} else {
				assert.Error(t, err, "Security analysis should fail")
			}
		})
	}
}

// testDomainAnalysis tests domain analysis functionality
func (suite *FeatureFunctionalityTestSuite) testDomainAnalysis(t *testing.T) {
	tests := []struct {
		name                string
		domainName          string
		expectedResult      bool
		expectedDomainScore float64
	}{
		{
			name:                "Established Domain Analysis",
			domainName:          "microsoft.com",
			expectedResult:      true,
			expectedDomainScore: 0.9,
		},
		{
			name:                "New Domain Analysis",
			domainName:          "newdomain123.com",
			expectedResult:      true,
			expectedDomainScore: 0.4,
		},
		{
			name:                "Suspicious Domain Analysis",
			domainName:          "suspicious-site.net",
			expectedResult:      true,
			expectedDomainScore: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create domain analysis request
			req := &risk_assessment.RiskAssessmentRequest{
				BusinessName: "Test Business",
				DomainName:   tt.domainName,
			}

			// Perform domain analysis
			domainResult, err := suite.riskAssessmentService.domainAnalyzer.AnalyzeDomain(ctx, req)

			if tt.expectedResult {
				require.NoError(t, err, "Domain analysis should succeed")
				require.NotNil(t, domainResult, "Domain result should not be nil")

				// Verify domain analysis structure
				assert.NotZero(t, domainResult.AnalysisTimestamp, "Analysis timestamp should be present")
				assert.True(t, domainResult.ProcessingTime > 0, "Processing time should be positive")
				assert.True(t, domainResult.DomainScore >= 0 && domainResult.DomainScore <= 1, "Domain score should be between 0 and 1")

				// Verify domain information
				assert.NotNil(t, domainResult.WHOISData, "WHOIS data should be present")
				assert.NotNil(t, domainResult.DNSRecords, "DNS records should be present")
				assert.NotNil(t, domainResult.DomainAge, "Domain age should be present")

				// Verify expected domain score
				assert.True(t, domainResult.DomainScore >= tt.expectedDomainScore,
					"Domain score should meet expected threshold: got %.2f, expected >= %.2f",
					domainResult.DomainScore, tt.expectedDomainScore)

			} else {
				assert.Error(t, err, "Domain analysis should fail")
			}
		})
	}
}

// testReputationAnalysis tests reputation analysis functionality
func (suite *FeatureFunctionalityTestSuite) testReputationAnalysis(t *testing.T) {
	tests := []struct {
		name                    string
		businessName            string
		websiteURL              string
		expectedResult          bool
		expectedReputationScore float64
	}{
		{
			name:                    "High Reputation Business Analysis",
			businessName:            "Microsoft Corporation",
			websiteURL:              "https://microsoft.com",
			expectedResult:          true,
			expectedReputationScore: 0.9,
		},
		{
			name:                    "Medium Reputation Business Analysis",
			businessName:            "Local Business Inc",
			websiteURL:              "https://localbusiness.com",
			expectedResult:          true,
			expectedReputationScore: 0.6,
		},
		{
			name:                    "Low Reputation Business Analysis",
			businessName:            "Controversial Company",
			websiteURL:              "https://controversial.com",
			expectedResult:          true,
			expectedReputationScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			// Create reputation analysis request
			req := &risk_assessment.RiskAssessmentRequest{
				BusinessName: tt.businessName,
				WebsiteURL:   tt.websiteURL,
			}

			// Perform reputation analysis
			reputationResult, err := suite.riskAssessmentService.reputationAnalyzer.AnalyzeReputation(ctx, req)

			if tt.expectedResult {
				require.NoError(t, err, "Reputation analysis should succeed")
				require.NotNil(t, reputationResult, "Reputation result should not be nil")

				// Verify reputation analysis structure
				assert.NotZero(t, reputationResult.AnalysisTimestamp, "Analysis timestamp should be present")
				assert.True(t, reputationResult.ProcessingTime > 0, "Processing time should be positive")
				assert.True(t, reputationResult.ReputationScore >= 0 && reputationResult.ReputationScore <= 1, "Reputation score should be between 0 and 1")

				// Verify reputation data
				assert.NotNil(t, reputationResult.ReviewAnalysis, "Review analysis should be present")
				assert.NotNil(t, reputationResult.SocialMediaAnalysis, "Social media analysis should be present")
				assert.NotNil(t, reputationResult.NewsAnalysis, "News analysis should be present")

				// Verify expected reputation score
				assert.True(t, reputationResult.ReputationScore >= tt.expectedReputationScore,
					"Reputation score should meet expected threshold: got %.2f, expected >= %.2f",
					reputationResult.ReputationScore, tt.expectedReputationScore)

			} else {
				assert.Error(t, err, "Reputation analysis should fail")
			}
		})
	}
}

// testComplianceAnalysis tests compliance analysis functionality
func (suite *FeatureFunctionalityTestSuite) testComplianceAnalysis(t *testing.T) {
	tests := []struct {
		name                    string
		businessName            string
		industry                string
		expectedResult          bool
		expectedComplianceScore float64
	}{
		{
			name:                    "Financial Services Compliance Analysis",
			businessName:            "Bank of America",
			industry:                "Financial Services",
			expectedResult:          true,
			expectedComplianceScore: 0.9,
		},
		{
			name:                    "Healthcare Compliance Analysis",
			businessName:            "Mayo Clinic",
			industry:                "Healthcare",
			expectedResult:          true,
			expectedComplianceScore: 0.9,
		},
		{
			name:                    "Technology Compliance Analysis",
			businessName:            "Tech Startup",
			industry:                "Technology",
			expectedResult:          true,
			expectedComplianceScore: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create compliance analysis request
			req := &risk_assessment.RiskAssessmentRequest{
				BusinessName: tt.businessName,
				Industry:     tt.industry,
			}

			// Perform compliance analysis
			complianceResult, err := suite.riskAssessmentService.complianceAnalyzer.AnalyzeCompliance(ctx, req)

			if tt.expectedResult {
				require.NoError(t, err, "Compliance analysis should succeed")
				require.NotNil(t, complianceResult, "Compliance result should not be nil")

				// Verify compliance analysis structure
				assert.NotZero(t, complianceResult.AnalysisTimestamp, "Analysis timestamp should be present")
				assert.True(t, complianceResult.ProcessingTime > 0, "Processing time should be positive")
				assert.True(t, complianceResult.ComplianceScore >= 0 && complianceResult.ComplianceScore <= 1, "Compliance score should be between 0 and 1")

				// Verify compliance data
				assert.NotNil(t, complianceResult.RegulatoryCompliance, "Regulatory compliance should be present")
				assert.NotNil(t, complianceResult.IndustryStandards, "Industry standards should be present")
				assert.NotNil(t, complianceResult.Certifications, "Certifications should be present")

				// Verify expected compliance score
				assert.True(t, complianceResult.ComplianceScore >= tt.expectedComplianceScore,
					"Compliance score should meet expected threshold: got %.2f, expected >= %.2f",
					complianceResult.ComplianceScore, tt.expectedComplianceScore)

			} else {
				assert.Error(t, err, "Compliance analysis should fail")
			}
		})
	}
}

// testFinancialAnalysis tests financial analysis functionality
func (suite *FeatureFunctionalityTestSuite) testFinancialAnalysis(t *testing.T) {
	tests := []struct {
		name                   string
		businessName           string
		industry               string
		expectedResult         bool
		expectedFinancialScore float64
	}{
		{
			name:                   "Established Company Financial Analysis",
			businessName:           "Microsoft Corporation",
			industry:               "Technology",
			expectedResult:         true,
			expectedFinancialScore: 0.9,
		},
		{
			name:                   "Startup Financial Analysis",
			businessName:           "Tech Startup",
			industry:               "Technology",
			expectedResult:         true,
			expectedFinancialScore: 0.5,
		},
		{
			name:                   "Small Business Financial Analysis",
			businessName:           "Local Business",
			industry:               "Retail",
			expectedResult:         true,
			expectedFinancialScore: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Perform financial analysis
			financialResult, err := suite.riskAssessmentService.financialAnalyzer.AnalyzeFinancial(ctx, tt.businessName, tt.industry)

			if tt.expectedResult {
				require.NoError(t, err, "Financial analysis should succeed")
				require.NotNil(t, financialResult, "Financial result should not be nil")

				// Verify financial analysis structure
				assert.NotZero(t, financialResult.AnalysisTimestamp, "Analysis timestamp should be present")
				assert.True(t, financialResult.ProcessingTime > 0, "Processing time should be positive")
				assert.True(t, financialResult.FinancialScore >= 0 && financialResult.FinancialScore <= 1, "Financial score should be between 0 and 1")

				// Verify financial data
				assert.NotNil(t, financialResult.FinancialMetrics, "Financial metrics should be present")
				assert.NotNil(t, financialResult.CreditAnalysis, "Credit analysis should be present")
				assert.NotNil(t, financialResult.RiskIndicators, "Risk indicators should be present")

				// Verify expected financial score
				assert.True(t, financialResult.FinancialScore >= tt.expectedFinancialScore,
					"Financial score should meet expected threshold: got %.2f, expected >= %.2f",
					financialResult.FinancialScore, tt.expectedFinancialScore)

			} else {
				assert.Error(t, err, "Financial analysis should fail")
			}
		})
	}
}

// testRiskScoring tests risk scoring functionality
func (suite *FeatureFunctionalityTestSuite) testRiskScoring(t *testing.T) {
	tests := []struct {
		name              string
		riskFactors       []risk_assessment.RiskFactor
		expectedResult    bool
		expectedRiskLevel string
	}{
		{
			name: "Low Risk Factors",
			riskFactors: []risk_assessment.RiskFactor{
				{Category: "security", Score: 0.9, Weight: 0.3},
				{Category: "reputation", Score: 0.8, Weight: 0.2},
				{Category: "compliance", Score: 0.9, Weight: 0.3},
				{Category: "financial", Score: 0.8, Weight: 0.2},
			},
			expectedResult:    true,
			expectedRiskLevel: "low",
		},
		{
			name: "High Risk Factors",
			riskFactors: []risk_assessment.RiskFactor{
				{Category: "security", Score: 0.3, Weight: 0.3},
				{Category: "reputation", Score: 0.2, Weight: 0.2},
				{Category: "compliance", Score: 0.4, Weight: 0.3},
				{Category: "financial", Score: 0.3, Weight: 0.2},
			},
			expectedResult:    true,
			expectedRiskLevel: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Create risk assessment result with test factors
			result := &risk_assessment.RiskAssessmentResult{
				BusinessName: "Test Business",
				RiskFactors:  tt.riskFactors,
			}

			// Perform risk scoring
			riskScore, err := suite.riskAssessmentService.riskScorer.CalculateRiskScore(ctx, result)

			if tt.expectedResult {
				require.NoError(t, err, "Risk scoring should succeed")
				require.NotNil(t, riskScore, "Risk score should not be nil")

				// Verify risk score structure
				assert.NotZero(t, riskScore.ScoreTimestamp, "Score timestamp should be present")
				assert.True(t, riskScore.OverallScore >= 0 && riskScore.OverallScore <= 1, "Overall score should be between 0 and 1")
				assert.Equal(t, tt.expectedRiskLevel, string(riskScore.RiskLevel), "Risk level should match expected")
				assert.True(t, riskScore.ConfidenceLevel >= 0 && riskScore.ConfidenceLevel <= 1, "Confidence level should be between 0 and 1")

				// Verify category scores
				assert.NotNil(t, riskScore.CategoryScores, "Category scores should be present")
				assert.NotNil(t, riskScore.WeightedFactors, "Weighted factors should be present")
				assert.NotNil(t, riskScore.ScoreBreakdown, "Score breakdown should be present")

			} else {
				assert.Error(t, err, "Risk scoring should fail")
			}
		})
	}
}

// Helper functions
func generateTestRequestID() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}

func extractDomainFromURL(url string) string {
	// Simple domain extraction for testing
	// In real implementation, this would use proper URL parsing
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}

	// Find first slash and remove path
	if slashIndex := strings.Index(url, "/"); slashIndex != -1 {
		url = url[:slashIndex]
	}

	return url
}
