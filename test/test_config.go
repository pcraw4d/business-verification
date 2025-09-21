package test

import (
	"os"
	"strconv"
	"time"
)

// TestConfig holds configuration for test execution
type TestConfig struct {
	// Database configuration
	DatabaseURL string

	// Test timeouts
	DefaultTimeout     time.Duration
	LongTestTimeout    time.Duration
	PerformanceTimeout time.Duration

	// Test data configuration
	LargeContentSize    int
	ConcurrentTestCount int

	// Performance thresholds
	MaxRiskDetectionTime  time.Duration
	MaxCrosswalkQueryTime time.Duration
	MaxRiskAssessmentTime time.Duration
	MaxConcurrentTestTime time.Duration

	// Test data paths
	TestDataDir string

	// Logging configuration
	VerboseLogging bool
	LogLevel       string
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL:           getEnvOrDefault("DATABASE_URL", ""),
		DefaultTimeout:        getDurationEnvOrDefault("TEST_DEFAULT_TIMEOUT", 30*time.Second),
		LongTestTimeout:       getDurationEnvOrDefault("TEST_LONG_TIMEOUT", 60*time.Second),
		PerformanceTimeout:    getDurationEnvOrDefault("TEST_PERFORMANCE_TIMEOUT", 120*time.Second),
		LargeContentSize:      getIntEnvOrDefault("TEST_LARGE_CONTENT_SIZE", 10000),
		ConcurrentTestCount:   getIntEnvOrDefault("TEST_CONCURRENT_COUNT", 10),
		MaxRiskDetectionTime:  getDurationEnvOrDefault("TEST_MAX_RISK_DETECTION_TIME", 1*time.Second),
		MaxCrosswalkQueryTime: getDurationEnvOrDefault("TEST_MAX_CROSSWALK_TIME", 2*time.Second),
		MaxRiskAssessmentTime: getDurationEnvOrDefault("TEST_MAX_RISK_ASSESSMENT_TIME", 5*time.Second),
		MaxConcurrentTestTime: getDurationEnvOrDefault("TEST_MAX_CONCURRENT_TIME", 15*time.Second),
		TestDataDir:           getEnvOrDefault("TEST_DATA_DIR", "testdata"),
		VerboseLogging:        getBoolEnvOrDefault("TEST_VERBOSE_LOGGING", false),
		LogLevel:              getEnvOrDefault("TEST_LOG_LEVEL", "info"),
	}
}

// TestData provides test data for various scenarios
type TestData struct {
	// High-risk business test data
	HighRiskBusinesses []BusinessTestData

	// Low-risk business test data
	LowRiskBusinesses []BusinessTestData

	// Medium-risk business test data
	MediumRiskBusinesses []BusinessTestData

	// Risk keywords test data
	RiskKeywords []RiskKeywordTestData

	// Crosswalk test data
	CrosswalkMappings []CrosswalkTestData
}

// BusinessTestData represents test data for business assessment
type BusinessTestData struct {
	Name              string
	WebsiteURL        string
	DomainName        string
	Industry          string
	BusinessType      string
	Description       string
	ExpectedRiskLevel string
	ExpectedRiskScore float64
}

// RiskKeywordTestData represents test data for risk keyword detection
type RiskKeywordTestData struct {
	Content            string
	Source             string
	ExpectedKeywords   []string
	ExpectedRiskLevel  string
	ExpectedRiskScore  float64
	ExpectedCategories []string
}

// CrosswalkTestData represents test data for crosswalk validation
type CrosswalkTestData struct {
	SourceCode         string
	SourceSystem       string
	TargetCode         string
	TargetSystem       string
	ExpectedMapping    bool
	ExpectedConfidence float64
}

// GetTestData returns comprehensive test data
func GetTestData() *TestData {
	return &TestData{
		HighRiskBusinesses: []BusinessTestData{
			{
				Name:              "Online Casino & Gambling Platform",
				WebsiteURL:        "https://example-casino.com",
				DomainName:        "example-casino.com",
				Industry:          "Gambling",
				BusinessType:      "Online Gaming",
				Description:       "Online casino offering poker, blackjack, and slot games",
				ExpectedRiskLevel: "high",
				ExpectedRiskScore: 0.8,
			},
			{
				Name:              "Adult Entertainment Services",
				WebsiteURL:        "https://adult-entertainment.com",
				DomainName:        "adult-entertainment.com",
				Industry:          "Adult Entertainment",
				BusinessType:      "Adult Services",
				Description:       "Adult entertainment and escort services",
				ExpectedRiskLevel: "high",
				ExpectedRiskScore: 0.9,
			},
			{
				Name:              "Cryptocurrency Exchange",
				WebsiteURL:        "https://crypto-exchange.com",
				DomainName:        "crypto-exchange.com",
				Industry:          "Financial Services",
				BusinessType:      "Cryptocurrency Exchange",
				Description:       "Digital currency exchange platform",
				ExpectedRiskLevel: "medium",
				ExpectedRiskScore: 0.6,
			},
		},

		LowRiskBusinesses: []BusinessTestData{
			{
				Name:              "Family Restaurant & Catering",
				WebsiteURL:        "https://family-restaurant.com",
				DomainName:        "family-restaurant.com",
				Industry:          "Food Service",
				BusinessType:      "Restaurant",
				Description:       "Family-owned restaurant serving local cuisine",
				ExpectedRiskLevel: "low",
				ExpectedRiskScore: 0.1,
			},
			{
				Name:              "Technology Consulting Services",
				WebsiteURL:        "https://tech-consulting.com",
				DomainName:        "tech-consulting.com",
				Industry:          "Technology",
				BusinessType:      "Consulting",
				Description:       "IT consulting and software development services",
				ExpectedRiskLevel: "low",
				ExpectedRiskScore: 0.2,
			},
			{
				Name:              "Local Bookstore",
				WebsiteURL:        "https://local-bookstore.com",
				DomainName:        "local-bookstore.com",
				Industry:          "Retail",
				BusinessType:      "Bookstore",
				Description:       "Independent bookstore selling books and educational materials",
				ExpectedRiskLevel: "low",
				ExpectedRiskScore: 0.1,
			},
		},

		MediumRiskBusinesses: []BusinessTestData{
			{
				Name:              "Pharmaceutical Distributor",
				WebsiteURL:        "https://pharma-distributor.com",
				DomainName:        "pharma-distributor.com",
				Industry:          "Healthcare",
				BusinessType:      "Pharmaceutical Distribution",
				Description:       "Licensed pharmaceutical distribution services",
				ExpectedRiskLevel: "medium",
				ExpectedRiskScore: 0.4,
			},
			{
				Name:              "Money Transfer Services",
				WebsiteURL:        "https://money-transfer.com",
				DomainName:        "money-transfer.com",
				Industry:          "Financial Services",
				BusinessType:      "Money Transfer",
				Description:       "International money transfer and remittance services",
				ExpectedRiskLevel: "medium",
				ExpectedRiskScore: 0.5,
			},
		},

		RiskKeywords: []RiskKeywordTestData{
			{
				Content:            "Welcome to our online casino and gambling platform. We offer the best poker, blackjack, and slot games.",
				Source:             "website",
				ExpectedKeywords:   []string{"casino", "gambling", "poker", "blackjack", "slot games"},
				ExpectedRiskLevel:  "high",
				ExpectedRiskScore:  0.8,
				ExpectedCategories: []string{"prohibited", "high_risk"},
			},
			{
				Content:            "We provide digital currency exchange services. Our platform supports Bitcoin, Ethereum, and other cryptocurrencies.",
				Source:             "website",
				ExpectedKeywords:   []string{"digital currency", "cryptocurrency", "Bitcoin", "Ethereum"},
				ExpectedRiskLevel:  "medium",
				ExpectedRiskScore:  0.6,
				ExpectedCategories: []string{"high_risk"},
			},
			{
				Content:            "Welcome to our family restaurant. We serve delicious meals and provide excellent customer service.",
				Source:             "website",
				ExpectedKeywords:   []string{},
				ExpectedRiskLevel:  "low",
				ExpectedRiskScore:  0.1,
				ExpectedCategories: []string{},
			},
		},

		CrosswalkMappings: []CrosswalkTestData{
			{
				SourceCode:         "5734",
				SourceSystem:       "MCC",
				TargetCode:         "1",
				TargetSystem:       "INDUSTRY",
				ExpectedMapping:    true,
				ExpectedConfidence: 0.8,
			},
			{
				SourceCode:         "541511",
				SourceSystem:       "NAICS",
				TargetCode:         "1",
				TargetSystem:       "INDUSTRY",
				ExpectedMapping:    true,
				ExpectedConfidence: 0.9,
			},
			{
				SourceCode:         "7372",
				SourceSystem:       "SIC",
				TargetCode:         "1",
				TargetSystem:       "INDUSTRY",
				ExpectedMapping:    true,
				ExpectedConfidence: 0.85,
			},
		},
	}
}

// Helper functions for environment variable handling

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Test utilities

// AssertRiskLevel asserts that the actual risk level matches expected
func AssertRiskLevel(actual, expected string) bool {
	return actual == expected
}

// AssertRiskScore asserts that the actual risk score is within acceptable range
func AssertRiskScore(actual, expected float64, tolerance float64) bool {
	diff := actual - expected
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// AssertKeywordsDetected asserts that expected keywords are detected
func AssertKeywordsDetected(actualKeywords []string, expectedKeywords []string) bool {
	if len(expectedKeywords) == 0 {
		return len(actualKeywords) == 0
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, actual := range actualKeywords {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// AssertCategoriesDetected asserts that expected categories are detected
func AssertCategoriesDetected(actualCategories []string, expectedCategories []string) bool {
	if len(expectedCategories) == 0 {
		return len(actualCategories) == 0
	}

	for _, expected := range expectedCategories {
		found := false
		for _, actual := range actualCategories {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// AssertPerformanceRequirement asserts that performance meets requirements
func AssertPerformanceRequirement(actualDuration, maxDuration time.Duration) bool {
	return actualDuration <= maxDuration
}

// GenerateTestContent generates test content of specified size
func GenerateTestContent(size int) string {
	baseContent := "This is a legitimate technology company providing software solutions. "
	content := ""
	for len(content) < size {
		content += baseContent
	}
	return content[:size]
}

// GenerateRiskContent generates content with risk keywords
func GenerateRiskContent(riskKeywords []string) string {
	content := "Welcome to our business. "
	for _, keyword := range riskKeywords {
		content += "We provide " + keyword + " services. "
	}
	content += "Contact us for more information."
	return content
}
