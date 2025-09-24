package performance

import (
	"fmt"
	"math/rand"
	"time"
)

// KYBTestScenarios provides predefined test scenarios for the KYB platform
type KYBTestScenarios struct {
	baseURL string
}

// NewKYBTestScenarios creates a new KYB test scenarios provider
func NewKYBTestScenarios(baseURL string) *KYBTestScenarios {
	return &KYBTestScenarios{
		baseURL: baseURL,
	}
}

// GetClassificationScenarios returns test scenarios for business classification
func (kts *KYBTestScenarios) GetClassificationScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "Classify Technology Business",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     "TechCorp Solutions",
				"description":       "Software development and cloud computing services",
				"website_url":       "https://techcorp.com",
				"industry_keywords": []string{"software", "technology", "cloud", "development"},
			},
			Weight:         30,
			ExpectedStatus: 200,
		},
		{
			Name:     "Classify Financial Services Business",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     "SecureBank Financial",
				"description":       "Personal and business banking services",
				"website_url":       "https://securebank.com",
				"industry_keywords": []string{"banking", "financial", "loans", "credit"},
			},
			Weight:         25,
			ExpectedStatus: 200,
		},
		{
			Name:     "Classify Healthcare Business",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     "MedCare Clinic",
				"description":       "Primary healthcare and medical services",
				"website_url":       "https://medcare.com",
				"industry_keywords": []string{"healthcare", "medical", "clinic", "doctor"},
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "Classify Retail Business",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     "Fashion Store",
				"description":       "Clothing and accessories retail",
				"website_url":       "https://fashionstore.com",
				"industry_keywords": []string{"retail", "clothing", "fashion", "accessories"},
			},
			Weight:         15,
			ExpectedStatus: 200,
		},
		{
			Name:     "Classify Manufacturing Business",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     "Industrial Manufacturing Co",
				"description":       "Heavy machinery and equipment manufacturing",
				"website_url":       "https://industrialmfg.com",
				"industry_keywords": []string{"manufacturing", "machinery", "industrial", "equipment"},
			},
			Weight:         10,
			ExpectedStatus: 200,
		},
	}
}

// GetRiskAssessmentScenarios returns test scenarios for risk assessment
func (kts *KYBTestScenarios) GetRiskAssessmentScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "Assess Low Risk Business",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":   kts.generateRandomBusinessID(),
				"business_name": "Legitimate Tech Company",
				"description":   "Software development and consulting services",
				"website_url":   "https://legit-tech.com",
				"industry":      "technology",
				"risk_factors":  []string{},
			},
			Weight:         40,
			ExpectedStatus: 200,
		},
		{
			Name:     "Assess Medium Risk Business",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":   kts.generateRandomBusinessID(),
				"business_name": "Online Gaming Platform",
				"description":   "Online gaming and entertainment services",
				"website_url":   "https://gaming-platform.com",
				"industry":      "entertainment",
				"risk_factors":  []string{"gambling", "online_gaming"},
			},
			Weight:         30,
			ExpectedStatus: 200,
		},
		{
			Name:     "Assess High Risk Business",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":   kts.generateRandomBusinessID(),
				"business_name": "Cryptocurrency Exchange",
				"description":   "Digital currency trading and exchange services",
				"website_url":   "https://crypto-exchange.com",
				"industry":      "financial_services",
				"risk_factors":  []string{"cryptocurrency", "high_risk", "unregulated"},
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "Assess Prohibited Business",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":   kts.generateRandomBusinessID(),
				"business_name": "Adult Entertainment Services",
				"description":   "Adult content and entertainment services",
				"website_url":   "https://adult-entertainment.com",
				"industry":      "entertainment",
				"risk_factors":  []string{"adult_content", "prohibited", "high_risk"},
			},
			Weight:         10,
			ExpectedStatus: 200,
		},
	}
}

// GetBusinessManagementScenarios returns test scenarios for business management
func (kts *KYBTestScenarios) GetBusinessManagementScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "Get Business Details",
			Method:   "GET",
			Endpoint: fmt.Sprintf("/api/v1/businesses/%s", kts.generateRandomBusinessID()),
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         25,
			ExpectedStatus: 200,
		},
		{
			Name:     "List Businesses",
			Method:   "GET",
			Endpoint: "/api/v1/businesses",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "Update Business Status",
			Method:   "PUT",
			Endpoint: fmt.Sprintf("/api/v1/businesses/%s/status", kts.generateRandomBusinessID()),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"status": "verified",
				"notes":  "Business verification completed",
			},
			Weight:         15,
			ExpectedStatus: 200,
		},
		{
			Name:     "Get Business Analytics",
			Method:   "GET",
			Endpoint: fmt.Sprintf("/api/v1/businesses/%s/analytics", kts.generateRandomBusinessID()),
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "Search Businesses",
			Method:   "GET",
			Endpoint: "/api/v1/businesses/search?q=technology&limit=20",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
	}
}

// GetUserManagementScenarios returns test scenarios for user management
func (kts *KYBTestScenarios) GetUserManagementScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "Get User Profile",
			Method:   "GET",
			Endpoint: fmt.Sprintf("/api/v1/users/%s", kts.generateRandomUserID()),
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         30,
			ExpectedStatus: 200,
		},
		{
			Name:     "Update User Profile",
			Method:   "PUT",
			Endpoint: fmt.Sprintf("/api/v1/users/%s", kts.generateRandomUserID()),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"name":  "Updated User Name",
				"email": "updated@example.com",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "List Users",
			Method:   "GET",
			Endpoint: "/api/v1/users",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         25,
			ExpectedStatus: 200,
		},
		{
			Name:     "Get User Permissions",
			Method:   "GET",
			Endpoint: fmt.Sprintf("/api/v1/users/%s/permissions", kts.generateRandomUserID()),
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         25,
			ExpectedStatus: 200,
		},
	}
}

// GetMonitoringScenarios returns test scenarios for monitoring endpoints
func (kts *KYBTestScenarios) GetMonitoringScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:           "Health Check",
			Method:         "GET",
			Endpoint:       "/api/v1/health",
			Headers:        map[string]string{},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:           "Metrics Endpoint",
			Method:         "GET",
			Endpoint:       "/api/v1/metrics",
			Headers:        map[string]string{},
			Weight:         15,
			ExpectedStatus: 200,
		},
		{
			Name:     "Performance Metrics",
			Method:   "GET",
			Endpoint: "/api/v1/performance",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         15,
			ExpectedStatus: 200,
		},
		{
			Name:     "System Status",
			Method:   "GET",
			Endpoint: "/api/v1/status",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
		{
			Name:     "Database Health",
			Method:   "GET",
			Endpoint: "/api/v1/health/database",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         15,
			ExpectedStatus: 200,
		},
		{
			Name:     "ML Service Health",
			Method:   "GET",
			Endpoint: "/api/v1/health/ml-service",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         15,
			ExpectedStatus: 200,
		},
	}
}

// GetComprehensiveScenarios returns all test scenarios combined
func (kts *KYBTestScenarios) GetComprehensiveScenarios() []TestScenario {
	var scenarios []TestScenario

	// Add all scenario types
	scenarios = append(scenarios, kts.GetClassificationScenarios()...)
	scenarios = append(scenarios, kts.GetRiskAssessmentScenarios()...)
	scenarios = append(scenarios, kts.GetBusinessManagementScenarios()...)
	scenarios = append(scenarios, kts.GetUserManagementScenarios()...)
	scenarios = append(scenarios, kts.GetMonitoringScenarios()...)

	return scenarios
}

// GetLoadTestScenarios returns scenarios optimized for load testing
func (kts *KYBTestScenarios) GetLoadTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "High-Frequency Classification",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     kts.generateRandomBusinessName(),
				"description":       kts.generateRandomDescription(),
				"website_url":       kts.generateRandomWebsiteURL(),
				"industry_keywords": kts.generateRandomKeywords(),
			},
			Weight:         50,
			ExpectedStatus: 200,
		},
		{
			Name:     "High-Frequency Risk Assessment",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":   kts.generateRandomBusinessID(),
				"business_name": kts.generateRandomBusinessName(),
				"description":   kts.generateRandomDescription(),
				"website_url":   kts.generateRandomWebsiteURL(),
				"industry":      kts.generateRandomIndustry(),
				"risk_factors":  kts.generateRandomRiskFactors(),
			},
			Weight:         30,
			ExpectedStatus: 200,
		},
		{
			Name:     "High-Frequency Business Lookup",
			Method:   "GET",
			Endpoint: fmt.Sprintf("/api/v1/businesses/%s", kts.generateRandomBusinessID()),
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         20,
			ExpectedStatus: 200,
		},
	}
}

// GetStressTestScenarios returns scenarios optimized for stress testing
func (kts *KYBTestScenarios) GetStressTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:     "Complex Classification Request",
			Method:   "POST",
			Endpoint: "/api/v1/classify",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_name":     kts.generateComplexBusinessName(),
				"description":       kts.generateComplexDescription(),
				"website_url":       kts.generateRandomWebsiteURL(),
				"industry_keywords": kts.generateComplexKeywords(),
				"additional_data":   kts.generateAdditionalData(),
			},
			Weight:         40,
			ExpectedStatus: 200,
		},
		{
			Name:     "Complex Risk Assessment",
			Method:   "POST",
			Endpoint: "/api/v1/risk/assess",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			Body: map[string]interface{}{
				"business_id":     kts.generateRandomBusinessID(),
				"business_name":   kts.generateComplexBusinessName(),
				"description":     kts.generateComplexDescription(),
				"website_url":     kts.generateRandomWebsiteURL(),
				"industry":        kts.generateRandomIndustry(),
				"risk_factors":    kts.generateComplexRiskFactors(),
				"historical_data": kts.generateHistoricalData(),
			},
			Weight:         35,
			ExpectedStatus: 200,
		},
		{
			Name:     "Bulk Business Search",
			Method:   "GET",
			Endpoint: "/api/v1/businesses/search?q=complex&limit=100&include_analytics=true",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			Weight:         25,
			ExpectedStatus: 200,
		},
	}
}

// Helper methods for generating test data

func (kts *KYBTestScenarios) generateRandomBusinessID() string {
	return fmt.Sprintf("biz_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

func (kts *KYBTestScenarios) generateRandomUserID() string {
	return fmt.Sprintf("user_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

func (kts *KYBTestScenarios) generateRandomBusinessName() string {
	names := []string{
		"TechCorp Solutions", "InnovateLab", "DataFlow Systems", "CloudTech Inc",
		"SecureBank Financial", "FinanceFirst", "Capital Partners", "Investment Group",
		"MedCare Clinic", "HealthFirst", "Wellness Center", "Medical Services",
		"Fashion Store", "Retail Plus", "Shopping Center", "Market Place",
		"Industrial Manufacturing", "Production Co", "Manufacturing Ltd", "Factory Inc",
	}
	return names[rand.Intn(len(names))]
}

func (kts *KYBTestScenarios) generateRandomDescription() string {
	descriptions := []string{
		"Leading technology company providing innovative solutions",
		"Financial services firm specializing in investment management",
		"Healthcare provider offering comprehensive medical services",
		"Retail business focused on customer satisfaction and quality products",
		"Manufacturing company producing high-quality industrial equipment",
		"Software development company creating cutting-edge applications",
		"Consulting firm providing strategic business advice",
		"E-commerce platform connecting buyers and sellers worldwide",
	}
	return descriptions[rand.Intn(len(descriptions))]
}

func (kts *KYBTestScenarios) generateRandomWebsiteURL() string {
	domains := []string{
		"techcorp.com", "innovate.com", "dataflow.com", "cloudtech.com",
		"securebank.com", "finance.com", "capital.com", "investment.com",
		"medcare.com", "health.com", "wellness.com", "medical.com",
		"fashion.com", "retail.com", "shopping.com", "market.com",
	}
	return fmt.Sprintf("https://%s", domains[rand.Intn(len(domains))])
}

func (kts *KYBTestScenarios) generateRandomKeywords() []string {
	keywordSets := [][]string{
		{"technology", "software", "development", "innovation"},
		{"financial", "banking", "investment", "capital"},
		{"healthcare", "medical", "wellness", "treatment"},
		{"retail", "shopping", "commerce", "sales"},
		{"manufacturing", "production", "industrial", "equipment"},
	}
	return keywordSets[rand.Intn(len(keywordSets))]
}

func (kts *KYBTestScenarios) generateRandomIndustry() string {
	industries := []string{
		"technology", "financial_services", "healthcare", "retail", "manufacturing",
		"entertainment", "education", "real_estate", "transportation", "energy",
	}
	return industries[rand.Intn(len(industries))]
}

func (kts *KYBTestScenarios) generateRandomRiskFactors() []string {
	riskFactors := []string{
		"high_volume", "international", "cash_intensive", "new_business",
		"regulatory_compliance", "data_privacy", "cybersecurity", "operational_risk",
	}

	// Return 0-3 random risk factors
	count := rand.Intn(4)
	if count == 0 {
		return []string{}
	}

	selected := make([]string, count)
	for i := 0; i < count; i++ {
		selected[i] = riskFactors[rand.Intn(len(riskFactors))]
	}
	return selected
}

func (kts *KYBTestScenarios) generateComplexBusinessName() string {
	baseNames := []string{
		"Advanced Technology Solutions", "Global Financial Services Group",
		"Comprehensive Healthcare Systems", "International Retail Corporation",
		"Industrial Manufacturing Enterprises", "Digital Innovation Labs",
	}
	return baseNames[rand.Intn(len(baseNames))]
}

func (kts *KYBTestScenarios) generateComplexDescription() string {
	descriptions := []string{
		"Multi-national technology corporation specializing in artificial intelligence, machine learning, cloud computing, and digital transformation services for enterprise clients across various industries including healthcare, finance, manufacturing, and retail sectors.",
		"Comprehensive financial services institution providing investment banking, wealth management, commercial lending, insurance products, and fintech solutions to individual and corporate clients with global operations and regulatory compliance across multiple jurisdictions.",
		"Integrated healthcare delivery system offering primary care, specialty medical services, diagnostic imaging, laboratory services, telemedicine, and health information technology solutions to patients and healthcare providers in urban and rural communities.",
	}
	return descriptions[rand.Intn(len(descriptions))]
}

func (kts *KYBTestScenarios) generateComplexKeywords() []string {
	complexKeywordSets := [][]string{
		{"artificial intelligence", "machine learning", "cloud computing", "digital transformation", "enterprise software", "data analytics", "cybersecurity", "blockchain"},
		{"investment banking", "wealth management", "commercial lending", "insurance", "fintech", "regulatory compliance", "risk management", "financial planning"},
		{"primary care", "specialty medicine", "diagnostic imaging", "laboratory services", "telemedicine", "health information technology", "patient care", "medical research"},
	}
	return complexKeywordSets[rand.Intn(len(complexKeywordSets))]
}

func (kts *KYBTestScenarios) generateComplexRiskFactors() []string {
	complexRiskFactors := []string{
		"high_volume_transactions", "international_operations", "regulatory_compliance", "data_privacy_concerns",
		"cybersecurity_risks", "operational_complexity", "market_volatility", "regulatory_changes",
		"technology_dependencies", "third_party_vendors", "geopolitical_risks", "environmental_factors",
	}

	// Return 3-6 random complex risk factors
	count := 3 + rand.Intn(4)
	selected := make([]string, count)
	for i := 0; i < count; i++ {
		selected[i] = complexRiskFactors[rand.Intn(len(complexRiskFactors))]
	}
	return selected
}

func (kts *KYBTestScenarios) generateAdditionalData() map[string]interface{} {
	return map[string]interface{}{
		"founded_year":      2000 + rand.Intn(24),
		"employee_count":    10 + rand.Intn(1000),
		"annual_revenue":    rand.Float64() * 10000000,
		"locations":         []string{"New York", "San Francisco", "London", "Tokyo"},
		"certifications":    []string{"ISO 27001", "SOC 2", "PCI DSS"},
		"compliance_status": "compliant",
	}
}

func (kts *KYBTestScenarios) generateHistoricalData() map[string]interface{} {
	return map[string]interface{}{
		"previous_assessments": []map[string]interface{}{
			{
				"date":   time.Now().AddDate(0, -6, 0).Format("2006-01-02"),
				"score":  0.85,
				"status": "approved",
			},
			{
				"date":   time.Now().AddDate(0, -12, 0).Format("2006-01-02"),
				"score":  0.92,
				"status": "approved",
			},
		},
		"transaction_history": map[string]interface{}{
			"total_volume":        rand.Float64() * 1000000,
			"transaction_count":   100 + rand.Intn(1000),
			"average_transaction": rand.Float64() * 10000,
		},
		"compliance_history": []string{
			"audit_passed_2023", "compliance_review_2023", "regulatory_update_2023",
		},
	}
}
