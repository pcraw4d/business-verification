package mocks

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/modules/risk_assessment"
)

// MockRiskAssessmentService provides a mock implementation of risk assessment service for E2E tests
type MockRiskAssessmentService struct {
	// Configuration for mock behavior
	ShouldFail   bool
	Delay        time.Duration
	MockResults  *risk_assessment.RiskAssessmentResult
	ErrorMessage string
}

// NewMockRiskAssessmentService creates a new mock risk assessment service
func NewMockRiskAssessmentService() *MockRiskAssessmentService {
	return &MockRiskAssessmentService{
		ShouldFail:   false,
		Delay:        200 * time.Millisecond,
		MockResults:  getDefaultMockRiskResults(),
		ErrorMessage: "mock risk assessment error",
	}
}

// AssessRisk implements the risk assessment service interface for E2E tests
func (m *MockRiskAssessmentService) AssessRisk(ctx context.Context, request *risk_assessment.RiskAssessmentRequest) (*risk_assessment.RiskAssessmentResult, error) {
	// Simulate processing delay
	if m.Delay > 0 {
		time.Sleep(m.Delay)
	}

	// Simulate failure if configured
	if m.ShouldFail {
		return nil, &RiskAssessmentError{
			Message: m.ErrorMessage,
			Code:    "MOCK_RISK_ASSESSMENT_ERROR",
		}
	}

	// Create mock response based on request
	result := &risk_assessment.RiskAssessmentResult{
		AssessmentID:       fmt.Sprintf("risk_assessment_%d", time.Now().Unix()),
		BusinessID:         request.BusinessID,
		BusinessName:       request.BusinessName,
		WebsiteURL:         request.WebsiteURL,
		OverallRiskScore:   0.35, // Low risk
		RiskLevel:          "LOW",
		AssessmentDate:     time.Now(),
		ProcessingTime:     m.Delay,
		AssessmentMethods:  request.AssessmentTypes,
		SecurityAnalysis:   getMockSecurityAnalysis(),
		DomainAnalysis:     getMockDomainAnalysis(),
		ReputationAnalysis: getMockReputationAnalysis(),
		ComplianceAnalysis: getMockComplianceAnalysis(),
		FinancialAnalysis:  getMockFinancialAnalysis(),
		Recommendations:    getMockRecommendations(),
		Metadata: map[string]interface{}{
			"source":    "mock_service",
			"version":   "1.0.0",
			"test_mode": true,
		},
	}

	return result, nil
}

// GetRiskAssessmentHistory returns mock risk assessment history
func (m *MockRiskAssessmentService) GetRiskAssessmentHistory(ctx context.Context, businessID string) ([]*risk_assessment.RiskAssessmentResult, error) {
	// Simulate processing delay
	if m.Delay > 0 {
		time.Sleep(m.Delay / 2)
	}

	// Simulate failure if configured
	if m.ShouldFail {
		return nil, &RiskAssessmentError{
			Message: m.ErrorMessage,
			Code:    "MOCK_RISK_HISTORY_ERROR",
		}
	}

	// Return mock history
	history := []*risk_assessment.RiskAssessmentResult{
		{
			AssessmentID:     "risk_assessment_1",
			BusinessID:       businessID,
			OverallRiskScore: 0.25,
			RiskLevel:        "LOW",
			AssessmentDate:   time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			AssessmentID:     "risk_assessment_2",
			BusinessID:       businessID,
			OverallRiskScore: 0.30,
			RiskLevel:        "LOW",
			AssessmentDate:   time.Now().Add(-14 * 24 * time.Hour),
		},
	}

	return history, nil
}

// SetMockResults allows configuring mock results for testing
func (m *MockRiskAssessmentService) SetMockResults(results *risk_assessment.RiskAssessmentResult) {
	m.MockResults = results
}

// SetFailureMode configures the mock to fail with a specific error
func (m *MockRiskAssessmentService) SetFailureMode(shouldFail bool, errorMessage string) {
	m.ShouldFail = shouldFail
	m.ErrorMessage = errorMessage
}

// SetDelay configures the processing delay for the mock
func (m *MockRiskAssessmentService) SetDelay(delay time.Duration) {
	m.Delay = delay
}

// getDefaultMockRiskResults returns default mock risk assessment results
func getDefaultMockRiskResults() *risk_assessment.RiskAssessmentResult {
	return &risk_assessment.RiskAssessmentResult{
		AssessmentID:       "mock_risk_assessment_1",
		BusinessID:         "mock_business_1",
		BusinessName:       "Mock Business",
		WebsiteURL:         "https://mockbusiness.com",
		OverallRiskScore:   0.35,
		RiskLevel:          "LOW",
		AssessmentDate:     time.Now(),
		ProcessingTime:     200 * time.Millisecond,
		AssessmentMethods:  []string{"security_analysis", "domain_analysis", "reputation_analysis"},
		SecurityAnalysis:   getMockSecurityAnalysis(),
		DomainAnalysis:     getMockDomainAnalysis(),
		ReputationAnalysis: getMockReputationAnalysis(),
		ComplianceAnalysis: getMockComplianceAnalysis(),
		FinancialAnalysis:  getMockFinancialAnalysis(),
		Recommendations:    getMockRecommendations(),
		Metadata: map[string]interface{}{
			"source":    "mock_service",
			"version":   "1.0.0",
			"test_mode": true,
		},
	}
}

// getMockSecurityAnalysis returns mock security analysis results
func getMockSecurityAnalysis() *risk_assessment.SecurityAnalysis {
	return &risk_assessment.SecurityAnalysis{
		SSLScore: 0.85,
		TLSScore: 0.90,
		SecurityHeaders: []risk_assessment.SecurityHeader{
			{Name: "HSTS", Present: true, Value: "max-age=31536000; includeSubDomains"},
			{Name: "CSP", Present: true, Value: "default-src 'self'"},
			{Name: "X-Frame-Options", Present: true, Value: "DENY"},
		},
		Vulnerabilities: []risk_assessment.Vulnerability{
			{Type: "SSL", Severity: "LOW", Description: "Minor SSL configuration issue"},
		},
		OverallSecurityScore: 0.80,
		Recommendations: []string{
			"Update SSL certificate",
			"Implement additional security headers",
		},
	}
}

// getMockDomainAnalysis returns mock domain analysis results
func getMockDomainAnalysis() *risk_assessment.DomainAnalysis {
	return &risk_assessment.DomainAnalysis{
		DomainAge:        365 * 2, // 2 years
		Registrar:        "Mock Registrar Inc",
		RegistrationDate: time.Now().Add(-2 * 365 * 24 * time.Hour),
		ExpirationDate:   time.Now().Add(365 * 24 * time.Hour),
		DNSSEC:           true,
		DNSRecords: []risk_assessment.DNSRecord{
			{Type: "A", Value: "192.168.1.1", TTL: 3600},
			{Type: "MX", Value: "mail.mockbusiness.com", TTL: 3600},
		},
		OverallDomainScore: 0.75,
		Recommendations: []string{
			"Enable DNSSEC",
			"Update DNS records",
		},
	}
}

// getMockReputationAnalysis returns mock reputation analysis results
func getMockReputationAnalysis() *risk_assessment.ReputationAnalysis {
	return &risk_assessment.ReputationAnalysis{
		OverallScore: 0.70,
		SocialMediaPresence: []risk_assessment.SocialMediaPresence{
			{Platform: "Twitter", Followers: 1000, Engagement: 0.05},
			{Platform: "LinkedIn", Followers: 500, Engagement: 0.03},
		},
		OnlineReviews: []risk_assessment.OnlineReview{
			{Platform: "Google", Rating: 4.2, ReviewCount: 25},
			{Platform: "Yelp", Rating: 3.8, ReviewCount: 15},
		},
		BrandMentions: []risk_assessment.BrandMention{
			{Source: "News Article", Sentiment: "POSITIVE", Date: time.Now().Add(-7 * 24 * time.Hour)},
		},
		Recommendations: []string{
			"Improve social media engagement",
			"Address negative reviews",
		},
	}
}

// getMockComplianceAnalysis returns mock compliance analysis results
func getMockComplianceAnalysis() *risk_assessment.ComplianceAnalysis {
	return &risk_assessment.ComplianceAnalysis{
		OverallComplianceScore: 0.85,
		ComplianceChecks: []risk_assessment.ComplianceCheck{
			{Type: "GDPR", Status: "COMPLIANT", Score: 0.90},
			{Type: "CCPA", Status: "COMPLIANT", Score: 0.85},
			{Type: "PCI-DSS", Status: "PARTIAL", Score: 0.70},
		},
		Certifications: []risk_assessment.Certification{
			{Name: "ISO 27001", Status: "VALID", ExpirationDate: time.Now().Add(365 * 24 * time.Hour)},
		},
		Recommendations: []string{
			"Complete PCI-DSS compliance",
			"Renew expiring certifications",
		},
	}
}

// getMockFinancialAnalysis returns mock financial analysis results
func getMockFinancialAnalysis() *risk_assessment.FinancialAnalysis {
	return &risk_assessment.FinancialAnalysis{
		OverallFinancialScore: 0.75,
		RevenueIndicators: []risk_assessment.RevenueIndicator{
			{Type: "EMPLOYEE_COUNT", Value: 50, Confidence: 0.90},
			{Type: "WEBSITE_TRAFFIC", Value: 10000, Confidence: 0.70},
		},
		StabilityMetrics: []risk_assessment.StabilityMetric{
			{Type: "DOMAIN_AGE", Value: 730, Confidence: 0.95},
			{Type: "SOCIAL_PRESENCE", Value: 0.70, Confidence: 0.80},
		},
		Recommendations: []string{
			"Improve financial transparency",
			"Provide additional financial documentation",
		},
	}
}

// getMockRecommendations returns mock risk recommendations
func getMockRecommendations() []risk_assessment.RiskRecommendation {
	return []risk_assessment.RiskRecommendation{
		{
			Category:    "SECURITY",
			Priority:    "MEDIUM",
			Description: "Update SSL certificate configuration",
			Action:      "Contact your hosting provider to update SSL settings",
			Impact:      "Improves security score by 10%",
		},
		{
			Category:    "REPUTATION",
			Priority:    "LOW",
			Description: "Improve social media engagement",
			Action:      "Increase posting frequency and respond to comments",
			Impact:      "Improves reputation score by 5%",
		},
		{
			Category:    "COMPLIANCE",
			Priority:    "HIGH",
			Description: "Complete PCI-DSS compliance",
			Action:      "Implement required security controls for payment processing",
			Impact:      "Improves compliance score by 15%",
		},
	}
}

// RiskAssessmentError represents a risk assessment-specific error
type RiskAssessmentError struct {
	Message string
	Code    string
}

func (e *RiskAssessmentError) Error() string {
	return e.Message
}
