package test

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
)

// ComprehensiveTestDataset represents a comprehensive test dataset for classification accuracy testing
type ComprehensiveTestDataset struct {
	TestCases []ClassificationTestCase
}

// ClassificationTestCase represents a single test case with known expected results
type ClassificationTestCase struct {
	ID                 string
	Name               string
	BusinessName       string
	Description        string
	WebsiteURL         string
	ExpectedIndustry   string
	ExpectedConfidence float64
	ExpectedMCCCodes   []string
	ExpectedSICCodes   []string
	ExpectedNAICSCodes []string
	TestCategory       string
	DifficultyLevel    string
	Keywords           []string
	BusinessType       string
	GeographicRegion   string
	CompanySize        string
	ExpectedKeywords   []string
	Notes              string
}

// NewComprehensiveTestDataset creates a new comprehensive test dataset
func NewComprehensiveTestDataset() *ComprehensiveTestDataset {
	return &ComprehensiveTestDataset{
		TestCases: []ClassificationTestCase{
			// =====================================================
			// TECHNOLOGY & SOFTWARE INDUSTRY
			// =====================================================
			{
				ID:                 "tech-001",
				Name:               "Software Development Company",
				BusinessName:       "TechCorp Solutions",
				Description:        "We develop innovative software solutions using cloud technology and AI for businesses. Our team specializes in custom software development, web applications, and mobile apps.",
				WebsiteURL:         "https://techcorp.com",
				ExpectedIndustry:   "Technology",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"5734", "7372"},
				ExpectedSICCodes:   []string{"7372", "7371"},
				ExpectedNAICSCodes: []string{"541511", "511210"},
				TestCategory:       "Technology",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"software", "development", "cloud", "technology", "AI", "web applications", "mobile apps"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"software", "technology", "cloud", "development"},
				Notes:              "Clear technology company with strong keyword matches",
			},
			{
				ID:                 "tech-002",
				Name:               "Cloud Computing Provider",
				BusinessName:       "CloudScale Infrastructure",
				Description:        "Enterprise cloud infrastructure and platform services. We provide scalable cloud solutions, serverless computing, and cloud migration services.",
				WebsiteURL:         "https://cloudscale.com",
				ExpectedIndustry:   "Cloud Computing",
				ExpectedConfidence: 0.80,
				ExpectedMCCCodes:   []string{"7372", "5734"},
				ExpectedSICCodes:   []string{"7373", "7372"},
				ExpectedNAICSCodes: []string{"541512", "518210"},
				TestCategory:       "Technology",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"cloud", "infrastructure", "platform", "serverless", "migration"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"cloud", "infrastructure", "platform"},
				Notes:              "Cloud computing specialization",
			},
			{
				ID:                 "tech-003",
				Name:               "AI/ML Startup",
				BusinessName:       "NeuralNet AI",
				Description:        "Artificial intelligence and machine learning services. We develop chatbots, intelligent automation, and AI-powered business solutions.",
				WebsiteURL:         "https://neuralnet.ai",
				ExpectedIndustry:   "Artificial Intelligence",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"7372", "5734"},
				ExpectedSICCodes:   []string{"7372", "7371"},
				ExpectedNAICSCodes: []string{"541511", "518210"},
				TestCategory:       "Technology",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"AI", "machine learning", "artificial intelligence", "chatbots", "automation"},
				BusinessType:       "Startup",
				GeographicRegion:   "North America",
				CompanySize:        "Small",
				ExpectedKeywords:   []string{"AI", "artificial intelligence", "machine learning"},
				Notes:              "AI/ML specialization with emerging technology focus",
			},

			// =====================================================
			// HEALTHCARE & MEDICAL INDUSTRY
			// =====================================================
			{
				ID:                 "health-001",
				Name:               "Medical Center",
				BusinessName:       "Medical Center Plus",
				Description:        "Comprehensive medical services and patient care at our modern clinic. We provide primary care, specialty services, and emergency medical treatment.",
				WebsiteURL:         "https://medicalcenter.com",
				ExpectedIndustry:   "Healthcare",
				ExpectedConfidence: 0.90,
				ExpectedMCCCodes:   []string{"8011", "8021"},
				ExpectedSICCodes:   []string{"8011", "8062"},
				ExpectedNAICSCodes: []string{"621111", "622110"},
				TestCategory:       "Healthcare",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"medical", "healthcare", "clinic", "patient care", "primary care", "emergency"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"medical", "healthcare", "clinic", "patient"},
				Notes:              "Clear healthcare provider with strong medical keywords",
			},
			{
				ID:                 "health-002",
				Name:               "Medical Technology Company",
				BusinessName:       "MedTech Innovations",
				Description:        "Medical devices, diagnostic tools, and health technology innovations. We develop cutting-edge medical equipment and diagnostic solutions.",
				WebsiteURL:         "https://medtech.com",
				ExpectedIndustry:   "Medical Technology",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"5047", "8011"},
				ExpectedSICCodes:   []string{"3841", "3842"},
				ExpectedNAICSCodes: []string{"334510", "339112"},
				TestCategory:       "Healthcare",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"medical devices", "diagnostic", "health technology", "medical equipment"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"medical", "devices", "diagnostic", "technology"},
				Notes:              "Medical technology with device focus",
			},
			{
				ID:                 "health-003",
				Name:               "Pharmaceutical Company",
				BusinessName:       "PharmaCorp Research",
				Description:        "Drug development, medical research, and pharmaceutical services. We specialize in innovative drug discovery and clinical research.",
				WebsiteURL:         "https://pharmacorp.com",
				ExpectedIndustry:   "Pharmaceuticals",
				ExpectedConfidence: 0.88,
				ExpectedMCCCodes:   []string{"5122", "5047"},
				ExpectedSICCodes:   []string{"2834", "2836"},
				ExpectedNAICSCodes: []string{"325412", "541711"},
				TestCategory:       "Healthcare",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"pharmaceutical", "drug development", "medical research", "clinical research"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"pharmaceutical", "drug", "research", "medical"},
				Notes:              "Pharmaceutical company with research focus",
			},

			// =====================================================
			// FINANCE & BANKING INDUSTRY
			// =====================================================
			{
				ID:                 "finance-001",
				Name:               "Commercial Bank",
				BusinessName:       "First National Bank",
				Description:        "Personal and business banking services with investment and credit solutions. We provide comprehensive financial services including loans, deposits, and investment management.",
				WebsiteURL:         "https://firstnational.com",
				ExpectedIndustry:   "Finance",
				ExpectedConfidence: 0.92,
				ExpectedMCCCodes:   []string{"6011", "6012"},
				ExpectedSICCodes:   []string{"6021", "6022"},
				ExpectedNAICSCodes: []string{"522110", "523110"},
				TestCategory:       "Finance",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"banking", "finance", "loans", "deposits", "investment", "credit"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"banking", "finance", "loans", "investment"},
				Notes:              "Traditional commercial bank with clear financial services",
			},
			{
				ID:                 "finance-002",
				Name:               "Fintech Startup",
				BusinessName:       "PayTech Solutions",
				Description:        "Financial technology, digital banking, and payment solutions. We provide innovative fintech services including mobile payments and digital wallets.",
				WebsiteURL:         "https://paytech.com",
				ExpectedIndustry:   "Fintech",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"6011", "5999"},
				ExpectedSICCodes:   []string{"7372", "7389"},
				ExpectedNAICSCodes: []string{"522320", "518210"},
				TestCategory:       "Finance",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"fintech", "digital banking", "payment solutions", "mobile payments", "digital wallets"},
				BusinessType:       "Startup",
				GeographicRegion:   "North America",
				CompanySize:        "Small",
				ExpectedKeywords:   []string{"fintech", "digital", "payment", "banking"},
				Notes:              "Fintech company with digital payment focus",
			},
			{
				ID:                 "finance-003",
				Name:               "Insurance Company",
				BusinessName:       "SecureLife Insurance",
				Description:        "Life, health, property, and casualty insurance services. We provide comprehensive insurance coverage and risk management solutions.",
				WebsiteURL:         "https://securelife.com",
				ExpectedIndustry:   "Insurance",
				ExpectedConfidence: 0.88,
				ExpectedMCCCodes:   []string{"6300", "5960"},
				ExpectedSICCodes:   []string{"6311", "6321"},
				ExpectedNAICSCodes: []string{"524113", "524114"},
				TestCategory:       "Finance",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"insurance", "life insurance", "health insurance", "property insurance", "casualty"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"insurance", "life", "health", "property"},
				Notes:              "Insurance company with multiple coverage types",
			},

			// =====================================================
			// RETAIL & E-COMMERCE INDUSTRY
			// =====================================================
			{
				ID:                 "retail-001",
				Name:               "Online Retail Store",
				BusinessName:       "ShopSmart Retail",
				Description:        "Online and offline retail store offering shopping experience for customers. We sell consumer goods, electronics, and home products.",
				WebsiteURL:         "https://shopsmart.com",
				ExpectedIndustry:   "Retail",
				ExpectedConfidence: 0.80,
				ExpectedMCCCodes:   []string{"5310", "5732"},
				ExpectedSICCodes:   []string{"5311", "5731"},
				ExpectedNAICSCodes: []string{"454110", "452111"},
				TestCategory:       "Retail",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"retail", "shopping", "consumer goods", "electronics", "home products"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"retail", "shopping", "consumer", "goods"},
				Notes:              "Traditional retail with online presence",
			},
			{
				ID:                 "retail-002",
				Name:               "E-commerce Platform",
				BusinessName:       "MarketPlace Pro",
				Description:        "E-commerce platform and marketplace services. We provide online shopping solutions, digital storefronts, and e-commerce infrastructure.",
				WebsiteURL:         "https://marketplacepro.com",
				ExpectedIndustry:   "Retail",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"5310", "5999"},
				ExpectedSICCodes:   []string{"7372", "5311"},
				ExpectedNAICSCodes: []string{"454110", "518210"},
				TestCategory:       "Retail",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"e-commerce", "marketplace", "online shopping", "digital storefronts"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"e-commerce", "marketplace", "online", "shopping"},
				Notes:              "E-commerce platform with marketplace focus",
			},

			// =====================================================
			// MANUFACTURING INDUSTRY
			// =====================================================
			{
				ID:                 "mfg-001",
				Name:               "Industrial Manufacturing",
				BusinessName:       "Industrial Manufacturing Co",
				Description:        "Factory production of industrial equipment and manufacturing solutions. We specialize in heavy machinery and industrial automation.",
				WebsiteURL:         "https://industrialmfg.com",
				ExpectedIndustry:   "Manufacturing",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"5085", "5047"},
				ExpectedSICCodes:   []string{"3531", "3532"},
				ExpectedNAICSCodes: []string{"333120", "332996"},
				TestCategory:       "Manufacturing",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"manufacturing", "factory", "production", "industrial equipment", "machinery", "automation"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"manufacturing", "factory", "production", "industrial"},
				Notes:              "Industrial manufacturing with equipment focus",
			},
			{
				ID:                 "mfg-002",
				Name:               "Food Manufacturing",
				BusinessName:       "FreshFood Processing",
				Description:        "Food processing and manufacturing company. We produce packaged foods, beverages, and food products for retail distribution.",
				WebsiteURL:         "https://freshfood.com",
				ExpectedIndustry:   "Manufacturing",
				ExpectedConfidence: 0.80,
				ExpectedMCCCodes:   []string{"5411", "5999"},
				ExpectedSICCodes:   []string{"2011", "2086"},
				ExpectedNAICSCodes: []string{"311111", "312111"},
				TestCategory:       "Manufacturing",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"food processing", "manufacturing", "packaged foods", "beverages", "food products"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"food", "processing", "manufacturing", "packaged"},
				Notes:              "Food manufacturing with processing focus",
			},

			// =====================================================
			// PROFESSIONAL SERVICES INDUSTRY
			// =====================================================
			{
				ID:                 "services-001",
				Name:               "Management Consulting",
				BusinessName:       "Strategic Consulting Group",
				Description:        "Management consulting and business strategy services. We provide strategic planning, organizational development, and business transformation consulting.",
				WebsiteURL:         "https://strategicconsulting.com",
				ExpectedIndustry:   "Professional Services",
				ExpectedConfidence: 0.80,
				ExpectedMCCCodes:   []string{"7392", "8999"},
				ExpectedSICCodes:   []string{"8742", "8741"},
				ExpectedNAICSCodes: []string{"541611", "541612"},
				TestCategory:       "Professional Services",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"consulting", "management", "strategy", "business transformation", "organizational development"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"consulting", "management", "strategy", "business"},
				Notes:              "Management consulting with strategy focus",
			},
			{
				ID:                 "services-002",
				Name:               "Legal Services",
				BusinessName:       "Legal Associates LLP",
				Description:        "Legal services and law firm. We provide corporate law, litigation, and legal advisory services to businesses and individuals.",
				WebsiteURL:         "https://legalassociates.com",
				ExpectedIndustry:   "Legal Services",
				ExpectedConfidence: 0.90,
				ExpectedMCCCodes:   []string{"8111", "8999"},
				ExpectedSICCodes:   []string{"8111", "8112"},
				ExpectedNAICSCodes: []string{"541110", "541199"},
				TestCategory:       "Professional Services",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"legal", "law", "litigation", "corporate law", "legal advisory"},
				BusinessType:       "Partnership",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"legal", "law", "litigation", "corporate"},
				Notes:              "Legal services with law firm focus",
			},

			// =====================================================
			// REAL ESTATE INDUSTRY
			// =====================================================
			{
				ID:                 "realestate-001",
				Name:               "Real Estate Agency",
				BusinessName:       "Premier Real Estate",
				Description:        "Real estate agency and property services. We provide residential and commercial real estate sales, leasing, and property management services.",
				WebsiteURL:         "https://premierrealestate.com",
				ExpectedIndustry:   "Real Estate",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"6513", "6514"},
				ExpectedSICCodes:   []string{"6531", "6512"},
				ExpectedNAICSCodes: []string{"531210", "531312"},
				TestCategory:       "Real Estate",
				DifficultyLevel:    "Easy",
				Keywords:           []string{"real estate", "property", "residential", "commercial", "sales", "leasing", "property management"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"real estate", "property", "residential", "commercial"},
				Notes:              "Real estate agency with sales and management focus",
			},

			// =====================================================
			// EDUCATION INDUSTRY
			// =====================================================
			{
				ID:                 "education-001",
				Name:               "Educational Technology",
				BusinessName:       "EduTech Solutions",
				Description:        "Educational technology and e-learning solutions. We provide online learning platforms, educational software, and digital learning tools.",
				WebsiteURL:         "https://edutech.com",
				ExpectedIndustry:   "Education",
				ExpectedConfidence: 0.80,
				ExpectedMCCCodes:   []string{"7372", "5999"},
				ExpectedSICCodes:   []string{"7372", "8299"},
				ExpectedNAICSCodes: []string{"611710", "518210"},
				TestCategory:       "Education",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"education", "e-learning", "online learning", "educational software", "digital learning"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"education", "e-learning", "online", "learning"},
				Notes:              "Educational technology with e-learning focus",
			},

			// =====================================================
			// ENERGY INDUSTRY
			// =====================================================
			{
				ID:                 "energy-001",
				Name:               "Renewable Energy",
				BusinessName:       "GreenEnergy Solutions",
				Description:        "Renewable energy and clean technology solutions. We provide solar energy, wind power, and sustainable energy infrastructure.",
				WebsiteURL:         "https://greenenergy.com",
				ExpectedIndustry:   "Energy",
				ExpectedConfidence: 0.85,
				ExpectedMCCCodes:   []string{"4900", "5999"},
				ExpectedSICCodes:   []string{"4911", "4953"},
				ExpectedNAICSCodes: []string{"221111", "221115"},
				TestCategory:       "Energy",
				DifficultyLevel:    "Medium",
				Keywords:           []string{"renewable energy", "solar", "wind power", "clean technology", "sustainable energy"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Large",
				ExpectedKeywords:   []string{"renewable", "energy", "solar", "wind"},
				Notes:              "Renewable energy with clean technology focus",
			},

			// =====================================================
			// CHALLENGING/EDGE CASES
			// =====================================================
			{
				ID:                 "edge-001",
				Name:               "Mixed Industry Business",
				BusinessName:       "TechHealth Solutions",
				Description:        "Combining technology and healthcare for innovative medical software. We develop healthcare technology solutions and medical software applications.",
				WebsiteURL:         "https://techhealth.com",
				ExpectedIndustry:   "Healthcare", // Should match healthcare keywords
				ExpectedConfidence: 0.75,
				ExpectedMCCCodes:   []string{"8011", "5734"},
				ExpectedSICCodes:   []string{"8011", "7372"},
				ExpectedNAICSCodes: []string{"621111", "541511"},
				TestCategory:       "Edge Cases",
				DifficultyLevel:    "Hard",
				Keywords:           []string{"technology", "healthcare", "medical software", "healthcare technology"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Medium",
				ExpectedKeywords:   []string{"healthcare", "medical", "technology", "software"},
				Notes:              "Mixed industry business - should prioritize healthcare keywords",
			},
			{
				ID:                 "edge-002",
				Name:               "Generic Business",
				BusinessName:       "General Business Services",
				Description:        "General business services and consulting. We provide various business solutions and professional services.",
				WebsiteURL:         "https://generalbusiness.com",
				ExpectedIndustry:   "General Business",
				ExpectedConfidence: 0.50,
				ExpectedMCCCodes:   []string{"8999"},
				ExpectedSICCodes:   []string{"7389"},
				ExpectedNAICSCodes: []string{"541990"},
				TestCategory:       "Edge Cases",
				DifficultyLevel:    "Hard",
				Keywords:           []string{"business", "services", "consulting", "solutions"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Small",
				ExpectedKeywords:   []string{"business", "services", "consulting"},
				Notes:              "Generic business with low specificity",
			},
			{
				ID:                 "edge-003",
				Name:               "Very Short Description",
				BusinessName:       "Quick Corp",
				Description:        "Business services.",
				WebsiteURL:         "https://quickcorp.com",
				ExpectedIndustry:   "General Business",
				ExpectedConfidence: 0.30,
				ExpectedMCCCodes:   []string{"8999"},
				ExpectedSICCodes:   []string{"7389"},
				ExpectedNAICSCodes: []string{"541990"},
				TestCategory:       "Edge Cases",
				DifficultyLevel:    "Hard",
				Keywords:           []string{"business", "services"},
				BusinessType:       "Corporation",
				GeographicRegion:   "North America",
				CompanySize:        "Small",
				ExpectedKeywords:   []string{"business", "services"},
				Notes:              "Very short description with minimal information",
			},
		},
	}
}

// GetTestCasesByCategory returns test cases filtered by category
func (d *ComprehensiveTestDataset) GetTestCasesByCategory(category string) []ClassificationTestCase {
	var filtered []ClassificationTestCase
	for _, tc := range d.TestCases {
		if tc.TestCategory == category {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetTestCasesByDifficulty returns test cases filtered by difficulty level
func (d *ComprehensiveTestDataset) GetTestCasesByDifficulty(difficulty string) []ClassificationTestCase {
	var filtered []ClassificationTestCase
	for _, tc := range d.TestCases {
		if tc.DifficultyLevel == difficulty {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetTestCasesByIndustry returns test cases filtered by expected industry
func (d *ComprehensiveTestDataset) GetTestCasesByIndustry(industry string) []ClassificationTestCase {
	var filtered []ClassificationTestCase
	for _, tc := range d.TestCases {
		if tc.ExpectedIndustry == industry {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetStatistics returns statistics about the test dataset
func (d *ComprehensiveTestDataset) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Total test cases
	stats["total_test_cases"] = len(d.TestCases)

	// Category breakdown
	categories := make(map[string]int)
	for _, tc := range d.TestCases {
		categories[tc.TestCategory]++
	}
	stats["categories"] = categories

	// Difficulty breakdown
	difficulties := make(map[string]int)
	for _, tc := range d.TestCases {
		difficulties[tc.DifficultyLevel]++
	}
	stats["difficulties"] = difficulties

	// Industry breakdown
	industries := make(map[string]int)
	for _, tc := range d.TestCases {
		industries[tc.ExpectedIndustry]++
	}
	stats["industries"] = industries

	// Business type breakdown
	businessTypes := make(map[string]int)
	for _, tc := range d.TestCases {
		businessTypes[tc.BusinessType]++
	}
	stats["business_types"] = businessTypes

	// Company size breakdown
	companySizes := make(map[string]int)
	for _, tc := range d.TestCases {
		companySizes[tc.CompanySize]++
	}
	stats["company_sizes"] = companySizes

	// Average confidence score
	totalConfidence := 0.0
	for _, tc := range d.TestCases {
		totalConfidence += tc.ExpectedConfidence
	}
	stats["average_confidence"] = totalConfidence / float64(len(d.TestCases))

	return stats
}

// RunClassificationAccuracyTest runs the comprehensive classification accuracy test
func RunClassificationAccuracyTest(t *testing.T, classifier *classification.ClassificationCodeGenerator, dataset *ComprehensiveTestDataset) {
	ctx := context.Background()

	// Test statistics
	var totalTests int
	var passedTests int
	var totalConfidence float64
	var accuracyByCategory = make(map[string]int)
	var accuracyByDifficulty = make(map[string]int)
	var accuracyByIndustry = make(map[string]int)

	// Run tests for each category
	categories := []string{"Technology", "Healthcare", "Finance", "Retail", "Manufacturing", "Professional Services", "Real Estate", "Education", "Energy", "Edge Cases"}

	for _, category := range categories {
		testCases := dataset.GetTestCasesByCategory(category)
		t.Logf("🧪 Testing %s category with %d test cases", category, len(testCases))

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				// Run classification
				startTime := time.Now()
				result, err := classifier.GenerateClassificationCodes(
					ctx,
					tc.Keywords,
					tc.BusinessName,
					tc.ExpectedConfidence,
				)
				responseTime := time.Since(startTime)

				// Validate results
				if err != nil {
					t.Errorf("❌ Classification failed for %s: %v", tc.Name, err)
					return
				}

				if result == nil {
					t.Errorf("❌ No classification result for %s", tc.Name)
					return
				}

				// Check confidence score
				actualConfidence := calculateOverallConfidence(result)
				if actualConfidence < tc.ExpectedConfidence-0.1 {
					t.Errorf("❌ Low confidence for %s: expected >= %.2f, got %.2f",
						tc.Name, tc.ExpectedConfidence-0.1, actualConfidence)
				}

				// Check if we have expected codes
				hasExpectedCodes := checkExpectedCodes(result, tc.ExpectedMCCCodes, tc.ExpectedSICCodes, tc.ExpectedNAICSCodes)
				if !hasExpectedCodes {
					t.Logf("⚠️  Missing expected codes for %s", tc.Name)
				}

				// Update statistics
				totalTests++
				totalConfidence += actualConfidence

				if actualConfidence >= tc.ExpectedConfidence-0.1 {
					passedTests++
					accuracyByCategory[category]++
					accuracyByDifficulty[tc.DifficultyLevel]++
					accuracyByIndustry[tc.ExpectedIndustry]++
				}

				// Log results
				t.Logf("✅ %s: Confidence %.2f, Response time %v",
					tc.Name, actualConfidence, responseTime)

				// Log detailed results for debugging
				if t.Failed() {
					t.Logf("📊 Detailed results for %s:", tc.Name)
					t.Logf("   Expected: %s (%.2f)", tc.ExpectedIndustry, tc.ExpectedConfidence)
					t.Logf("   Actual confidence: %.2f", actualConfidence)
					t.Logf("   MCC codes: %d", len(result.MCC))
					t.Logf("   SIC codes: %d", len(result.SIC))
					t.Logf("   NAICS codes: %d", len(result.NAICS))
				}
			})
		}
	}

	// Calculate and log overall statistics
	overallAccuracy := float64(passedTests) / float64(totalTests) * 100
	averageConfidence := totalConfidence / float64(totalTests)

	t.Logf("📊 Classification Accuracy Test Results:")
	t.Logf("   Total tests: %d", totalTests)
	t.Logf("   Passed tests: %d", passedTests)
	t.Logf("   Overall accuracy: %.1f%%", overallAccuracy)
	t.Logf("   Average confidence: %.2f", averageConfidence)

	// Log accuracy by category
	t.Logf("📈 Accuracy by Category:")
	for category, passed := range accuracyByCategory {
		total := 0
		for _, tc := range dataset.TestCases {
			if tc.TestCategory == category {
				total++
			}
		}
		accuracy := float64(passed) / float64(total) * 100
		t.Logf("   %s: %d/%d (%.1f%%)", category, passed, total, accuracy)
	}

	// Log accuracy by difficulty
	t.Logf("📈 Accuracy by Difficulty:")
	for difficulty, passed := range accuracyByDifficulty {
		total := 0
		for _, tc := range dataset.TestCases {
			if tc.DifficultyLevel == difficulty {
				total++
			}
		}
		accuracy := float64(passed) / float64(total) * 100
		t.Logf("   %s: %d/%d (%.1f%%)", difficulty, passed, total, accuracy)
	}

	// Assert minimum accuracy threshold
	if overallAccuracy < 70.0 {
		t.Errorf("❌ Overall accuracy %.1f%% is below minimum threshold of 70%%", overallAccuracy)
	}
}

// Helper functions

// calculateOverallConfidence calculates the overall confidence score from classification results
func calculateOverallConfidence(result *classification.ClassificationCodesInfo) float64 {
	if result == nil {
		return 0.0
	}

	totalConfidence := 0.0
	count := 0

	// Calculate average confidence from all codes
	for _, mcc := range result.MCC {
		totalConfidence += mcc.Confidence
		count++
	}
	for _, sic := range result.SIC {
		totalConfidence += sic.Confidence
		count++
	}
	for _, naics := range result.NAICS {
		totalConfidence += naics.Confidence
		count++
	}

	if count == 0 {
		return 0.0
	}

	return totalConfidence / float64(count)
}

// checkExpectedCodes checks if the classification result contains expected codes
func checkExpectedCodes(result *classification.ClassificationCodesInfo, expectedMCC, expectedSIC, expectedNAICS []string) bool {
	if result == nil {
		return false
	}

	// Check MCC codes
	mccFound := false
	for _, expected := range expectedMCC {
		for _, actual := range result.MCC {
			if actual.Code == expected {
				mccFound = true
				break
			}
		}
		if mccFound {
			break
		}
	}

	// Check SIC codes
	sicFound := false
	for _, expected := range expectedSIC {
		for _, actual := range result.SIC {
			if actual.Code == expected {
				sicFound = true
				break
			}
		}
		if sicFound {
			break
		}
	}

	// Check NAICS codes
	naicsFound := false
	for _, expected := range expectedNAICS {
		for _, actual := range result.NAICS {
			if actual.Code == expected {
				naicsFound = true
				break
			}
		}
		if naicsFound {
			break
		}
	}

	// Return true if at least one expected code is found in each category
	return mccFound && sicFound && naicsFound
}
