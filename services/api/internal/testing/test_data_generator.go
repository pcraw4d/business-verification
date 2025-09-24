package testing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

// TestDataGenerator generates comprehensive test data for classification accuracy testing
type TestDataGenerator struct {
	db     *sql.DB
	logger *log.Logger
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(db *sql.DB, logger *log.Logger) *TestDataGenerator {
	return &TestDataGenerator{
		db:     db,
		logger: logger,
	}
}

// GenerateComprehensiveTestData generates comprehensive test samples across all categories
func (tdg *TestDataGenerator) GenerateComprehensiveTestData(ctx context.Context) error {
	tdg.logger.Println("Generating comprehensive test data for classification accuracy testing...")

	// Generate test samples for each category
	categories := []string{"primary", "edge_case", "high_risk", "emerging", "crosswalk", "confidence"}

	for _, category := range categories {
		tdg.logger.Printf("Generating test data for category: %s", category)

		switch category {
		case "primary":
			if err := tdg.generatePrimaryIndustrySamples(ctx); err != nil {
				return fmt.Errorf("failed to generate primary industry samples: %w", err)
			}
		case "edge_case":
			if err := tdg.generateEdgeCaseSamples(ctx); err != nil {
				return fmt.Errorf("failed to generate edge case samples: %w", err)
			}
		case "high_risk":
			if err := tdg.generateHighRiskSamples(ctx); err != nil {
				return fmt.Errorf("failed to generate high risk samples: %w", err)
			}
		case "emerging":
			if err := tdg.generateEmergingIndustrySamples(ctx); err != nil {
				return fmt.Errorf("failed to generate emerging industry samples: %w", err)
			}
		case "crosswalk":
			if err := tdg.generateCrosswalkValidationSamples(ctx); err != nil {
				return fmt.Errorf("failed to generate crosswalk validation samples: %w", err)
			}
		case "confidence":
			if err := tdg.generateConfidenceValidationSamples(ctx); err != nil {
				return fmt.Errorf("failed to generate confidence validation samples: %w", err)
			}
		}
	}

	tdg.logger.Println("Comprehensive test data generation completed successfully")
	return nil
}

// generatePrimaryIndustrySamples generates samples for primary industries
func (tdg *TestDataGenerator) generatePrimaryIndustrySamples(ctx context.Context) error {
	samples := []TestSampleData{
		// Technology
		{
			BusinessName:     "Google LLC",
			Description:      "Multinational technology company specializing in Internet-related services and products",
			WebsiteURL:       "https://google.com",
			ExpectedMCC:      "5733",
			ExpectedNAICS:    "541511",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "primary",
			Confidence:       0.98,
		},
		{
			BusinessName:     "Facebook Inc.",
			Description:      "Social media and social networking service company",
			WebsiteURL:       "https://facebook.com",
			ExpectedMCC:      "5733",
			ExpectedNAICS:    "518210",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "primary",
			Confidence:       0.95,
		},

		// Financial Services
		{
			BusinessName:     "Bank of America Corporation",
			Description:      "Multinational investment bank and financial services holding company",
			WebsiteURL:       "https://bankofamerica.com",
			ExpectedMCC:      "6012",
			ExpectedNAICS:    "522110",
			ExpectedSIC:      "6021",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "primary",
			Confidence:       0.98,
		},
		{
			BusinessName:     "Goldman Sachs Group Inc.",
			Description:      "Multinational investment bank and financial services company",
			WebsiteURL:       "https://goldmansachs.com",
			ExpectedMCC:      "6012",
			ExpectedNAICS:    "523110",
			ExpectedSIC:      "6211",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "primary",
			Confidence:       0.97,
		},

		// Healthcare
		{
			BusinessName:     "Pfizer Inc.",
			Description:      "Multinational pharmaceutical and biotechnology corporation",
			WebsiteURL:       "https://pfizer.com",
			ExpectedMCC:      "5122",
			ExpectedNAICS:    "325412",
			ExpectedSIC:      "2834",
			ExpectedIndustry: "Healthcare",
			TestCategory:     "primary",
			Confidence:       0.98,
		},
		{
			BusinessName:     "UnitedHealth Group Inc.",
			Description:      "Diversified health care company offering health care products and insurance services",
			WebsiteURL:       "https://unitedhealthgroup.com",
			ExpectedMCC:      "6300",
			ExpectedNAICS:    "524114",
			ExpectedSIC:      "6321",
			ExpectedIndustry: "Healthcare",
			TestCategory:     "primary",
			Confidence:       0.96,
		},

		// Retail
		{
			BusinessName:     "Walmart Inc.",
			Description:      "Multinational retail corporation that operates a chain of hypermarkets, discount department stores, and grocery stores",
			WebsiteURL:       "https://walmart.com",
			ExpectedMCC:      "5310",
			ExpectedNAICS:    "452111",
			ExpectedSIC:      "5310",
			ExpectedIndustry: "Retail",
			TestCategory:     "primary",
			Confidence:       0.99,
		},
		{
			BusinessName:     "Target Corporation",
			Description:      "General merchandise retailer offering clothing, electronics, home goods, and groceries",
			WebsiteURL:       "https://target.com",
			ExpectedMCC:      "5310",
			ExpectedNAICS:    "452111",
			ExpectedSIC:      "5310",
			ExpectedIndustry: "Retail",
			TestCategory:     "primary",
			Confidence:       0.97,
		},

		// Manufacturing
		{
			BusinessName:     "Ford Motor Company",
			Description:      "Multinational automobile manufacturer that designs, manufactures, markets, and services a full line of vehicles",
			WebsiteURL:       "https://ford.com",
			ExpectedMCC:      "5511",
			ExpectedNAICS:    "336111",
			ExpectedSIC:      "3711",
			ExpectedIndustry: "Manufacturing",
			TestCategory:     "primary",
			Confidence:       0.98,
		},
		{
			BusinessName:     "Boeing Company",
			Description:      "Multinational corporation that designs, manufactures, and sells airplanes, rotorcraft, rockets, satellites, and telecommunications equipment",
			WebsiteURL:       "https://boeing.com",
			ExpectedMCC:      "3720",
			ExpectedNAICS:    "336411",
			ExpectedSIC:      "3721",
			ExpectedIndustry: "Manufacturing",
			TestCategory:     "primary",
			Confidence:       0.97,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// generateEdgeCaseSamples generates samples for edge cases
func (tdg *TestDataGenerator) generateEdgeCaseSamples(ctx context.Context) error {
	samples := []TestSampleData{
		// Ambiguous business names
		{
			BusinessName:     "ABC Corporation",
			Description:      "General business services and consulting",
			WebsiteURL:       "https://abccorp.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541611",
			ExpectedSIC:      "8742",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "edge_case",
			Confidence:       0.70,
		},
		{
			BusinessName:     "Global Solutions Inc.",
			Description:      "International business solutions and services",
			WebsiteURL:       "https://globalsolutions.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541611",
			ExpectedSIC:      "8742",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "edge_case",
			Confidence:       0.65,
		},

		// Multi-industry businesses
		{
			BusinessName:     "TechMed Solutions",
			Description:      "Technology solutions for healthcare industry including software, hardware, and consulting services",
			WebsiteURL:       "https://techmedsolutions.com",
			ExpectedMCC:      "5733",
			ExpectedNAICS:    "541511",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "edge_case",
			Confidence:       0.75,
		},
		{
			BusinessName:     "FinTech Innovations",
			Description:      "Financial technology company providing digital banking and payment solutions",
			WebsiteURL:       "https://fintechinnovations.com",
			ExpectedMCC:      "6012",
			ExpectedNAICS:    "523110",
			ExpectedSIC:      "6211",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "edge_case",
			Confidence:       0.80,
		},

		// Unusual business models
		{
			BusinessName:     "Subscription Box Co.",
			Description:      "Monthly subscription service delivering curated products to customers",
			WebsiteURL:       "https://subscriptionbox.com",
			ExpectedMCC:      "5969",
			ExpectedNAICS:    "454110",
			ExpectedSIC:      "5961",
			ExpectedIndustry: "Retail",
			TestCategory:     "edge_case",
			Confidence:       0.85,
		},
		{
			BusinessName:     "Gig Economy Platform",
			Description:      "Online platform connecting freelancers with clients for various services",
			WebsiteURL:       "https://gigeconomy.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "518210",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "edge_case",
			Confidence:       0.75,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// generateHighRiskSamples generates samples for high-risk industries
func (tdg *TestDataGenerator) generateHighRiskSamples(ctx context.Context) error {
	samples := []TestSampleData{
		// Cryptocurrency
		{
			BusinessName:     "Crypto Exchange Pro",
			Description:      "Digital currency exchange platform for buying and selling cryptocurrencies",
			WebsiteURL:       "https://cryptoexchangepro.com",
			ExpectedMCC:      "5999",
			ExpectedNAICS:    "523130",
			ExpectedSIC:      "7389",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "high_risk",
			Confidence:       0.85,
		},
		{
			BusinessName:     "Bitcoin Mining Corp",
			Description:      "Cryptocurrency mining operations and blockchain technology services",
			WebsiteURL:       "https://bitcoinmining.com",
			ExpectedMCC:      "5999",
			ExpectedNAICS:    "523130",
			ExpectedSIC:      "7389",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "high_risk",
			Confidence:       0.80,
		},

		// Adult Entertainment
		{
			BusinessName:     "Adult Entertainment Network",
			Description:      "Adult entertainment and media services",
			WebsiteURL:       "https://adultnetwork.com",
			ExpectedMCC:      "7273",
			ExpectedNAICS:    "713290",
			ExpectedSIC:      "7999",
			ExpectedIndustry: "Adult Entertainment",
			TestCategory:     "high_risk",
			Confidence:       0.90,
		},

		// Gambling
		{
			BusinessName:     "Online Casino Platform",
			Description:      "Online gambling and casino gaming platform",
			WebsiteURL:       "https://onlinecasino.com",
			ExpectedMCC:      "7995",
			ExpectedNAICS:    "713290",
			ExpectedSIC:      "7999",
			ExpectedIndustry: "Gambling",
			TestCategory:     "high_risk",
			Confidence:       0.95,
		},

		// Money Services
		{
			BusinessName:     "Quick Cash Services",
			Description:      "Money transfer and check cashing services",
			WebsiteURL:       "https://quickcash.com",
			ExpectedMCC:      "6012",
			ExpectedNAICS:    "522320",
			ExpectedSIC:      "6099",
			ExpectedIndustry: "Financial Services",
			TestCategory:     "high_risk",
			Confidence:       0.88,
		},

		// Pharmaceuticals
		{
			BusinessName:     "Online Pharmacy Direct",
			Description:      "Online pharmacy and prescription drug services",
			WebsiteURL:       "https://onlinepharmacy.com",
			ExpectedMCC:      "5122",
			ExpectedNAICS:    "446191",
			ExpectedSIC:      "5912",
			ExpectedIndustry: "Healthcare",
			TestCategory:     "high_risk",
			Confidence:       0.85,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// generateEmergingIndustrySamples generates samples for emerging industries
func (tdg *TestDataGenerator) generateEmergingIndustrySamples(ctx context.Context) error {
	samples := []TestSampleData{
		// AI/ML
		{
			BusinessName:     "AI Solutions Inc.",
			Description:      "Artificial intelligence and machine learning solutions for businesses",
			WebsiteURL:       "https://aisolutions.com",
			ExpectedMCC:      "5733",
			ExpectedNAICS:    "541511",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "emerging",
			Confidence:       0.80,
		},
		{
			BusinessName:     "Machine Learning Labs",
			Description:      "Machine learning research and development company",
			WebsiteURL:       "https://mllabs.com",
			ExpectedMCC:      "5733",
			ExpectedNAICS:    "541511",
			ExpectedSIC:      "7372",
			ExpectedIndustry: "Technology",
			TestCategory:     "emerging",
			Confidence:       0.75,
		},

		// Green Energy
		{
			BusinessName:     "Solar Power Systems",
			Description:      "Solar panel installation and renewable energy solutions",
			WebsiteURL:       "https://solarpower.com",
			ExpectedMCC:      "1711",
			ExpectedNAICS:    "238220",
			ExpectedSIC:      "1629",
			ExpectedIndustry: "Green Energy",
			TestCategory:     "emerging",
			Confidence:       0.85,
		},
		{
			BusinessName:     "Wind Energy Corp",
			Description:      "Wind turbine manufacturing and renewable energy development",
			WebsiteURL:       "https://windenergy.com",
			ExpectedMCC:      "3720",
			ExpectedNAICS:    "333611",
			ExpectedSIC:      "3621",
			ExpectedIndustry: "Green Energy",
			TestCategory:     "emerging",
			Confidence:       0.80,
		},

		// E-commerce
		{
			BusinessName:     "Direct-to-Consumer Brand",
			Description:      "Online-only consumer goods brand selling directly to customers",
			WebsiteURL:       "https://dtcbrand.com",
			ExpectedMCC:      "5969",
			ExpectedNAICS:    "454110",
			ExpectedSIC:      "5961",
			ExpectedIndustry: "Retail",
			TestCategory:     "emerging",
			Confidence:       0.85,
		},

		// Space Technology
		{
			BusinessName:     "Space Technology Ventures",
			Description:      "Commercial space technology and satellite services",
			WebsiteURL:       "https://spacetech.com",
			ExpectedMCC:      "3720",
			ExpectedNAICS:    "336414",
			ExpectedSIC:      "3721",
			ExpectedIndustry: "Technology",
			TestCategory:     "emerging",
			Confidence:       0.70,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// generateCrosswalkValidationSamples generates samples for crosswalk validation
func (tdg *TestDataGenerator) generateCrosswalkValidationSamples(ctx context.Context) error {
	samples := []TestSampleData{
		// Restaurants
		{
			BusinessName:     "Fine Dining Restaurant",
			Description:      "Upscale restaurant offering fine dining experiences",
			WebsiteURL:       "https://finedining.com",
			ExpectedMCC:      "5812",
			ExpectedNAICS:    "722511",
			ExpectedSIC:      "5812",
			ExpectedIndustry: "Food Service",
			TestCategory:     "crosswalk",
			Confidence:       0.95,
		},
		{
			BusinessName:     "Fast Food Chain",
			Description:      "Quick service restaurant chain",
			WebsiteURL:       "https://fastfood.com",
			ExpectedMCC:      "5812",
			ExpectedNAICS:    "722513",
			ExpectedSIC:      "5812",
			ExpectedIndustry: "Food Service",
			TestCategory:     "crosswalk",
			Confidence:       0.98,
		},

		// Retail Stores
		{
			BusinessName:     "Department Store",
			Description:      "Large retail store selling a wide variety of merchandise",
			WebsiteURL:       "https://departmentstore.com",
			ExpectedMCC:      "5310",
			ExpectedNAICS:    "452111",
			ExpectedSIC:      "5311",
			ExpectedIndustry: "Retail",
			TestCategory:     "crosswalk",
			Confidence:       0.97,
		},
		{
			BusinessName:     "Specialty Retail Store",
			Description:      "Retail store specializing in specific product categories",
			WebsiteURL:       "https://specialtyretail.com",
			ExpectedMCC:      "5999",
			ExpectedNAICS:    "453220",
			ExpectedSIC:      "5999",
			ExpectedIndustry: "Retail",
			TestCategory:     "crosswalk",
			Confidence:       0.90,
		},

		// Professional Services
		{
			BusinessName:     "Law Firm",
			Description:      "Legal services and law practice",
			WebsiteURL:       "https://lawfirm.com",
			ExpectedMCC:      "8111",
			ExpectedNAICS:    "541110",
			ExpectedSIC:      "8111",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "crosswalk",
			Confidence:       0.98,
		},
		{
			BusinessName:     "Accounting Firm",
			Description:      "Accounting and tax preparation services",
			WebsiteURL:       "https://accountingfirm.com",
			ExpectedMCC:      "8931",
			ExpectedNAICS:    "541211",
			ExpectedSIC:      "8721",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "crosswalk",
			Confidence:       0.97,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// generateConfidenceValidationSamples generates samples for confidence validation
func (tdg *TestDataGenerator) generateConfidenceValidationSamples(ctx context.Context) error {
	samples := []TestSampleData{
		// High confidence cases
		{
			BusinessName:     "McDonald's Corporation",
			Description:      "Fast food restaurant chain",
			WebsiteURL:       "https://mcdonalds.com",
			ExpectedMCC:      "5812",
			ExpectedNAICS:    "722513",
			ExpectedSIC:      "5812",
			ExpectedIndustry: "Food Service",
			TestCategory:     "confidence",
			Confidence:       0.99,
		},
		{
			BusinessName:     "Starbucks Corporation",
			Description:      "Coffeehouse chain and roastery",
			WebsiteURL:       "https://starbucks.com",
			ExpectedMCC:      "5812",
			ExpectedNAICS:    "722513",
			ExpectedSIC:      "5812",
			ExpectedIndustry: "Food Service",
			TestCategory:     "confidence",
			Confidence:       0.98,
		},

		// Medium confidence cases
		{
			BusinessName:     "Consulting Services LLC",
			Description:      "Business consulting and advisory services",
			WebsiteURL:       "https://consultingservices.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541611",
			ExpectedSIC:      "8742",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "confidence",
			Confidence:       0.75,
		},
		{
			BusinessName:     "Innovation Labs",
			Description:      "Research and development laboratory",
			WebsiteURL:       "https://innovationlabs.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541711",
			ExpectedSIC:      "8731",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "confidence",
			Confidence:       0.70,
		},

		// Low confidence cases
		{
			BusinessName:     "Global Enterprises",
			Description:      "International business operations",
			WebsiteURL:       "https://globalenterprises.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541611",
			ExpectedSIC:      "8742",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "confidence",
			Confidence:       0.50,
		},
		{
			BusinessName:     "Multi-Services Inc.",
			Description:      "Various business services and solutions",
			WebsiteURL:       "https://multiservices.com",
			ExpectedMCC:      "7399",
			ExpectedNAICS:    "541611",
			ExpectedSIC:      "8742",
			ExpectedIndustry: "Professional Services",
			TestCategory:     "confidence",
			Confidence:       0.45,
		},
	}

	return tdg.insertTestSamples(ctx, samples)
}

// TestSampleData represents test sample data for insertion
type TestSampleData struct {
	BusinessName     string
	Description      string
	WebsiteURL       string
	ExpectedMCC      string
	ExpectedNAICS    string
	ExpectedSIC      string
	ExpectedIndustry string
	TestCategory     string
	Confidence       float64
}

// insertTestSamples inserts test samples into the database
func (tdg *TestDataGenerator) insertTestSamples(ctx context.Context, samples []TestSampleData) error {
	query := `
		INSERT INTO classification_test_samples (
			business_name, description, website_url, expected_mcc, expected_naics, 
			expected_sic, expected_industry, manual_classification, test_category, 
			created_by, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (business_name) DO NOTHING
	`

	for _, sample := range samples {
		// Create manual classification JSON
		manualClassification := ManualClassification{
			MCCCode:          sample.ExpectedMCC,
			MCCDescription:   tdg.getMCCDescription(sample.ExpectedMCC),
			NAICSCode:        sample.ExpectedNAICS,
			NAICSDescription: tdg.getNAICSDescription(sample.ExpectedNAICS),
			SICCode:          sample.ExpectedSIC,
			SICDescription:   tdg.getSICDescription(sample.ExpectedSIC),
			IndustryID:       tdg.getIndustryID(sample.ExpectedIndustry),
			IndustryName:     sample.ExpectedIndustry,
			Confidence:       sample.Confidence,
			Notes:            fmt.Sprintf("Generated test sample for %s category", sample.TestCategory),
			ClassifiedBy:     "test_data_generator",
			ClassifiedAt:     time.Now(),
		}

		manualClassificationJSON, err := json.Marshal(manualClassification)
		if err != nil {
			return fmt.Errorf("failed to marshal manual classification: %w", err)
		}

		_, err = tdg.db.ExecContext(ctx, query,
			sample.BusinessName,
			sample.Description,
			sample.WebsiteURL,
			sample.ExpectedMCC,
			sample.ExpectedNAICS,
			sample.ExpectedSIC,
			sample.ExpectedIndustry,
			string(manualClassificationJSON),
			sample.TestCategory,
			"test_data_generator",
			true,
		)

		if err != nil {
			return fmt.Errorf("failed to insert test sample %s: %w", sample.BusinessName, err)
		}
	}

	tdg.logger.Printf("Inserted %d test samples", len(samples))
	return nil
}

// Helper methods for getting descriptions and IDs
func (tdg *TestDataGenerator) getMCCDescription(code string) string {
	descriptions := map[string]string{
		"5733": "Computer Software Stores",
		"6012": "Financial Institutions - Merchandise, Services",
		"5122": "Drugs, Drug Proprietaries, and Druggist Sundries",
		"5310": "Department Stores",
		"5511": "Automotive Dealers (New & Used) Sales, Service, Repairs Parts and Leasing",
		"3720": "Aircraft - Airplanes and Helicopters",
		"7399": "Business Services, Not Elsewhere Classified",
		"5969": "Miscellaneous and Specialty Retail Stores",
		"5812": "Eating Places, Restaurants",
		"5999": "Miscellaneous and Specialty Retail Stores",
		"7273": "Dating Services",
		"7995": "Betting (Wagering) Shops",
		"6300": "Insurance Sales, Underwriting, and Premiums",
		"1711": "Air Conditioning Contractors - Sales and Installation",
		"8111": "Legal Services, Attorneys",
		"8931": "Accounting, Auditing, and Bookkeeping Services",
	}
	return descriptions[code]
}

func (tdg *TestDataGenerator) getNAICSDescription(code string) string {
	descriptions := map[string]string{
		"541511": "Custom Computer Programming Services",
		"522110": "Commercial Banking",
		"325412": "Pharmaceutical Preparation Manufacturing",
		"452111": "Department Stores",
		"336111": "Automobile Manufacturing",
		"541611": "Administrative Management and General Management Consulting Services",
		"518210": "Data Processing, Hosting, and Related Services",
		"523110": "Investment Banking and Securities Dealing",
		"523130": "Securities and Commodity Exchanges",
		"713290": "Other Gambling Industries",
		"522320": "Financial Transactions Processing, Reserve, and Clearinghouse Activities",
		"446191": "Food (Health) Supplement Stores",
		"238220": "Plumbing, Heating, and Air-Conditioning Contractors",
		"333611": "Turbine and Turbine Generator Set Units Manufacturing",
		"454110": "Electronic Shopping and Mail-Order Houses",
		"336414": "Guided Missile and Space Vehicle Manufacturing",
		"722511": "Full-Service Restaurants",
		"722513": "Limited-Service Restaurants",
		"453220": "Gift, Novelty, and Souvenir Stores",
		"541110": "Offices of Lawyers",
		"541211": "Offices of Certified Public Accountants",
		"541711": "Research and Development in the Physical, Engineering, and Life Sciences",
	}
	return descriptions[code]
}

func (tdg *TestDataGenerator) getSICDescription(code string) string {
	descriptions := map[string]string{
		"7372": "Prepackaged Software",
		"6021": "National Commercial Banks",
		"2834": "Pharmaceutical Preparations",
		"5310": "Department Stores",
		"3711": "Motor Vehicles and Passenger Car Bodies",
		"8742": "Management Consulting Services",
		"6211": "Security Brokers, Dealers, and Flotation Companies",
		"7389": "Business Services, Not Elsewhere Classified",
		"7999": "Amusement and Recreation Services, Not Elsewhere Classified",
		"6099": "Functions Related to Depository Banking, Not Elsewhere Classified",
		"5912": "Drug Stores and Proprietary Stores",
		"1629": "Heavy Construction, Not Elsewhere Classified",
		"3621": "Motors and Generators",
		"5961": "Catalog and Mail-Order Houses",
		"3721": "Aircraft",
		"5812": "Eating Places",
		"5311": "Department Stores",
		"5999": "Miscellaneous Retail Stores, Not Elsewhere Classified",
		"8111": "Legal Services",
		"8721": "Accounting, Auditing, and Bookkeeping Services",
		"8731": "Commercial Physical and Biological Research",
	}
	return descriptions[code]
}

func (tdg *TestDataGenerator) getIndustryID(industryName string) int {
	industryIDs := map[string]int{
		"Technology":            1,
		"Financial Services":    2,
		"Healthcare":            3,
		"Retail":                4,
		"Manufacturing":         5,
		"Adult Entertainment":   6,
		"Green Energy":          7,
		"Food Service":          8,
		"Professional Services": 9,
		"Gambling":              10,
	}
	return industryIDs[industryName]
}

// GenerateRandomTestSamples generates additional random test samples for stress testing
func (tdg *TestDataGenerator) GenerateRandomTestSamples(ctx context.Context, count int) error {
	tdg.logger.Printf("Generating %d random test samples for stress testing...", count)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Sample data pools
	businessNames := []string{
		"Alpha Corporation", "Beta Industries", "Gamma Solutions", "Delta Services",
		"Epsilon Technologies", "Zeta Enterprises", "Eta Consulting", "Theta Group",
		"Iota Systems", "Kappa Holdings", "Lambda Corp", "Mu Industries",
		"Nu Solutions", "Xi Services", "Omicron Technologies", "Pi Enterprises",
		"Rho Consulting", "Sigma Group", "Tau Systems", "Upsilon Holdings",
	}

	descriptions := []string{
		"Comprehensive business solutions and services",
		"Professional consulting and advisory services",
		"Technology solutions and software development",
		"Manufacturing and industrial services",
		"Retail and consumer products",
		"Financial services and investment solutions",
		"Healthcare and medical services",
		"Educational and training services",
		"Real estate and property services",
		"Transportation and logistics services",
	}

	industries := []string{"Technology", "Financial Services", "Healthcare", "Retail", "Manufacturing", "Professional Services"}
	mccCodes := []string{"5733", "6012", "5122", "5310", "5511", "7399", "5969", "5812"}
	naicsCodes := []string{"541511", "522110", "325412", "452111", "336111", "541611", "518210", "523110"}
	sicCodes := []string{"7372", "6021", "2834", "5310", "3711", "8742", "6211", "7389"}

	var samples []TestSampleData

	for i := 0; i < count; i++ {
		businessName := businessNames[rand.Intn(len(businessNames))]
		description := descriptions[rand.Intn(len(descriptions))]
		industry := industries[rand.Intn(len(industries))]
		mccCode := mccCodes[rand.Intn(len(mccCodes))]
		naicsCode := naicsCodes[rand.Intn(len(naicsCodes))]
		sicCode := sicCodes[rand.Intn(len(sicCodes))]

		// Add random suffix to make names unique
		businessName = fmt.Sprintf("%s %d", businessName, i+1)

		sample := TestSampleData{
			BusinessName:     businessName,
			Description:      description,
			WebsiteURL:       fmt.Sprintf("https://%s.com", businessName),
			ExpectedMCC:      mccCode,
			ExpectedNAICS:    naicsCode,
			ExpectedSIC:      sicCode,
			ExpectedIndustry: industry,
			TestCategory:     "random",
			Confidence:       0.5 + rand.Float64()*0.4, // Random confidence between 0.5 and 0.9
		}

		samples = append(samples, sample)
	}

	return tdg.insertTestSamples(ctx, samples)
}
