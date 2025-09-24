package risk_assessment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ComplianceAnalyzer provides regulatory compliance analysis capabilities
type ComplianceAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
}

// ComplianceAnalysisResult contains comprehensive compliance analysis results
type ComplianceAnalysisResult struct {
	BusinessName           string                  `json:"business_name"`
	IndustryCompliance     *IndustryCompliance     `json:"industry_compliance,omitempty"`
	PrivacyPolicyAnalysis  *PrivacyPolicyAnalysis  `json:"privacy_policy_analysis,omitempty"`
	TermsOfServiceAnalysis *TermsOfServiceAnalysis `json:"terms_of_service_analysis,omitempty"`
	CertificationsAnalysis *CertificationsAnalysis `json:"certifications_analysis,omitempty"`
	OverallScore           float64                 `json:"overall_score"`
	RiskFactors            []RiskFactor            `json:"risk_factors"`
	Recommendations        []string                `json:"recommendations"`
	LastUpdated            time.Time               `json:"last_updated"`
}

// IndustryCompliance contains industry-specific compliance checks
type IndustryCompliance struct {
	IndustryType           string                      `json:"industry_type"`
	RegulatoryFrameworks   []RegulatoryFramework       `json:"regulatory_frameworks"`
	ComplianceStatus       map[string]ComplianceStatus `json:"compliance_status"`
	RequiredCertifications []string                    `json:"required_certifications"`
	ComplianceScore        float64                     `json:"compliance_score"`
	RiskLevel              string                      `json:"risk_level"`
}

// RegulatoryFramework represents a regulatory framework
type RegulatoryFramework struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Applicable     bool       `json:"applicable"`
	Requirements   []string   `json:"requirements"`
	ComplianceDate *time.Time `json:"compliance_date,omitempty"`
	Status         string     `json:"status"`
}

// ComplianceStatus represents compliance status for a framework
type ComplianceStatus struct {
	Status        string    `json:"status"`
	LastChecked   time.Time `json:"last_checked"`
	NextReview    time.Time `json:"next_review"`
	Issues        []string  `json:"issues"`
	Score         float64   `json:"score"`
	Documentation string    `json:"documentation"`
}

// PrivacyPolicyAnalysis contains privacy policy analysis results
type PrivacyPolicyAnalysis struct {
	PolicyURL           string                 `json:"policy_url"`
	LastUpdated         time.Time              `json:"last_updated"`
	PolicyLength        int                    `json:"policy_length"`
	ComplianceChecks    map[string]PolicyCheck `json:"compliance_checks"`
	GDPRCompliance      bool                   `json:"gdpr_compliance"`
	CCPACompliance      bool                   `json:"ccpa_compliance"`
	DataRetentionPolicy string                 `json:"data_retention_policy"`
	ThirdPartySharing   bool                   `json:"third_party_sharing"`
	UserRights          []string               `json:"user_rights"`
	OverallScore        float64                `json:"overall_score"`
	Issues              []string               `json:"issues"`
}

// PolicyCheck represents a specific policy compliance check
type PolicyCheck struct {
	Requirement    string `json:"requirement"`
	Compliant      bool   `json:"compliant"`
	Evidence       string `json:"evidence"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
}

// TermsOfServiceAnalysis contains terms of service analysis results
type TermsOfServiceAnalysis struct {
	TermsURL          string                 `json:"terms_url"`
	LastUpdated       time.Time              `json:"last_updated"`
	TermsLength       int                    `json:"terms_length"`
	ComplianceChecks  map[string]PolicyCheck `json:"compliance_checks"`
	LiabilityLimits   string                 `json:"liability_limits"`
	DisputeResolution string                 `json:"dispute_resolution"`
	Jurisdiction      string                 `json:"jurisdiction"`
	OverallScore      float64                `json:"overall_score"`
	Issues            []string               `json:"issues"`
}

// CertificationsAnalysis contains certification analysis results
type CertificationsAnalysis struct {
	Certifications     []Certification `json:"certifications"`
	RequiredCerts      []string        `json:"required_certs"`
	MissingCerts       []string        `json:"missing_certs"`
	ExpiringCerts      []Certification `json:"expiring_certs"`
	OverallScore       float64         `json:"overall_score"`
	CertificationScore float64         `json:"certification_score"`
}

// Certification represents a business certification
type Certification struct {
	Name            string     `json:"name"`
	IssuingBody     string     `json:"issuing_body"`
	IssueDate       time.Time  `json:"issue_date"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	Status          string     `json:"status"`
	CertificationID string     `json:"certification_id"`
	Scope           string     `json:"scope"`
	Verified        bool       `json:"verified"`
}

// NewComplianceAnalyzer creates a new compliance analyzer
func NewComplianceAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *ComplianceAnalyzer {
	return &ComplianceAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeCompliance performs comprehensive compliance analysis
func (ca *ComplianceAnalyzer) AnalyzeCompliance(ctx context.Context, businessName string, industry string, websiteURL string) (*ComplianceAnalysisResult, error) {
	ca.logger.Info("Starting compliance analysis",
		zap.String("business", businessName),
		zap.String("industry", industry))

	result := &ComplianceAnalysisResult{
		BusinessName: businessName,
		LastUpdated:  time.Now(),
	}

	// Analyze industry-specific compliance if enabled
	if ca.config.ComplianceCheckEnabled {
		industryCompliance, err := ca.analyzeIndustryCompliance(ctx, businessName, industry)
		if err != nil {
			ca.logger.Warn("Industry compliance analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "compliance",
				Factor:      "industry_compliance_analysis",
				Description: fmt.Sprintf("Industry compliance analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Unable to assess compliance status",
			})
		} else {
			result.IndustryCompliance = industryCompliance
		}
	}

	// Analyze privacy policy if enabled
	if ca.config.ComplianceCheckEnabled {
		privacyAnalysis, err := ca.analyzePrivacyPolicy(ctx, businessName, websiteURL)
		if err != nil {
			ca.logger.Warn("Privacy policy analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "compliance",
				Factor:      "privacy_policy_analysis",
				Description: fmt.Sprintf("Privacy policy analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify privacy policy compliance",
			})
		} else {
			result.PrivacyPolicyAnalysis = privacyAnalysis
		}
	}

	// Analyze terms of service if enabled
	if ca.config.ComplianceCheckEnabled {
		termsAnalysis, err := ca.analyzeTermsOfService(ctx, businessName, websiteURL)
		if err != nil {
			ca.logger.Warn("Terms of service analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "compliance",
				Factor:      "terms_of_service_analysis",
				Description: fmt.Sprintf("Terms of service analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify terms of service compliance",
			})
		} else {
			result.TermsOfServiceAnalysis = termsAnalysis
		}
	}

	// Analyze certifications if enabled
	if ca.config.ComplianceCheckEnabled {
		certificationsAnalysis, err := ca.analyzeCertifications(ctx, businessName, industry)
		if err != nil {
			ca.logger.Warn("Certifications analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "compliance",
				Factor:      "certifications_analysis",
				Description: fmt.Sprintf("Certifications analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify certification status",
			})
		} else {
			result.CertificationsAnalysis = certificationsAnalysis
		}
	}

	// Calculate overall score
	result.OverallScore = ca.calculateOverallScore(result)

	// Generate recommendations
	result.Recommendations = ca.generateRecommendations(result)

	ca.logger.Info("Compliance analysis completed",
		zap.String("business", businessName),
		zap.Float64("score", result.OverallScore))

	return result, nil
}

// analyzeIndustryCompliance analyzes industry-specific compliance requirements
func (ca *ComplianceAnalyzer) analyzeIndustryCompliance(ctx context.Context, businessName string, industry string) (*IndustryCompliance, error) {
	ca.logger.Debug("Analyzing industry compliance",
		zap.String("business", businessName),
		zap.String("industry", industry))

	// In a real implementation, this would query compliance databases
	// For now, we'll simulate industry-specific compliance data
	industryCompliance := &IndustryCompliance{
		IndustryType:     industry,
		ComplianceStatus: make(map[string]ComplianceStatus),
	}

	// Define regulatory frameworks based on industry
	frameworks := ca.getRegulatoryFrameworks(industry)
	industryCompliance.RegulatoryFrameworks = frameworks

	// Simulate compliance status for each framework
	for _, framework := range frameworks {
		if framework.Applicable {
			status := ca.simulateComplianceStatus(framework)
			industryCompliance.ComplianceStatus[framework.Name] = status
		}
	}

	// Calculate compliance score
	industryCompliance.ComplianceScore = ca.calculateIndustryComplianceScore(industryCompliance)
	industryCompliance.RiskLevel = ca.determineRiskLevel(industryCompliance.ComplianceScore)

	// Set required certifications
	industryCompliance.RequiredCertifications = ca.getRequiredCertifications(industry)

	return industryCompliance, nil
}

// analyzePrivacyPolicy analyzes privacy policy compliance
func (ca *ComplianceAnalyzer) analyzePrivacyPolicy(ctx context.Context, businessName string, websiteURL string) (*PrivacyPolicyAnalysis, error) {
	ca.logger.Debug("Analyzing privacy policy",
		zap.String("business", businessName))

	// In a real implementation, this would fetch and parse the privacy policy
	// For now, we'll simulate privacy policy analysis
	privacyAnalysis := &PrivacyPolicyAnalysis{
		PolicyURL:        fmt.Sprintf("%s/privacy-policy", websiteURL),
		LastUpdated:      time.Now().AddDate(0, -2, -15), // 2.5 months ago
		PolicyLength:     2500,                           // words
		ComplianceChecks: make(map[string]PolicyCheck),
	}

	// Simulate GDPR compliance checks
	privacyAnalysis.ComplianceChecks["gdpr_data_collection"] = PolicyCheck{
		Requirement:    "Clear data collection purpose",
		Compliant:      true,
		Evidence:       "Policy clearly states data collection purposes",
		Severity:       "high",
		Recommendation: "Continue maintaining clear data collection notices",
	}

	privacyAnalysis.ComplianceChecks["gdpr_user_rights"] = PolicyCheck{
		Requirement:    "User rights (access, rectification, deletion)",
		Compliant:      true,
		Evidence:       "Policy includes user rights section",
		Severity:       "high",
		Recommendation: "Ensure user rights are easily exercisable",
	}

	privacyAnalysis.ComplianceChecks["gdpr_data_retention"] = PolicyCheck{
		Requirement:    "Data retention policy",
		Compliant:      true,
		Evidence:       "Policy specifies data retention periods",
		Severity:       "medium",
		Recommendation: "Regularly review retention periods",
	}

	privacyAnalysis.ComplianceChecks["gdpr_third_party"] = PolicyCheck{
		Requirement:    "Third-party data sharing disclosure",
		Compliant:      false,
		Evidence:       "Third-party sharing not clearly disclosed",
		Severity:       "high",
		Recommendation: "Add clear third-party sharing disclosure",
	}

	// Set compliance flags
	privacyAnalysis.GDPRCompliance = ca.checkGDPRCompliance(privacyAnalysis.ComplianceChecks)
	privacyAnalysis.CCPACompliance = ca.checkCCPACompliance(privacyAnalysis.ComplianceChecks)
	privacyAnalysis.DataRetentionPolicy = "Data retained for 2 years unless deletion requested"
	privacyAnalysis.ThirdPartySharing = true
	privacyAnalysis.UserRights = []string{"Access", "Rectification", "Deletion", "Portability"}

	// Calculate overall score
	privacyAnalysis.OverallScore = ca.calculatePrivacyPolicyScore(privacyAnalysis)

	// Identify issues
	privacyAnalysis.Issues = ca.identifyPrivacyPolicyIssues(privacyAnalysis)

	return privacyAnalysis, nil
}

// analyzeTermsOfService analyzes terms of service compliance
func (ca *ComplianceAnalyzer) analyzeTermsOfService(ctx context.Context, businessName string, websiteURL string) (*TermsOfServiceAnalysis, error) {
	ca.logger.Debug("Analyzing terms of service",
		zap.String("business", businessName))

	// In a real implementation, this would fetch and parse the terms of service
	// For now, we'll simulate terms of service analysis
	termsAnalysis := &TermsOfServiceAnalysis{
		TermsURL:         fmt.Sprintf("%s/terms-of-service", websiteURL),
		LastUpdated:      time.Now().AddDate(0, -1, -10), // 1.3 months ago
		TermsLength:      1800,                           // words
		ComplianceChecks: make(map[string]PolicyCheck),
	}

	// Simulate terms of service compliance checks
	termsAnalysis.ComplianceChecks["liability_limits"] = PolicyCheck{
		Requirement:    "Clear liability limitations",
		Compliant:      true,
		Evidence:       "Terms include liability limitation clauses",
		Severity:       "medium",
		Recommendation: "Ensure liability limits are reasonable",
	}

	termsAnalysis.ComplianceChecks["dispute_resolution"] = PolicyCheck{
		Requirement:    "Dispute resolution process",
		Compliant:      true,
		Evidence:       "Terms specify dispute resolution procedures",
		Severity:       "medium",
		Recommendation: "Consider adding mediation options",
	}

	termsAnalysis.ComplianceChecks["jurisdiction"] = PolicyCheck{
		Requirement:    "Clear jurisdiction specification",
		Compliant:      false,
		Evidence:       "Jurisdiction not clearly specified",
		Severity:       "high",
		Recommendation: "Add clear jurisdiction clause",
	}

	// Set additional fields
	termsAnalysis.LiabilityLimits = "Limited to service fees paid"
	termsAnalysis.DisputeResolution = "Arbitration required before litigation"
	termsAnalysis.Jurisdiction = "Not specified"

	// Calculate overall score
	termsAnalysis.OverallScore = ca.calculateTermsOfServiceScore(termsAnalysis)

	// Identify issues
	termsAnalysis.Issues = ca.identifyTermsOfServiceIssues(termsAnalysis)

	return termsAnalysis, nil
}

// analyzeCertifications analyzes business certifications
func (ca *ComplianceAnalyzer) analyzeCertifications(ctx context.Context, businessName string, industry string) (*CertificationsAnalysis, error) {
	ca.logger.Debug("Analyzing certifications",
		zap.String("business", businessName),
		zap.String("industry", industry))

	// In a real implementation, this would query certification databases
	// For now, we'll simulate certification data
	certificationsAnalysis := &CertificationsAnalysis{
		Certifications: make([]Certification, 0),
		RequiredCerts:  ca.getRequiredCertifications(industry),
		MissingCerts:   make([]string, 0),
		ExpiringCerts:  make([]Certification, 0),
	}

	// Simulate existing certifications
	iso9001 := Certification{
		Name:            "ISO 9001:2015",
		IssuingBody:     "ISO",
		IssueDate:       time.Now().AddDate(-1, -6, 0), // 1.5 years ago
		ExpiryDate:      &time.Time{},                  // Will be set below
		Status:          "active",
		CertificationID: "ISO-9001-2023-001",
		Scope:           "Quality Management System",
		Verified:        true,
	}
	expiry := time.Now().AddDate(0, 6, 0) // Expires in 6 months
	iso9001.ExpiryDate = &expiry
	certificationsAnalysis.Certifications = append(certificationsAnalysis.Certifications, iso9001)

	// Add more certifications based on industry
	if strings.Contains(strings.ToLower(industry), "financial") {
		soxCert := Certification{
			Name:            "SOX Compliance",
			IssuingBody:     "Internal Audit",
			IssueDate:       time.Now().AddDate(0, -3, 0), // 3 months ago
			Status:          "active",
			CertificationID: "SOX-2024-001",
			Scope:           "Financial Reporting Controls",
			Verified:        true,
		}
		certificationsAnalysis.Certifications = append(certificationsAnalysis.Certifications, soxCert)
	}

	// Identify missing certifications
	requiredCerts := ca.getRequiredCertifications(industry)
	existingCerts := make(map[string]bool)
	for _, cert := range certificationsAnalysis.Certifications {
		existingCerts[cert.Name] = true
	}

	for _, required := range requiredCerts {
		if !existingCerts[required] {
			certificationsAnalysis.MissingCerts = append(certificationsAnalysis.MissingCerts, required)
		}
	}

	// Identify expiring certifications
	for _, cert := range certificationsAnalysis.Certifications {
		if cert.ExpiryDate != nil && time.Until(*cert.ExpiryDate) < 6*30*24*time.Hour { // 6 months
			certificationsAnalysis.ExpiringCerts = append(certificationsAnalysis.ExpiringCerts, cert)
		}
	}

	// Calculate scores
	certificationsAnalysis.OverallScore = ca.calculateCertificationsScore(certificationsAnalysis)
	certificationsAnalysis.CertificationScore = ca.calculateCertificationCoverageScore(certificationsAnalysis)

	return certificationsAnalysis, nil
}

// Helper methods for compliance analysis

func (ca *ComplianceAnalyzer) getRegulatoryFrameworks(industry string) []RegulatoryFramework {
	industry = strings.ToLower(industry)
	var frameworks []RegulatoryFramework

	// Common frameworks for most industries
	frameworks = append(frameworks, RegulatoryFramework{
		Name:         "GDPR",
		Description:  "General Data Protection Regulation",
		Applicable:   true,
		Requirements: []string{"Data protection", "User consent", "Data portability"},
		Status:       "compliant",
	})

	frameworks = append(frameworks, RegulatoryFramework{
		Name:         "CCPA",
		Description:  "California Consumer Privacy Act",
		Applicable:   true,
		Requirements: []string{"Privacy notices", "Consumer rights", "Data disclosure"},
		Status:       "compliant",
	})

	// Industry-specific frameworks
	if strings.Contains(industry, "financial") {
		frameworks = append(frameworks, RegulatoryFramework{
			Name:         "SOX",
			Description:  "Sarbanes-Oxley Act",
			Applicable:   true,
			Requirements: []string{"Financial reporting", "Internal controls", "Audit requirements"},
			Status:       "compliant",
		})
	}

	if strings.Contains(industry, "healthcare") {
		frameworks = append(frameworks, RegulatoryFramework{
			Name:         "HIPAA",
			Description:  "Health Insurance Portability and Accountability Act",
			Applicable:   true,
			Requirements: []string{"Patient privacy", "Data security", "Breach notification"},
			Status:       "compliant",
		})
	}

	return frameworks
}

func (ca *ComplianceAnalyzer) simulateComplianceStatus(framework RegulatoryFramework) ComplianceStatus {
	// Simulate compliance status with some variation
	status := ComplianceStatus{
		LastChecked: time.Now().AddDate(0, -1, -15), // 1.5 months ago
		NextReview:  time.Now().AddDate(0, 1, 0),    // 1 month from now
		Score:       0.85,                           // 85% compliance
	}

	// Add some issues for realism
	if framework.Name == "GDPR" {
		status.Issues = []string{"Third-party data sharing needs clarification"}
		status.Score = 0.78
	}

	status.Status = "compliant"
	if status.Score < 0.8 {
		status.Status = "needs_improvement"
	}

	return status
}

func (ca *ComplianceAnalyzer) getRequiredCertifications(industry string) []string {
	industry = strings.ToLower(industry)
	var certs []string

	// Common certifications
	certs = append(certs, "ISO 9001")

	// Industry-specific certifications
	if strings.Contains(industry, "financial") {
		certs = append(certs, "SOX Compliance", "PCI DSS")
	}

	if strings.Contains(industry, "healthcare") {
		certs = append(certs, "HIPAA Compliance", "HITECH")
	}

	if strings.Contains(industry, "technology") {
		certs = append(certs, "ISO 27001", "SOC 2")
	}

	return certs
}

func (ca *ComplianceAnalyzer) checkGDPRCompliance(checks map[string]PolicyCheck) bool {
	requiredChecks := []string{"gdpr_data_collection", "gdpr_user_rights", "gdpr_data_retention"}

	for _, check := range requiredChecks {
		if policyCheck, exists := checks[check]; !exists || !policyCheck.Compliant {
			return false
		}
	}

	return true
}

func (ca *ComplianceAnalyzer) checkCCPACompliance(checks map[string]PolicyCheck) bool {
	// Simplified CCPA compliance check
	// In a real implementation, this would check for CCPA-specific requirements
	return true
}

// Scoring methods

func (ca *ComplianceAnalyzer) calculateIndustryComplianceScore(compliance *IndustryCompliance) float64 {
	if len(compliance.ComplianceStatus) == 0 {
		return 0.0
	}

	totalScore := 0.0
	count := 0

	for _, status := range compliance.ComplianceStatus {
		totalScore += status.Score
		count++
	}

	return totalScore / float64(count)
}

func (ca *ComplianceAnalyzer) calculatePrivacyPolicyScore(analysis *PrivacyPolicyAnalysis) float64 {
	if len(analysis.ComplianceChecks) == 0 {
		return 0.0
	}

	totalScore := 0.0
	count := 0

	for _, check := range analysis.ComplianceChecks {
		if check.Compliant {
			totalScore += 1.0
		}
		count++
	}

	baseScore := totalScore / float64(count)

	// Bonus for recent updates
	daysSinceUpdate := time.Since(analysis.LastUpdated).Hours() / 24
	if daysSinceUpdate < 90 { // Less than 3 months
		baseScore += 0.1
	}

	return ca.max(0.0, ca.min(1.0, baseScore))
}

func (ca *ComplianceAnalyzer) calculateTermsOfServiceScore(analysis *TermsOfServiceAnalysis) float64 {
	if len(analysis.ComplianceChecks) == 0 {
		return 0.0
	}

	totalScore := 0.0
	count := 0

	for _, check := range analysis.ComplianceChecks {
		if check.Compliant {
			totalScore += 1.0
		}
		count++
	}

	baseScore := totalScore / float64(count)

	// Bonus for recent updates
	daysSinceUpdate := time.Since(analysis.LastUpdated).Hours() / 24
	if daysSinceUpdate < 90 { // Less than 3 months
		baseScore += 0.1
	}

	return ca.max(0.0, ca.min(1.0, baseScore))
}

func (ca *ComplianceAnalyzer) calculateCertificationsScore(analysis *CertificationsAnalysis) float64 {
	if len(analysis.RequiredCerts) == 0 {
		return 1.0 // No requirements means perfect score
	}

	existingCerts := make(map[string]bool)
	for _, cert := range analysis.Certifications {
		existingCerts[cert.Name] = true
	}

	compliantCount := 0
	for _, required := range analysis.RequiredCerts {
		if existingCerts[required] {
			compliantCount++
		}
	}

	return float64(compliantCount) / float64(len(analysis.RequiredCerts))
}

func (ca *ComplianceAnalyzer) calculateCertificationCoverageScore(analysis *CertificationsAnalysis) float64 {
	if len(analysis.Certifications) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, cert := range analysis.Certifications {
		if cert.Status == "active" && cert.Verified {
			totalScore += 1.0
		} else if cert.Status == "active" {
			totalScore += 0.8
		}
	}

	return totalScore / float64(len(analysis.Certifications))
}

func (ca *ComplianceAnalyzer) calculateOverallScore(result *ComplianceAnalysisResult) float64 {
	score := 0.5 // Base score

	// Industry compliance score (30% weight)
	if result.IndustryCompliance != nil {
		score += result.IndustryCompliance.ComplianceScore * 0.30
	}

	// Privacy policy score (25% weight)
	if result.PrivacyPolicyAnalysis != nil {
		score += result.PrivacyPolicyAnalysis.OverallScore * 0.25
	}

	// Terms of service score (20% weight)
	if result.TermsOfServiceAnalysis != nil {
		score += result.TermsOfServiceAnalysis.OverallScore * 0.20
	}

	// Certifications score (25% weight)
	if result.CertificationsAnalysis != nil {
		score += result.CertificationsAnalysis.OverallScore * 0.25
	}

	return ca.max(0.0, ca.min(1.0, score))
}

// Issue identification methods

func (ca *ComplianceAnalyzer) identifyPrivacyPolicyIssues(analysis *PrivacyPolicyAnalysis) []string {
	var issues []string

	for _, check := range analysis.ComplianceChecks {
		if !check.Compliant {
			issues = append(issues, check.Recommendation)
		}
	}

	if time.Since(analysis.LastUpdated) > 365*24*time.Hour { // More than 1 year
		issues = append(issues, "Privacy policy should be updated annually")
	}

	return issues
}

func (ca *ComplianceAnalyzer) identifyTermsOfServiceIssues(analysis *TermsOfServiceAnalysis) []string {
	var issues []string

	for _, check := range analysis.ComplianceChecks {
		if !check.Compliant {
			issues = append(issues, check.Recommendation)
		}
	}

	if time.Since(analysis.LastUpdated) > 365*24*time.Hour { // More than 1 year
		issues = append(issues, "Terms of service should be updated annually")
	}

	return issues
}

// Utility methods

func (ca *ComplianceAnalyzer) determineRiskLevel(score float64) string {
	if score >= 0.9 {
		return "low"
	} else if score >= 0.7 {
		return "medium"
	} else if score >= 0.5 {
		return "high"
	}
	return "critical"
}

func (ca *ComplianceAnalyzer) generateRecommendations(result *ComplianceAnalysisResult) []string {
	var recommendations []string

	// Industry compliance recommendations
	if result.IndustryCompliance != nil {
		if result.IndustryCompliance.ComplianceScore < 0.8 {
			recommendations = append(recommendations, "Improve industry compliance to meet regulatory requirements")
		}
		for framework, status := range result.IndustryCompliance.ComplianceStatus {
			if status.Status == "needs_improvement" {
				recommendations = append(recommendations, fmt.Sprintf("Address compliance issues with %s framework", framework))
			}
		}
	}

	// Privacy policy recommendations
	if result.PrivacyPolicyAnalysis != nil {
		if !result.PrivacyPolicyAnalysis.GDPRCompliance {
			recommendations = append(recommendations, "Ensure GDPR compliance in privacy policy")
		}
		if !result.PrivacyPolicyAnalysis.CCPACompliance {
			recommendations = append(recommendations, "Ensure CCPA compliance in privacy policy")
		}
		for _, issue := range result.PrivacyPolicyAnalysis.Issues {
			recommendations = append(recommendations, issue)
		}
	}

	// Terms of service recommendations
	if result.TermsOfServiceAnalysis != nil {
		for _, issue := range result.TermsOfServiceAnalysis.Issues {
			recommendations = append(recommendations, issue)
		}
	}

	// Certification recommendations
	if result.CertificationsAnalysis != nil {
		for _, missing := range result.CertificationsAnalysis.MissingCerts {
			recommendations = append(recommendations, fmt.Sprintf("Obtain required certification: %s", missing))
		}
		for _, expiring := range result.CertificationsAnalysis.ExpiringCerts {
			recommendations = append(recommendations, fmt.Sprintf("Renew expiring certification: %s", expiring.Name))
		}
	}

	return recommendations
}

func (ca *ComplianceAnalyzer) max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (ca *ComplianceAnalyzer) min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
