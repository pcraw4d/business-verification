//go:build e2e_railway
// +build e2e_railway

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRailwayComprehensiveE2EClassification runs comprehensive end-to-end tests
// against Railway production environment covering:
// - Web scraping and crawling strategies
// - Classification accuracy
// - Code and explanation generation
// - Performance and reliability analysis
func TestRailwayComprehensiveE2EClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive Railway E2E test in short mode")
	}

	// Get Railway API URL
	apiURL := os.Getenv("RAILWAY_API_URL")
	if apiURL == "" {
		apiURL = "https://classification-service-production.up.railway.app"
		t.Logf("Using default Railway URL: %s", apiURL)
	} else {
		t.Logf("Using Railway URL from environment: %s", apiURL)
	}

	// Verify service is accessible
	if !verifyServiceHealth(t, apiURL) {
		t.Fatalf("Service at %s is not accessible", apiURL)
	}

	// Load diverse test samples covering different industries, complexities, and scraping scenarios
	testSamples := generateComprehensiveTestSamples()
	t.Logf("🚀 Starting comprehensive Railway E2E tests with %d samples", len(testSamples))

	// Initialize test runner
	runner := NewRailwayE2ETestRunner(t, apiURL)

	// Run comprehensive tests
	startTime := time.Now()
	results := runner.RunComprehensiveTests(testSamples)
	totalDuration := time.Since(startTime)

	t.Logf("✅ Completed all tests in %v", totalDuration)

	// Calculate comprehensive metrics
	runner.CalculateMetrics()

	// Generate detailed analysis report
	report := runner.GenerateComprehensiveReport(totalDuration)

	// Analyze strengths, weaknesses, and opportunities
	analysis := runner.AnalyzeClassificationProcess(report)

	// Validate results
	validateE2EResults(t, report)

	// Save reports
	timestamp := time.Now().Format("20060102_150405")
	reportPath := fmt.Sprintf("test/results/railway_e2e_classification_%s.json", timestamp)
	analysisPath := fmt.Sprintf("test/results/railway_e2e_analysis_%s.json", timestamp)

	if err := saveReport(report, reportPath); err != nil {
		t.Errorf("Failed to save report: %v", err)
	} else {
		t.Logf("📊 Test report saved to %s", reportPath)
	}

	if err := saveReport(analysis, analysisPath); err != nil {
		t.Errorf("Failed to save analysis: %v", err)
	} else {
		t.Logf("📊 Analysis report saved to %s", analysisPath)
	}

	// Print comprehensive summary
	runner.PrintComprehensiveSummary()
}

// verifyServiceHealth checks if the Railway service is accessible
func verifyServiceHealth(t *testing.T, apiURL string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	
	healthURL := strings.TrimSuffix(apiURL, "/") + "/health"
	resp, err := client.Get(healthURL)
	if err != nil {
		t.Logf("⚠️ Health check failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Logf("✅ Service health check passed")
		return true
	}

	t.Logf("⚠️ Health check returned status: %d", resp.StatusCode)
	return false
}

// RailwayE2ETestRunner runs comprehensive Railway E2E tests
type RailwayE2ETestRunner struct {
	t          *testing.T
	apiURL     string
	httpClient *http.Client
	results    []RailwayE2ETestResult
	metrics    *RailwayE2EMetrics
	mu         sync.Mutex
}

// RailwayE2ETestResult extends ClassificationTestResult with additional E2E metrics
type RailwayE2ETestResult struct {
	ClassificationTestResult // Embed existing result type
	// Additional E2E-specific fields
	TestCategory          string                 `json:"test_category"`
	ScrapingDifficulty    string                 `json:"scraping_difficulty"`
	PagesCrawled          int                    `json:"pages_crawled"`
	PagesAnalyzed         int                    `json:"pages_analyzed"`
	CrawlingSuccess       bool                   `json:"crawling_success"`
	CrawlingErrors        []string               `json:"crawling_errors,omitempty"`
	RobotsTxtRespected    bool                   `json:"robots_txt_respected"`
	StructuredDataFound   bool                   `json:"structured_data_found"`
	StructuredDataType    string                 `json:"structured_data_type,omitempty"`
	Top3CodesGenerated    bool                   `json:"top3_codes_generated"`
	CodeConfidenceAvg     float64                `json:"code_confidence_avg"`
	CodeDescriptionsValid bool                   `json:"code_descriptions_valid"`
	ExplanationQuality    float64                `json:"explanation_quality"`
	ExplanationLength     int                    `json:"explanation_length"`
	ExplanationKeywords   []string               `json:"explanation_keywords,omitempty"`
	ExpectedIndustry      string                 `json:"expected_industry,omitempty"`
	IndustryMatch         bool                   `json:"industry_match,omitempty"`
	ExpectedMCC           []string               `json:"expected_mcc,omitempty"`
	ExpectedNAICS         []string               `json:"expected_naics,omitempty"`
	ExpectedSIC           []string               `json:"expected_sic,omitempty"`
	
	// Enhanced code accuracy fields
	MCCMatch              bool    `json:"mcc_match,omitempty"`              // Backward compatibility
	NAICSMatch            bool    `json:"naics_match,omitempty"`            // Backward compatibility
	SICMatch              bool    `json:"sic_match,omitempty"`              // Backward compatibility
	MCCTop1Match          bool    `json:"mcc_top1_match,omitempty"`         // Top 1 position match
	MCCTop3Match          bool    `json:"mcc_top3_match,omitempty"`         // Top 3 position match
	NAICSTop1Match        bool    `json:"naics_top1_match,omitempty"`
	NAICSTop3Match        bool    `json:"naics_top3_match,omitempty"`
	SICTop1Match          bool    `json:"sic_top1_match,omitempty"`
	SICTop3Match          bool    `json:"sic_top3_match,omitempty"`
	MCCAccuracyScore      float64 `json:"mcc_accuracy_score"`              // Rank-based score (0.0-1.0)
	NAICSAccuracyScore    float64 `json:"naics_accuracy_score"`
	SICAccuracyScore      float64 `json:"sic_accuracy_score"`
	MCCMatchedRank        int     `json:"mcc_matched_rank,omitempty"`        // Rank where code was found (0 = not found)
	NAICSMatchedRank      int     `json:"naics_matched_rank,omitempty"`
	SICMatchedRank        int     `json:"sic_matched_rank,omitempty"`
	
	ErrorType             string                 `json:"error_type,omitempty"`
}

// EnhancedCodeAccuracyMetrics tracks comprehensive code accuracy metrics
type EnhancedCodeAccuracyMetrics struct {
	// Rank-based accuracy scores (0.0-1.0)
	MCCAccuracyScore   float64 `json:"mcc_accuracy_score"`
	NAICSAccuracyScore float64 `json:"naics_accuracy_score"`
	SICAccuracyScore   float64 `json:"sic_accuracy_score"`
	
	// Top 1 accuracy (exact match in first position)
	MCCTop1Accuracy    float64 `json:"mcc_top1_accuracy"`
	NAICSTop1Accuracy  float64 `json:"naics_top1_accuracy"`
	SICTop1Accuracy    float64 `json:"sic_top1_accuracy"`
	
	// Top 3 accuracy (code appears in top 3)
	MCCTop3Accuracy    float64 `json:"mcc_top3_accuracy"`
	NAICSTop3Accuracy  float64 `json:"naics_top3_accuracy"`
	SICTop3Accuracy    float64 `json:"sic_top3_accuracy"`
	
	// Overall code accuracy (weighted average)
	OverallCodeAccuracy float64 `json:"overall_code_accuracy"`
	
	// Code accuracy by industry
	CodeAccuracyByIndustry map[string]EnhancedCodeAccuracyMetrics `json:"code_accuracy_by_industry,omitempty"`
}

// RailwayE2EMetrics tracks comprehensive metrics
type RailwayE2EMetrics struct {
	TotalTests            int                       `json:"total_tests"`
	SuccessfulTests       int                       `json:"successful_tests"`
	FailedTests           int                       `json:"failed_tests"`
	
	// Scraping & Crawling Metrics
	ScrapingSuccessRate    float64                   `json:"scraping_success_rate"`
	AveragePagesCrawled  float64                   `json:"average_pages_crawled"`
	StrategyDistribution map[string]int            `json:"strategy_distribution"`
	StrategySuccessRate  map[string]float64        `json:"strategy_success_rate"`
	StructuredDataRate   float64                   `json:"structured_data_rate"`
	
	// Classification Metrics
	ClassificationAccuracy float64                  `json:"classification_accuracy"`
	AverageConfidence      float64                  `json:"average_confidence"`
	IndustryAccuracy       map[string]float64        `json:"industry_accuracy"`
	
	// Code Generation Metrics
	CodeGenerationRate    float64                  `json:"code_generation_rate"`
	Top3CodeRate          float64                  `json:"top3_code_rate"`
	CodeConfidenceAvg     float64                  `json:"code_confidence_avg"`
	
	// Enhanced Code Accuracy Metrics
	CodeAccuracy          EnhancedCodeAccuracyMetrics `json:"code_accuracy"`
	
	// Explanation Metrics
	ExplanationGenerationRate float64              `json:"explanation_generation_rate"`
	AverageExplanationQuality  float64             `json:"average_explanation_quality"`
	AverageExplanationLength   int                 `json:"average_explanation_length"`
	
	// Performance Metrics
	AverageLatency        time.Duration            `json:"average_latency_ms"`
	P50Latency            time.Duration             `json:"p50_latency_ms"`
	P95Latency            time.Duration             `json:"p95_latency_ms"`
	P99Latency            time.Duration             `json:"p99_latency_ms"`
	CacheHitRate          float64                   `json:"cache_hit_rate"`
	EarlyExitRate         float64                  `json:"early_exit_rate"`
	FallbackRate          float64                  `json:"fallback_rate"`
	
	// Error Analysis
	ErrorDistribution     map[string]int           `json:"error_distribution"`
	ErrorRate             float64                   `json:"error_rate"`
}

// NewRailwayE2ETestRunner creates a new Railway E2E test runner
func NewRailwayE2ETestRunner(t *testing.T, apiURL string) *RailwayE2ETestRunner {
	return &RailwayE2ETestRunner{
		t:          t,
		apiURL:     apiURL,
		httpClient: &http.Client{Timeout: 180 * time.Second}, // Extended timeout for Railway
		results:    make([]RailwayE2ETestResult, 0),
		metrics: &RailwayE2EMetrics{
			StrategyDistribution: make(map[string]int),
			StrategySuccessRate:  make(map[string]float64),
			IndustryAccuracy:     make(map[string]float64),
			ErrorDistribution:    make(map[string]int),
			CodeAccuracy: EnhancedCodeAccuracyMetrics{
				CodeAccuracyByIndustry: make(map[string]EnhancedCodeAccuracyMetrics),
			},
		},
	}
}

// generateComprehensiveTestSamples generates 385 diverse test samples for statistical confidence
func generateComprehensiveTestSamples() []TestSample {
	samples := make([]TestSample, 0, 385)
	
	// Real-world well-known businesses (~50 samples)
	realWorldSamples := []TestSample{
		// E-commerce & Retail
		{ID: "ecom_001", BusinessName: "Amazon", Description: "Online retail marketplace", WebsiteURL: "https://www.amazon.com", ExpectedIndustry: "retail", ExpectedMCC: []string{"5999", "5311", "5331"}, Category: "ecommerce", Complexity: "high", ScrapingDifficulty: "medium"},
		{ID: "ecom_002", BusinessName: "Shopify", Description: "E-commerce platform", WebsiteURL: "https://www.shopify.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"5734", "7372"}, Category: "saas", Complexity: "medium", ScrapingDifficulty: "low"},
		{ID: "ecom_003", BusinessName: "eBay", Description: "Online auction and marketplace", WebsiteURL: "https://www.ebay.com", ExpectedIndustry: "retail", ExpectedMCC: []string{"5999", "5311"}, Category: "ecommerce", Complexity: "high", ScrapingDifficulty: "medium"},
		{ID: "ecom_004", BusinessName: "Walmart", Description: "Retail corporation", WebsiteURL: "https://www.walmart.com", ExpectedIndustry: "retail", ExpectedMCC: []string{"5311", "5331"}, Category: "retail", Complexity: "high", ScrapingDifficulty: "medium"},
		{ID: "ecom_005", BusinessName: "Target", Description: "Retail store chain", WebsiteURL: "https://www.target.com", ExpectedIndustry: "retail", ExpectedMCC: []string{"5311", "5331"}, Category: "retail", Complexity: "medium", ScrapingDifficulty: "medium"},
		
		// Technology & Software
		{ID: "tech_001", BusinessName: "Microsoft", Description: "Technology corporation", WebsiteURL: "https://www.microsoft.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"5734", "7372"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_002", BusinessName: "Apple", Description: "Consumer electronics and software", WebsiteURL: "https://www.apple.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"5734", "7372"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_003", BusinessName: "Google", Description: "Internet search and cloud services", WebsiteURL: "https://www.google.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"7372", "5734"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_004", BusinessName: "Meta", Description: "Social media and technology company", WebsiteURL: "https://www.meta.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"7372", "5734"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_005", BusinessName: "Stripe", Description: "Payment processing platform", WebsiteURL: "https://stripe.com", ExpectedIndustry: "financial services", ExpectedMCC: []string{"5999", "6012"}, Category: "fintech", Complexity: "medium", ScrapingDifficulty: "low"},
		{ID: "tech_006", BusinessName: "Salesforce", Description: "Cloud-based CRM platform", WebsiteURL: "https://www.salesforce.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"7372", "5734"}, Category: "saas", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_007", BusinessName: "Oracle", Description: "Database and cloud services", WebsiteURL: "https://www.oracle.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"7372", "5734"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "tech_008", BusinessName: "IBM", Description: "Technology and consulting services", WebsiteURL: "https://www.ibm.com", ExpectedIndustry: "technology", ExpectedMCC: []string{"7372", "5734"}, Category: "technology", Complexity: "high", ScrapingDifficulty: "low"},
		
		// Food & Beverage
		{ID: "food_001", BusinessName: "Starbucks", Description: "Coffee chain", WebsiteURL: "https://www.starbucks.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5812", "5814"}, Category: "restaurant", Complexity: "medium", ScrapingDifficulty: "medium"},
		{ID: "food_002", BusinessName: "McDonald's", Description: "Fast food restaurant chain", WebsiteURL: "https://www.mcdonalds.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5814"}, Category: "restaurant", Complexity: "medium", ScrapingDifficulty: "medium"},
		{ID: "food_003", BusinessName: "Coca-Cola", Description: "Beverage company", WebsiteURL: "https://www.coca-cola.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5499", "5441"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "high"},
		{ID: "food_004", BusinessName: "PepsiCo", Description: "Food and beverage corporation", WebsiteURL: "https://www.pepsico.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5499", "5441"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "medium"},
		{ID: "food_005", BusinessName: "Domino's Pizza", Description: "Pizza delivery chain", WebsiteURL: "https://www.dominos.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5812", "5814"}, Category: "restaurant", Complexity: "medium", ScrapingDifficulty: "medium"},
		{ID: "food_006", BusinessName: "Subway", Description: "Sandwich restaurant chain", WebsiteURL: "https://www.subway.com", ExpectedIndustry: "food & beverage", ExpectedMCC: []string{"5812", "5814"}, Category: "restaurant", Complexity: "medium", ScrapingDifficulty: "medium"},
		
		// Healthcare
		{ID: "health_001", BusinessName: "UnitedHealth Group", Description: "Healthcare and insurance company", WebsiteURL: "https://www.unitedhealthgroup.com", ExpectedIndustry: "healthcare", ExpectedMCC: []string{"6300", "8011"}, Category: "healthcare", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "health_002", BusinessName: "CVS Health", Description: "Pharmacy and healthcare services", WebsiteURL: "https://www.cvshealth.com", ExpectedIndustry: "healthcare", ExpectedMCC: []string{"5912", "8011"}, Category: "healthcare", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "health_003", BusinessName: "Walgreens", Description: "Pharmacy and retail chain", WebsiteURL: "https://www.walgreens.com", ExpectedIndustry: "healthcare", ExpectedMCC: []string{"5912", "8011"}, Category: "healthcare", Complexity: "medium", ScrapingDifficulty: "medium"},
		{ID: "health_004", BusinessName: "Mayo Clinic", Description: "Medical center and hospital", WebsiteURL: "https://www.mayoclinic.org", ExpectedIndustry: "healthcare", ExpectedMCC: []string{"8011", "8062"}, Category: "healthcare", Complexity: "high", ScrapingDifficulty: "low"},
		
		// Financial Services
		{ID: "finance_001", BusinessName: "JPMorgan Chase", Description: "Banking and financial services", WebsiteURL: "https://www.jpmorganchase.com", ExpectedIndustry: "banking", ExpectedMCC: []string{"6011", "6012"}, Category: "banking", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "finance_002", BusinessName: "Bank of America", Description: "Banking and financial services", WebsiteURL: "https://www.bankofamerica.com", ExpectedIndustry: "banking", ExpectedMCC: []string{"6011", "6012"}, Category: "banking", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "finance_003", BusinessName: "Wells Fargo", Description: "Banking and financial services", WebsiteURL: "https://www.wellsfargo.com", ExpectedIndustry: "banking", ExpectedMCC: []string{"6011", "6012"}, Category: "banking", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "finance_004", BusinessName: "Goldman Sachs", Description: "Investment banking and securities", WebsiteURL: "https://www.goldmansachs.com", ExpectedIndustry: "financial services", ExpectedMCC: []string{"6012", "6211"}, Category: "banking", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "finance_005", BusinessName: "PayPal", Description: "Online payment system", WebsiteURL: "https://www.paypal.com", ExpectedIndustry: "financial services", ExpectedMCC: []string{"6012", "5999"}, Category: "fintech", Complexity: "medium", ScrapingDifficulty: "low"},
		
		// Manufacturing
		{ID: "mfg_001", BusinessName: "Tesla", Description: "Electric vehicle manufacturer", WebsiteURL: "https://www.tesla.com", ExpectedIndustry: "manufacturing", ExpectedMCC: []string{"5511", "5533"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "mfg_002", BusinessName: "Ford Motor Company", Description: "Automobile manufacturer", WebsiteURL: "https://www.ford.com", ExpectedIndustry: "manufacturing", ExpectedMCC: []string{"5511", "5533"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "mfg_003", BusinessName: "General Electric", Description: "Industrial manufacturing", WebsiteURL: "https://www.ge.com", ExpectedIndustry: "manufacturing", ExpectedMCC: []string{"5084", "5065"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "low"},
		
		// Entertainment
		{ID: "ent_001", BusinessName: "Netflix", Description: "Streaming entertainment service", WebsiteURL: "https://www.netflix.com", ExpectedIndustry: "arts & entertainment", ExpectedMCC: []string{"7829", "7832"}, Category: "entertainment", Complexity: "medium", ScrapingDifficulty: "medium"},
		{ID: "ent_002", BusinessName: "Disney", Description: "Entertainment and media company", WebsiteURL: "https://www.disney.com", ExpectedIndustry: "arts & entertainment", ExpectedMCC: []string{"7829", "7832"}, Category: "entertainment", Complexity: "high", ScrapingDifficulty: "medium"},
		{ID: "ent_003", BusinessName: "Spotify", Description: "Music streaming service", WebsiteURL: "https://www.spotify.com", ExpectedIndustry: "arts & entertainment", ExpectedMCC: []string{"5735", "7829"}, Category: "entertainment", Complexity: "medium", ScrapingDifficulty: "low"},
		
		// Professional Services
		{ID: "prof_001", BusinessName: "Deloitte", Description: "Professional services firm", WebsiteURL: "https://www2.deloitte.com", ExpectedIndustry: "professional services", ExpectedMCC: []string{"8931", "8999"}, Category: "consulting", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "prof_002", BusinessName: "PricewaterhouseCoopers", Description: "Professional services network", WebsiteURL: "https://www.pwc.com", ExpectedIndustry: "professional services", ExpectedMCC: []string{"8931", "8999"}, Category: "consulting", Complexity: "high", ScrapingDifficulty: "low"},
		{ID: "prof_003", BusinessName: "Ernst & Young", Description: "Professional services firm", WebsiteURL: "https://www.ey.com", ExpectedIndustry: "professional services", ExpectedMCC: []string{"8931", "8999"}, Category: "consulting", Complexity: "high", ScrapingDifficulty: "low"},
		
		// Construction
		{ID: "const_001", BusinessName: "Caterpillar", Description: "Construction and mining equipment", WebsiteURL: "https://www.caterpillar.com", ExpectedIndustry: "construction", ExpectedMCC: []string{"5082", "5083"}, Category: "manufacturing", Complexity: "high", ScrapingDifficulty: "low"},
		
		// Real Estate
		{ID: "real_001", BusinessName: "Zillow", Description: "Real estate marketplace", WebsiteURL: "https://www.zillow.com", ExpectedIndustry: "real estate", ExpectedMCC: []string{"6513", "6531"}, Category: "real estate", Complexity: "medium", ScrapingDifficulty: "medium"},
		
		// Transportation
		{ID: "trans_001", BusinessName: "Uber", Description: "Ride-sharing and transportation", WebsiteURL: "https://www.uber.com", ExpectedIndustry: "transportation", ExpectedMCC: []string{"4121", "4789"}, Category: "transportation", Complexity: "medium", ScrapingDifficulty: "low"},
		{ID: "trans_002", BusinessName: "FedEx", Description: "Shipping and logistics", WebsiteURL: "https://www.fedex.com", ExpectedIndustry: "transportation", ExpectedMCC: []string{"4214", "4215"}, Category: "transportation", Complexity: "high", ScrapingDifficulty: "medium"},
		
		// Energy
		{ID: "energy_001", BusinessName: "ExxonMobil", Description: "Oil and gas corporation", WebsiteURL: "https://www.exxonmobil.com", ExpectedIndustry: "energy", ExpectedMCC: []string{"5542", "5541"}, Category: "energy", Complexity: "high", ScrapingDifficulty: "low"},
	}
	
	samples = append(samples, realWorldSamples...)
	
	// Generate additional samples programmatically to reach 385 total
	// Industry distribution: ~30-40 samples per major industry
	industryTemplates := []struct {
		industry      string
		mccCodes      []string
		categories    []string
		complexities  []string
		scrapingDiffs []string
		count         int
	}{
		{"technology", []string{"5734", "7372"}, []string{"saas", "software", "tech"}, []string{"low", "medium", "high"}, []string{"low", "medium"}, 50},
		{"retail", []string{"5311", "5331", "5999"}, []string{"ecommerce", "retail", "marketplace"}, []string{"low", "medium", "high"}, []string{"low", "medium", "high"}, 45},
		{"food & beverage", []string{"5812", "5814", "5499"}, []string{"restaurant", "cafe", "bar"}, []string{"low", "medium"}, []string{"none", "low", "medium"}, 40},
		{"healthcare", []string{"8011", "5912", "6300"}, []string{"healthcare", "medical", "pharmacy"}, []string{"low", "medium", "high"}, []string{"low", "medium"}, 35},
		{"financial services", []string{"6011", "6012", "6211"}, []string{"banking", "fintech", "insurance"}, []string{"medium", "high"}, []string{"low", "medium"}, 35},
		{"manufacturing", []string{"5511", "5533", "5084"}, []string{"manufacturing", "industrial"}, []string{"medium", "high"}, []string{"low", "medium"}, 30},
		{"construction", []string{"1711", "1521", "1541"}, []string{"construction", "contractor"}, []string{"low", "medium"}, []string{"none", "low"}, 25},
		{"professional services", []string{"8931", "8999", "7392"}, []string{"consulting", "legal", "accounting"}, []string{"medium", "high"}, []string{"low", "medium"}, 30},
		{"arts & entertainment", []string{"7829", "7832", "5735"}, []string{"entertainment", "media", "streaming"}, []string{"low", "medium"}, []string{"low", "medium"}, 25},
		{"real estate", []string{"6513", "6531", "1521"}, []string{"real estate", "property"}, []string{"low", "medium"}, []string{"low", "medium"}, 20},
		{"transportation", []string{"4121", "4214", "4789"}, []string{"transportation", "logistics"}, []string{"low", "medium", "high"}, []string{"low", "medium"}, 20},
		{"education", []string{"8299", "8220", "8241"}, []string{"education", "training"}, []string{"low", "medium"}, []string{"low", "medium"}, 15},
		{"energy", []string{"5542", "5541", "5983"}, []string{"energy", "utilities"}, []string{"medium", "high"}, []string{"low", "medium"}, 15},
		{"agriculture", []string{"5999", "5261", "5193"}, []string{"agriculture", "farming"}, []string{"low", "medium"}, []string{"none", "low"}, 10},
	}
	
	businessNames := []string{
		"Global", "Premier", "Elite", "Advanced", "Pro", "Prime", "Superior", "Ultimate", "Premium",
		"National", "Regional", "Local", "Metro", "City", "State", "Coastal", "Mountain", "Valley",
		"Solutions", "Services", "Group", "Corp", "Inc", "LLC", "Partners", "Associates", "Enterprises",
		"Tech", "Digital", "Smart", "Innovative", "Modern", "Next", "Future", "NextGen", "ProTech",
	}
	
	businessTypes := []string{
		"Consulting", "Solutions", "Services", "Systems", "Technologies", "Industries", "Ventures",
		"Group", "Corporation", "Company", "Partners", "Associates", "Enterprises", "Holdings",
	}
	
	sampleID := len(realWorldSamples) + 1
	
	for _, template := range industryTemplates {
		for i := 0; i < template.count && len(samples) < 385; i++ {
			nameIdx := (sampleID + i) % len(businessNames)
			typeIdx := (sampleID + i) % len(businessTypes)
			categoryIdx := (sampleID + i) % len(template.categories)
			complexityIdx := (sampleID + i) % len(template.complexities)
			scrapingIdx := (sampleID + i) % len(template.scrapingDiffs)
			
			businessName := fmt.Sprintf("%s %s %s", businessNames[nameIdx], template.industry, businessTypes[typeIdx])
			if len(businessName) > 50 {
				businessName = businessName[:50]
			}
			
			description := fmt.Sprintf("%s %s", template.industry, template.categories[categoryIdx])
			if template.industry == "technology" {
				description = fmt.Sprintf("Software and %s services", template.categories[categoryIdx])
			} else if template.industry == "retail" {
				description = fmt.Sprintf("%s retail and e-commerce", template.categories[categoryIdx])
			} else if template.industry == "food & beverage" {
				description = fmt.Sprintf("%s restaurant and dining", template.categories[categoryIdx])
			}
			
			websiteURL := ""
			if template.scrapingDiffs[scrapingIdx] != "none" {
				// Generate realistic-looking URLs (but these won't actually exist)
				domain := strings.ToLower(strings.ReplaceAll(businessName, " ", ""))
				domain = strings.ReplaceAll(domain, "'", "")
				domain = strings.ReplaceAll(domain, ".", "")
				if len(domain) > 30 {
					domain = domain[:30]
				}
				websiteURL = fmt.Sprintf("https://www.%s.com", domain)
			}
			
			sample := TestSample{
				ID:                fmt.Sprintf("gen_%03d", sampleID),
				BusinessName:      businessName,
				Description:       description,
				WebsiteURL:        websiteURL,
				ExpectedIndustry:  template.industry,
				ExpectedMCC:       template.mccCodes,
				Category:          template.categories[categoryIdx],
				Complexity:        template.complexities[complexityIdx],
				ScrapingDifficulty: template.scrapingDiffs[scrapingIdx],
			}
			
			samples = append(samples, sample)
			sampleID++
		}
	}
	
	// Add small business samples without websites to reach exactly 385
	for len(samples) < 385 {
		smallBusinessTypes := []struct {
			name        string
			description string
			industry    string
			mcc         []string
		}{
			{"Joe's Pizza", "Local pizza restaurant", "food & beverage", []string{"5812"}},
			{"ABC Plumbing", "Residential plumbing contractor", "construction", []string{"1711"}},
			{"Main Street Bakery", "Local bakery and cafe", "food & beverage", []string{"5462"}},
			{"City Auto Repair", "Automotive repair shop", "automotive", []string{"7538"}},
			{"Green Lawn Care", "Landscaping services", "agriculture", []string{"0780"}},
			{"Quick Print Shop", "Printing and copying services", "professional services", []string{"5970"}},
			{"Corner Grocery", "Neighborhood grocery store", "retail", []string{"5411"}},
			{"Family Dentistry", "Dental practice", "healthcare", []string{"8021"}},
			{"Home Cleaning Co", "Residential cleaning services", "professional services", []string{"7349"}},
			{"Tech Support Plus", "Computer repair services", "technology", []string{"7379"}},
		}
		
		typeIdx := (len(samples) - len(realWorldSamples)) % len(smallBusinessTypes)
		businessType := smallBusinessTypes[typeIdx]
		
		sample := TestSample{
			ID:                fmt.Sprintf("small_%03d", len(samples)+1),
			BusinessName:      businessType.name,
			Description:       businessType.description,
			WebsiteURL:        "",
			ExpectedIndustry:  businessType.industry,
			ExpectedMCC:       businessType.mcc,
			Category:          "small business",
			Complexity:        "low",
			ScrapingDifficulty: "none",
		}
		
		samples = append(samples, sample)
	}
	
	// Ensure we have exactly 385 samples
	if len(samples) > 385 {
		samples = samples[:385]
	}
	
	return samples
}

// RunComprehensiveTests runs all test samples
func (r *RailwayE2ETestRunner) RunComprehensiveTests(samples []TestSample) []RailwayE2ETestResult {
	r.t.Logf("🚀 Starting comprehensive Railway E2E tests with %d samples", len(samples))

	// Run tests with controlled concurrency to avoid overwhelming Railway
	semaphore := make(chan struct{}, 3) // Max 3 concurrent requests
	var wg sync.WaitGroup

	for i, sample := range samples {
		wg.Add(1)
		go func(idx int, s TestSample) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			r.t.Logf("Running test %d/%d: %s", idx+1, len(samples), s.BusinessName)
			result := r.runSingleTest(s)
			r.addResult(result)

			// Small delay to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}(i, sample)
	}

	wg.Wait()

	r.t.Logf("✅ Completed all tests")
	return r.results
}

// runSingleTest runs a single comprehensive test
func (r *RailwayE2ETestRunner) runSingleTest(sample TestSample) RailwayE2ETestResult {
	result := RailwayE2ETestResult{
		ClassificationTestResult: ClassificationTestResult{
			SampleID:     sample.ID,
			BusinessName: sample.BusinessName,
			WebsiteURL:   sample.WebsiteURL,
			ExpectedIndustry: sample.ExpectedIndustry,
			Timestamp:    time.Now(),
		},
		TestCategory:       sample.Category,
		ScrapingDifficulty: sample.ScrapingDifficulty,
		ExpectedIndustry:   sample.ExpectedIndustry,
		ExpectedMCC:        sample.ExpectedMCC,
		ExpectedNAICS:      sample.ExpectedNAICS,
		ExpectedSIC:        sample.ExpectedSIC,
	}

	startTime := time.Now()

	// Make classification request
	reqBody := map[string]interface{}{
		"business_name": sample.BusinessName,
		"description":   sample.Description,
		"website_url":   sample.WebsiteURL,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal request: %v", err)
		result.ErrorType = "request_error"
		result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))
		return result
	}

	req, err := http.NewRequest("POST", r.apiURL+"/v1/classify", bytes.NewBuffer(reqJSON))
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.ErrorType = "request_error"
		result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		result.ErrorType = "network_error"
		result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))
		return result
	}
	defer resp.Body.Close()

	result.ProcessingTime = DurationMsFromDuration(time.Since(startTime))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		result.ErrorType = "response_error"
		return result
	}

	// Parse response
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		result.ErrorType = "parse_error"
		return result
	}

	// Extract comprehensive data
	r.extractResponseData(&result, apiResponse, sample)

	return result
}

// extractResponseData extracts all relevant data from API response
func (r *RailwayE2ETestRunner) extractResponseData(result *RailwayE2ETestResult, response map[string]interface{}, sample TestSample) {
	// Basic success indicators
	result.Success = extractBool(response, "success")
	result.ActualIndustry = extractString(response, "primary_industry")
	result.ConfidenceScore = extractFloat64(response, "confidence_score")
	result.CacheHit = extractBool(response, "from_cache")
	result.Explanation = extractString(response, "explanation")

	// Extract explanation metrics
	if result.Explanation != "" {
		result.ExplanationLength = len(result.Explanation)
		result.ExplanationQuality = r.calculateExplanationQuality(result.Explanation)
		result.ExplanationKeywords = r.extractKeywordsFromExplanation(result.Explanation)
	}

	// Extract classification codes
	if classification, ok := response["classification"].(map[string]interface{}); ok {
		result.MCCCodes = extractIndustryCodes(classification, "mcc_codes")
		result.NAICSCodes = extractIndustryCodes(classification, "naics_codes")
		result.SICCodes = extractIndustryCodes(classification, "sic_codes")

		// Check top 3 codes
		result.Top3CodesGenerated = len(result.MCCCodes) >= 3 || len(result.NAICSCodes) >= 3 || len(result.SICCodes) >= 3

		// Calculate average code confidence
		result.CodeConfidenceAvg = r.calculateAverageCodeConfidence(result.MCCCodes, result.NAICSCodes, result.SICCodes)

		// Validate code descriptions
		result.CodeDescriptionsValid = r.validateCodeDescriptions(result.MCCCodes, result.NAICSCodes, result.SICCodes)

		// Extract structured explanation
		if explanation, ok := classification["explanation"].(map[string]interface{}); ok {
			result.ExplanationStructured = explanation
		}
	}

	// Extract metadata for scraping and crawling
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		result.ScrapingStrategy = extractString(metadata, "scraping_strategy")
		result.EarlyExit = extractBool(metadata, "early_exit")
		result.FallbackUsed = extractBool(metadata, "fallback_used")
		result.FallbackType = extractString(metadata, "fallback_type")
		result.ScrapingTime = DurationMsFromDuration(extractDuration(metadata, "scraping_time_ms"))
		result.ClassificationTime = DurationMsFromDuration(extractDuration(metadata, "classification_time_ms"))

		// Extract crawling metrics
		if pagesCrawled, ok := metadata["pages_crawled"].(float64); ok {
			result.PagesCrawled = int(pagesCrawled)
		}
		if pagesAnalyzed, ok := metadata["pages_analyzed"].(float64); ok {
			result.PagesAnalyzed = int(pagesAnalyzed)
		}
		if structuredData, ok := metadata["structured_data_found"].(bool); ok {
			result.StructuredDataFound = structuredData
		}
		if structuredDataType, ok := metadata["structured_data_type"].(string); ok {
			result.StructuredDataType = structuredDataType
		}
	}

	// Determine crawling success
	result.CrawlingSuccess = result.PagesCrawled > 0 || result.WebsiteURL == ""

	// Check accuracy
	if sample.ExpectedIndustry != "" {
		result.IndustryMatch = normalizeIndustryName(result.ActualIndustry) == normalizeIndustryName(sample.ExpectedIndustry)
		result.Accuracy = result.IndustryMatch
	}

	// Enhanced code accuracy calculation
	if len(sample.ExpectedMCC) > 0 && len(result.MCCCodes) > 0 {
		result.MCCTop1Match, result.MCCTop3Match, result.MCCAccuracyScore, result.MCCMatchedRank = 
			r.calculateCodeAccuracy(sample.ExpectedMCC, result.MCCCodes)
		result.MCCMatch = result.MCCTop3Match // Backward compatibility
	}
	
	if len(sample.ExpectedNAICS) > 0 && len(result.NAICSCodes) > 0 {
		result.NAICSTop1Match, result.NAICSTop3Match, result.NAICSAccuracyScore, result.NAICSMatchedRank = 
			r.calculateCodeAccuracy(sample.ExpectedNAICS, result.NAICSCodes)
		result.NAICSMatch = result.NAICSTop3Match // Backward compatibility
	}
	
	if len(sample.ExpectedSIC) > 0 && len(result.SICCodes) > 0 {
		result.SICTop1Match, result.SICTop3Match, result.SICAccuracyScore, result.SICMatchedRank = 
			r.calculateCodeAccuracy(sample.ExpectedSIC, result.SICCodes)
		result.SICMatch = result.SICTop3Match // Backward compatibility
	}
}

// Helper methods for analysis
func (r *RailwayE2ETestRunner) calculateExplanationQuality(explanation string) float64 {
	quality := 0.0
	
	// Length check (good explanations are 50-500 characters)
	if len(explanation) >= 50 && len(explanation) <= 500 {
		quality += 0.3
	} else if len(explanation) > 0 {
		quality += 0.1
	}

	// Keyword presence
	keywords := []string{"industry", "business", "classification", "code", "primary", "confidence"}
	foundKeywords := 0
	lowerExplanation := strings.ToLower(explanation)
	for _, keyword := range keywords {
		if strings.Contains(lowerExplanation, keyword) {
			foundKeywords++
		}
	}
	quality += float64(foundKeywords) / float64(len(keywords)) * 0.4

	// Structure check (sentences, punctuation)
	if strings.Contains(explanation, ".") {
		quality += 0.2
	}
	if strings.Count(explanation, " ") >= 5 {
		quality += 0.1
	}

	return quality
}

func (r *RailwayE2ETestRunner) extractKeywordsFromExplanation(explanation string) []string {
	// Simple keyword extraction
	words := strings.Fields(strings.ToLower(explanation))
	keywords := []string{}
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}
	
	for _, word := range words {
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func (r *RailwayE2ETestRunner) calculateAverageCodeConfidence(mcc, naics, sic []IndustryCode) float64 {
	allCodes := append(append(mcc, naics...), sic...)
	if len(allCodes) == 0 {
		return 0.0
	}
	
	sum := 0.0
	for _, code := range allCodes {
		sum += code.Confidence
	}
	return sum / float64(len(allCodes))
}

func (r *RailwayE2ETestRunner) validateCodeDescriptions(mcc, naics, sic []IndustryCode) bool {
	allCodes := append(append(mcc, naics...), sic...)
	if len(allCodes) == 0 {
		return false
	}
	
	validCount := 0
	for _, code := range allCodes {
		if code.Description != "" && len(code.Description) > 5 {
			validCount++
		}
	}
	
	return float64(validCount)/float64(len(allCodes)) >= 0.8 // 80% should have valid descriptions
}

// calculateCodeAccuracy calculates rank-based code accuracy
// Returns: top1Match, top3Match, accuracyScore (0.0-1.0), matchedRank (0 = not found, 1 = top 1, etc.)
func (r *RailwayE2ETestRunner) calculateCodeAccuracy(
	expectedCodes []string,
	actualCodes []IndustryCode,
) (top1Match bool, top3Match bool, accuracyScore float64, matchedRank int) {
	if len(expectedCodes) == 0 || len(actualCodes) == 0 {
		return false, false, 0.0, 0
	}
	
	expectedMap := make(map[string]bool)
	for _, code := range expectedCodes {
		expectedMap[code] = true
	}
	
	// Check top 1 (rank 1)
	if len(actualCodes) > 0 && expectedMap[actualCodes[0].Code] {
		top1Match = true
		top3Match = true
		accuracyScore = 1.0 // Perfect score for top 1
		matchedRank = 1
		return
	}
	
	// Check top 3 with rank-based scoring
	for i := 1; i < 3 && i < len(actualCodes); i++ {
		if expectedMap[actualCodes[i].Code] {
			top3Match = true
			// Rank-based scoring: rank 1 = 1.0, rank 2 = 0.9, rank 3 = 0.8
			accuracyScore = 1.0 - (float64(i) * 0.1)
			matchedRank = i + 1
			return
		}
	}
	
	// Code not found in top 3
	return false, false, 0.0, 0
}

// addResult adds a result to the collection
func (r *RailwayE2ETestRunner) addResult(result RailwayE2ETestResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, result)
}

// CalculateMetrics calculates comprehensive metrics
func (r *RailwayE2ETestRunner) CalculateMetrics() {
	r.metrics.TotalTests = len(r.results)
	if r.metrics.TotalTests == 0 {
		return
	}

	successful := 0
	scrapingSuccess := 0
	classificationAccurate := 0
	codeGenerated := 0
	top3Generated := 0
	explanationGenerated := 0
	cacheHits := 0
	earlyExits := 0
	fallbacks := 0
	errors := 0

	var totalPagesCrawled int
	var totalConfidence float64
	var totalCodeConfidence float64
	var totalExplanationQuality float64
	var totalExplanationLength int
	latencies := []time.Duration{}
	strategyCounts := make(map[string]int)
	strategySuccess := make(map[string]int)
	industryAccuracy := make(map[string]struct{correct, total int})
	errorTypes := make(map[string]int)
	
	// Enhanced code accuracy tracking
	var mccTop1Matches, mccTop3Matches int
	var naicsTop1Matches, naicsTop3Matches int
	var sicTop1Matches, sicTop3Matches int
	var totalMCCAccuracy, totalNAICSAccuracy, totalSICAccuracy float64
	var mccChecks, naicsChecks, sicChecks int
	codeAccuracyByIndustry := make(map[string]struct{
		mccTop1, mccTop3, naicsTop1, naicsTop3, sicTop1, sicTop3 int
		mccScore, naicsScore, sicScore float64
		total int
	})

	for _, result := range r.results {
		if result.Success {
			successful++
		}
		if result.CrawlingSuccess {
			scrapingSuccess++
		}
		if result.IndustryMatch {
			classificationAccurate++
		}
		if len(result.MCCCodes) > 0 || len(result.NAICSCodes) > 0 || len(result.SICCodes) > 0 {
			codeGenerated++
		}
		if result.Top3CodesGenerated {
			top3Generated++
		}
		if result.Explanation != "" {
			explanationGenerated++
			totalExplanationQuality += result.ExplanationQuality
			totalExplanationLength += result.ExplanationLength
		}
		if result.CacheHit {
			cacheHits++
		}
		if result.EarlyExit {
			earlyExits++
		}
		if result.FallbackUsed {
			fallbacks++
		}
		if result.Error != "" {
			errors++
			errorTypes[result.ErrorType]++
		}

		totalPagesCrawled += result.PagesCrawled
		totalConfidence += result.ConfidenceScore
		if result.CodeConfidenceAvg > 0 {
			totalCodeConfidence += result.CodeConfidenceAvg
		}
		latencies = append(latencies, result.ProcessingTime.ToDuration())

		if result.ScrapingStrategy != "" {
			strategyCounts[result.ScrapingStrategy]++
			if result.CrawlingSuccess {
				strategySuccess[result.ScrapingStrategy]++
			}
		}

		// Track accuracy by industry
		if result.ExpectedIndustry != "" {
			stats := industryAccuracy[result.ExpectedIndustry]
			if result.IndustryMatch {
				stats.correct++
			}
			stats.total++
			industryAccuracy[result.ExpectedIndustry] = stats
		}
		
		// Enhanced code accuracy tracking
		if len(result.ExpectedMCC) > 0 && len(result.MCCCodes) > 0 {
			mccChecks++
			if result.MCCTop1Match {
				mccTop1Matches++
			}
			if result.MCCTop3Match {
				mccTop3Matches++
			}
			totalMCCAccuracy += result.MCCAccuracyScore
			
			// Track by industry
			if result.ExpectedIndustry != "" {
				stats := codeAccuracyByIndustry[result.ExpectedIndustry]
				if result.MCCTop1Match {
					stats.mccTop1++
				}
				if result.MCCTop3Match {
					stats.mccTop3++
				}
				stats.mccScore += result.MCCAccuracyScore
				stats.total++
				codeAccuracyByIndustry[result.ExpectedIndustry] = stats
			}
		}
		
		if len(result.ExpectedNAICS) > 0 && len(result.NAICSCodes) > 0 {
			naicsChecks++
			if result.NAICSTop1Match {
				naicsTop1Matches++
			}
			if result.NAICSTop3Match {
				naicsTop3Matches++
			}
			totalNAICSAccuracy += result.NAICSAccuracyScore
			
			// Track by industry
			if result.ExpectedIndustry != "" {
				stats := codeAccuracyByIndustry[result.ExpectedIndustry]
				if result.NAICSTop1Match {
					stats.naicsTop1++
				}
				if result.NAICSTop3Match {
					stats.naicsTop3++
				}
				stats.naicsScore += result.NAICSAccuracyScore
				stats.total++
				codeAccuracyByIndustry[result.ExpectedIndustry] = stats
			}
		}
		
		if len(result.ExpectedSIC) > 0 && len(result.SICCodes) > 0 {
			sicChecks++
			if result.SICTop1Match {
				sicTop1Matches++
			}
			if result.SICTop3Match {
				sicTop3Matches++
			}
			totalSICAccuracy += result.SICAccuracyScore
			
			// Track by industry
			if result.ExpectedIndustry != "" {
				stats := codeAccuracyByIndustry[result.ExpectedIndustry]
				if result.SICTop1Match {
					stats.sicTop1++
				}
				if result.SICTop3Match {
					stats.sicTop3++
				}
				stats.sicScore += result.SICAccuracyScore
				stats.total++
				codeAccuracyByIndustry[result.ExpectedIndustry] = stats
			}
		}
	}

	// Calculate rates
	r.metrics.SuccessfulTests = successful
	r.metrics.FailedTests = r.metrics.TotalTests - successful
	r.metrics.ScrapingSuccessRate = float64(scrapingSuccess) / float64(r.metrics.TotalTests)
	r.metrics.ClassificationAccuracy = float64(classificationAccurate) / float64(r.metrics.TotalTests)
	r.metrics.CodeGenerationRate = float64(codeGenerated) / float64(r.metrics.TotalTests)
	r.metrics.Top3CodeRate = float64(top3Generated) / float64(r.metrics.TotalTests)
	r.metrics.ExplanationGenerationRate = float64(explanationGenerated) / float64(r.metrics.TotalTests)
	r.metrics.CacheHitRate = float64(cacheHits) / float64(r.metrics.TotalTests)
	r.metrics.EarlyExitRate = float64(earlyExits) / float64(r.metrics.TotalTests)
	r.metrics.FallbackRate = float64(fallbacks) / float64(r.metrics.TotalTests)
	r.metrics.ErrorRate = float64(errors) / float64(r.metrics.TotalTests)

	// Calculate averages
	if r.metrics.TotalTests > 0 {
		r.metrics.AveragePagesCrawled = float64(totalPagesCrawled) / float64(r.metrics.TotalTests)
		r.metrics.AverageConfidence = totalConfidence / float64(r.metrics.TotalTests)
		if codeGenerated > 0 {
			r.metrics.CodeConfidenceAvg = totalCodeConfidence / float64(codeGenerated)
		}
		if explanationGenerated > 0 {
			r.metrics.AverageExplanationQuality = totalExplanationQuality / float64(explanationGenerated)
			r.metrics.AverageExplanationLength = totalExplanationLength / explanationGenerated
		}
	}

	// Calculate latency percentiles
	sortDurations(latencies)
	if len(latencies) > 0 {
		r.metrics.AverageLatency = averageDuration(latencies)
		r.metrics.P50Latency = percentileDuration(latencies, 50)
		r.metrics.P95Latency = percentileDuration(latencies, 95)
		r.metrics.P99Latency = percentileDuration(latencies, 99)
	}

	// Strategy distribution
	r.metrics.StrategyDistribution = strategyCounts
	for strategy, count := range strategyCounts {
		if count > 0 {
			r.metrics.StrategySuccessRate[strategy] = float64(strategySuccess[strategy]) / float64(count)
		}
	}

	// Industry accuracy
	for industry, stats := range industryAccuracy {
		if stats.total > 0 {
			r.metrics.IndustryAccuracy[industry] = float64(stats.correct) / float64(stats.total)
		}
	}
	
	// Enhanced code accuracy metrics
	if mccChecks > 0 {
		r.metrics.CodeAccuracy.MCCTop1Accuracy = float64(mccTop1Matches) / float64(mccChecks)
		r.metrics.CodeAccuracy.MCCTop3Accuracy = float64(mccTop3Matches) / float64(mccChecks)
		r.metrics.CodeAccuracy.MCCAccuracyScore = totalMCCAccuracy / float64(mccChecks)
	}
	
	if naicsChecks > 0 {
		r.metrics.CodeAccuracy.NAICSTop1Accuracy = float64(naicsTop1Matches) / float64(naicsChecks)
		r.metrics.CodeAccuracy.NAICSTop3Accuracy = float64(naicsTop3Matches) / float64(naicsChecks)
		r.metrics.CodeAccuracy.NAICSAccuracyScore = totalNAICSAccuracy / float64(naicsChecks)
	}
	
	if sicChecks > 0 {
		r.metrics.CodeAccuracy.SICTop1Accuracy = float64(sicTop1Matches) / float64(sicChecks)
		r.metrics.CodeAccuracy.SICTop3Accuracy = float64(sicTop3Matches) / float64(sicChecks)
		r.metrics.CodeAccuracy.SICAccuracyScore = totalSICAccuracy / float64(sicChecks)
	}
	
	// Overall code accuracy (weighted average)
	totalAccuracy := 0.0
	count := 0
	if mccChecks > 0 {
		totalAccuracy += r.metrics.CodeAccuracy.MCCAccuracyScore
		count++
	}
	if naicsChecks > 0 {
		totalAccuracy += r.metrics.CodeAccuracy.NAICSAccuracyScore
		count++
	}
	if sicChecks > 0 {
		totalAccuracy += r.metrics.CodeAccuracy.SICAccuracyScore
		count++
	}
	if count > 0 {
		r.metrics.CodeAccuracy.OverallCodeAccuracy = totalAccuracy / float64(count)
	}
	
	// Code accuracy by industry
	for industry, stats := range codeAccuracyByIndustry {
		if stats.total > 0 {
			industryMetrics := EnhancedCodeAccuracyMetrics{}
			industryMetrics.MCCTop1Accuracy = float64(stats.mccTop1) / float64(stats.total)
			industryMetrics.MCCTop3Accuracy = float64(stats.mccTop3) / float64(stats.total)
			industryMetrics.MCCAccuracyScore = stats.mccScore / float64(stats.total)
			industryMetrics.NAICSTop1Accuracy = float64(stats.naicsTop1) / float64(stats.total)
			industryMetrics.NAICSTop3Accuracy = float64(stats.naicsTop3) / float64(stats.total)
			industryMetrics.NAICSAccuracyScore = stats.naicsScore / float64(stats.total)
			industryMetrics.SICTop1Accuracy = float64(stats.sicTop1) / float64(stats.total)
			industryMetrics.SICTop3Accuracy = float64(stats.sicTop3) / float64(stats.total)
			industryMetrics.SICAccuracyScore = stats.sicScore / float64(stats.total)
			
			// Overall for this industry
			industryCount := 0
			industryTotal := 0.0
			if stats.mccScore > 0 {
				industryTotal += industryMetrics.MCCAccuracyScore
				industryCount++
			}
			if stats.naicsScore > 0 {
				industryTotal += industryMetrics.NAICSAccuracyScore
				industryCount++
			}
			if stats.sicScore > 0 {
				industryTotal += industryMetrics.SICAccuracyScore
				industryCount++
			}
			if industryCount > 0 {
				industryMetrics.OverallCodeAccuracy = industryTotal / float64(industryCount)
			}
			
			r.metrics.CodeAccuracy.CodeAccuracyByIndustry[industry] = industryMetrics
		}
	}

	// Error distribution
	r.metrics.ErrorDistribution = errorTypes
}

// GenerateComprehensiveReport generates a detailed report
func (r *RailwayE2ETestRunner) GenerateComprehensiveReport(totalDuration time.Duration) map[string]interface{} {
	report := map[string]interface{}{
		"test_summary": map[string]interface{}{
			"total_samples":    r.metrics.TotalTests,
			"successful_tests":  r.metrics.SuccessfulTests,
			"failed_tests":      r.metrics.FailedTests,
			"test_duration":     totalDuration.String(),
			"timestamp":         time.Now().Format(time.RFC3339),
			"api_url":           r.apiURL,
		},
		"scraping_metrics": map[string]interface{}{
			"scraping_success_rate":   r.metrics.ScrapingSuccessRate,
			"average_pages_crawled":   r.metrics.AveragePagesCrawled,
			"strategy_distribution":   r.metrics.StrategyDistribution,
			"strategy_success_rate":   r.metrics.StrategySuccessRate,
			"structured_data_rate":    r.metrics.StructuredDataRate,
		},
		"classification_metrics": map[string]interface{}{
			"classification_accuracy": r.metrics.ClassificationAccuracy,
			"average_confidence":      r.metrics.AverageConfidence,
			"industry_accuracy":       r.metrics.IndustryAccuracy,
		},
		"code_generation_metrics": map[string]interface{}{
			"code_generation_rate": r.metrics.CodeGenerationRate,
			"top3_code_rate":      r.metrics.Top3CodeRate,
			"code_confidence_avg": r.metrics.CodeConfidenceAvg,
		},
		"code_accuracy_metrics": map[string]interface{}{
			"mcc_accuracy_score":   r.metrics.CodeAccuracy.MCCAccuracyScore,
			"naics_accuracy_score": r.metrics.CodeAccuracy.NAICSAccuracyScore,
			"sic_accuracy_score":   r.metrics.CodeAccuracy.SICAccuracyScore,
			"mcc_top1_accuracy":    r.metrics.CodeAccuracy.MCCTop1Accuracy,
			"mcc_top3_accuracy":    r.metrics.CodeAccuracy.MCCTop3Accuracy,
			"naics_top1_accuracy":  r.metrics.CodeAccuracy.NAICSTop1Accuracy,
			"naics_top3_accuracy":  r.metrics.CodeAccuracy.NAICSTop3Accuracy,
			"sic_top1_accuracy":    r.metrics.CodeAccuracy.SICTop1Accuracy,
			"sic_top3_accuracy":    r.metrics.CodeAccuracy.SICTop3Accuracy,
			"overall_code_accuracy": r.metrics.CodeAccuracy.OverallCodeAccuracy,
			"code_accuracy_by_industry": r.metrics.CodeAccuracy.CodeAccuracyByIndustry,
		},
		"explanation_metrics": map[string]interface{}{
			"explanation_generation_rate": r.metrics.ExplanationGenerationRate,
			"average_explanation_quality": r.metrics.AverageExplanationQuality,
			"average_explanation_length":  r.metrics.AverageExplanationLength,
		},
		"performance_metrics": map[string]interface{}{
			"average_latency_ms": r.metrics.AverageLatency.Milliseconds(),
			"p50_latency_ms":     r.metrics.P50Latency.Milliseconds(),
			"p95_latency_ms":     r.metrics.P95Latency.Milliseconds(),
			"p99_latency_ms":     r.metrics.P99Latency.Milliseconds(),
			"cache_hit_rate":     r.metrics.CacheHitRate,
			"early_exit_rate":    r.metrics.EarlyExitRate,
			"fallback_rate":      r.metrics.FallbackRate,
		},
		"error_analysis": map[string]interface{}{
			"error_rate":         r.metrics.ErrorRate,
			"error_distribution": r.metrics.ErrorDistribution,
		},
		"detailed_results": r.results,
	}

	return report
}

// AnalyzeClassificationProcess analyzes strengths, weaknesses, and opportunities
func (r *RailwayE2ETestRunner) AnalyzeClassificationProcess(report map[string]interface{}) map[string]interface{} {
	analysis := map[string]interface{}{
		"strengths":     []string{},
		"weaknesses":   []string{},
		"opportunities": []string{},
		"recommendations": []string{},
	}

	scrapingMetrics := report["scraping_metrics"].(map[string]interface{})
	classificationMetrics := report["classification_metrics"].(map[string]interface{})
	codeMetrics := report["code_generation_metrics"].(map[string]interface{})
	codeAccuracyMetrics := report["code_accuracy_metrics"].(map[string]interface{})
	explanationMetrics := report["explanation_metrics"].(map[string]interface{})
	performanceMetrics := report["performance_metrics"].(map[string]interface{})

	// Analyze strengths
	scrapingSuccessRate := scrapingMetrics["scraping_success_rate"].(float64)
	if scrapingSuccessRate >= 0.9 {
		analysis["strengths"] = append(analysis["strengths"].([]string), 
			fmt.Sprintf("High scraping success rate (%.1f%%)", scrapingSuccessRate*100))
	}

	classificationAccuracy := classificationMetrics["classification_accuracy"].(float64)
	if classificationAccuracy >= 0.85 {
		analysis["strengths"] = append(analysis["strengths"].([]string),
			fmt.Sprintf("Good classification accuracy (%.1f%%)", classificationAccuracy*100))
	}

	codeGenRate := codeMetrics["code_generation_rate"].(float64)
	if codeGenRate >= 0.95 {
		analysis["strengths"] = append(analysis["strengths"].([]string),
			fmt.Sprintf("Excellent code generation rate (%.1f%%)", codeGenRate*100))
	}
	
	// Code accuracy strengths
	overallCodeAccuracy := codeAccuracyMetrics["overall_code_accuracy"].(float64)
	if overallCodeAccuracy >= 0.85 {
		analysis["strengths"] = append(analysis["strengths"].([]string),
			fmt.Sprintf("High overall code accuracy (%.1f%%)", overallCodeAccuracy*100))
	}
	
	mccTop1Accuracy := codeAccuracyMetrics["mcc_top1_accuracy"].(float64)
	if mccTop1Accuracy >= 0.7 {
		analysis["strengths"] = append(analysis["strengths"].([]string),
			fmt.Sprintf("Strong MCC top 1 accuracy (%.1f%%)", mccTop1Accuracy*100))
	}

	explanationGenRate := explanationMetrics["explanation_generation_rate"].(float64)
	if explanationGenRate >= 0.9 {
		analysis["strengths"] = append(analysis["strengths"].([]string),
			fmt.Sprintf("High explanation generation rate (%.1f%%)", explanationGenRate*100))
	}

	// Analyze weaknesses
	if scrapingSuccessRate < 0.7 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("Low scraping success rate (%.1f%%) - may need better error handling or retry logic", scrapingSuccessRate*100))
	}

	if classificationAccuracy < 0.8 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("Classification accuracy below target (%.1f%%) - may need improved classification algorithms", classificationAccuracy*100))
	}
	
	// Code accuracy weaknesses
	if overallCodeAccuracy < 0.7 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("Low overall code accuracy (%.1f%%) - may need improved code matching algorithms", overallCodeAccuracy*100))
	}
	
	if mccTop1Accuracy < 0.5 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("Low MCC top 1 accuracy (%.1f%%) - primary code selection needs improvement", mccTop1Accuracy*100))
	}

	avgLatency := performanceMetrics["average_latency_ms"].(int64)
	if avgLatency > 5000 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("High average latency (%dms) - may need performance optimization", avgLatency))
	}

	errorRate := report["error_analysis"].(map[string]interface{})["error_rate"].(float64)
	if errorRate > 0.1 {
		analysis["weaknesses"] = append(analysis["weaknesses"].([]string),
			fmt.Sprintf("High error rate (%.1f%%) - needs investigation", errorRate*100))
	}

	// Identify opportunities
	cacheHitRate := performanceMetrics["cache_hit_rate"].(float64)
	if cacheHitRate < 0.5 {
		analysis["opportunities"] = append(analysis["opportunities"].([]string),
			"Increase cache hit rate to improve performance and reduce costs")
	}

	top3CodeRate := codeMetrics["top3_code_rate"].(float64)
	if top3CodeRate < 0.9 {
		analysis["opportunities"] = append(analysis["opportunities"].([]string),
			"Improve top 3 code generation rate for better classification completeness")
	}
	
	// Code accuracy opportunities
	mccTop3Accuracy := codeAccuracyMetrics["mcc_top3_accuracy"].(float64)
	if mccTop3Accuracy > mccTop1Accuracy && mccTop1Accuracy < 0.6 {
		analysis["opportunities"] = append(analysis["opportunities"].([]string),
			"Improve MCC top 1 accuracy - codes are often in top 3 but not top 1")
	}

	explanationQuality := explanationMetrics["average_explanation_quality"].(float64)
	if explanationQuality < 0.7 {
		analysis["opportunities"] = append(analysis["opportunities"].([]string),
			"Enhance explanation quality to provide more meaningful insights")
	}

	// Generate recommendations
	if classificationAccuracy < 0.9 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string),
			"Consider implementing ensemble methods or ML models to improve classification accuracy")
	}
	
	if overallCodeAccuracy < 0.8 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string),
			"Review and improve code matching algorithms, especially for top 1 position accuracy")
	}

	if avgLatency > 3000 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string),
			"Optimize scraping and crawling strategies to reduce latency")
	}

	if errorRate > 0.05 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string),
			"Implement better error handling and retry mechanisms")
	}

	return analysis
}

// PrintComprehensiveSummary prints a detailed summary
func (r *RailwayE2ETestRunner) PrintComprehensiveSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("RAILWAY COMPREHENSIVE E2E CLASSIFICATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	
	fmt.Printf("\n📊 Test Overview:\n")
	fmt.Printf("  Total Tests: %d\n", r.metrics.TotalTests)
	fmt.Printf("  Successful: %d (%.1f%%)\n", r.metrics.SuccessfulTests, 
		float64(r.metrics.SuccessfulTests)/float64(r.metrics.TotalTests)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n", r.metrics.FailedTests,
		float64(r.metrics.FailedTests)/float64(r.metrics.TotalTests)*100)

	fmt.Printf("\n🌐 Scraping & Crawling:\n")
	fmt.Printf("  Success Rate: %.1f%%\n", r.metrics.ScrapingSuccessRate*100)
	fmt.Printf("  Avg Pages Crawled: %.1f\n", r.metrics.AveragePagesCrawled)
	fmt.Printf("  Strategy Distribution: %v\n", r.metrics.StrategyDistribution)

	fmt.Printf("\n🎯 Classification:\n")
	fmt.Printf("  Accuracy: %.1f%%\n", r.metrics.ClassificationAccuracy*100)
	fmt.Printf("  Avg Confidence: %.2f\n", r.metrics.AverageConfidence)

	fmt.Printf("\n📝 Code Generation:\n")
	fmt.Printf("  Generation Rate: %.1f%%\n", r.metrics.CodeGenerationRate*100)
	fmt.Printf("  Top 3 Code Rate: %.1f%%\n", r.metrics.Top3CodeRate*100)
	fmt.Printf("  Avg Code Confidence: %.2f\n", r.metrics.CodeConfidenceAvg)
	
	fmt.Printf("\n🎯 Code Accuracy (Enhanced):\n")
	fmt.Printf("  Overall Code Accuracy: %.1f%%\n", r.metrics.CodeAccuracy.OverallCodeAccuracy*100)
	fmt.Printf("  MCC - Top 1: %.1f%%, Top 3: %.1f%%, Score: %.2f\n", 
		r.metrics.CodeAccuracy.MCCTop1Accuracy*100,
		r.metrics.CodeAccuracy.MCCTop3Accuracy*100,
		r.metrics.CodeAccuracy.MCCAccuracyScore)
	fmt.Printf("  NAICS - Top 1: %.1f%%, Top 3: %.1f%%, Score: %.2f\n",
		r.metrics.CodeAccuracy.NAICSTop1Accuracy*100,
		r.metrics.CodeAccuracy.NAICSTop3Accuracy*100,
		r.metrics.CodeAccuracy.NAICSAccuracyScore)
	fmt.Printf("  SIC - Top 1: %.1f%%, Top 3: %.1f%%, Score: %.2f\n",
		r.metrics.CodeAccuracy.SICTop1Accuracy*100,
		r.metrics.CodeAccuracy.SICTop3Accuracy*100,
		r.metrics.CodeAccuracy.SICAccuracyScore)

	fmt.Printf("\n💬 Explanation Generation:\n")
	fmt.Printf("  Generation Rate: %.1f%%\n", r.metrics.ExplanationGenerationRate*100)
	fmt.Printf("  Avg Quality: %.2f\n", r.metrics.AverageExplanationQuality)
	fmt.Printf("  Avg Length: %d chars\n", r.metrics.AverageExplanationLength)

	fmt.Printf("\n⚡ Performance:\n")
	fmt.Printf("  Avg Latency: %v\n", r.metrics.AverageLatency)
	fmt.Printf("  P95 Latency: %v\n", r.metrics.P95Latency)
	fmt.Printf("  Cache Hit Rate: %.1f%%\n", r.metrics.CacheHitRate*100)
	fmt.Printf("  Early Exit Rate: %.1f%%\n", r.metrics.EarlyExitRate*100)

	fmt.Printf("\n❌ Errors:\n")
	fmt.Printf("  Error Rate: %.1f%%\n", r.metrics.ErrorRate*100)
	fmt.Printf("  Error Distribution: %v\n", r.metrics.ErrorDistribution)

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// validateE2EResults validates test results
func validateE2EResults(t *testing.T, report map[string]interface{}) {
	scrapingMetrics := report["scraping_metrics"].(map[string]interface{})
	classificationMetrics := report["classification_metrics"].(map[string]interface{})
	codeMetrics := report["code_generation_metrics"].(map[string]interface{})
	codeAccuracyMetrics := report["code_accuracy_metrics"].(map[string]interface{})
	performanceMetrics := report["performance_metrics"].(map[string]interface{})

	// Validate scraping success rate
	scrapingSuccessRate := scrapingMetrics["scraping_success_rate"].(float64)
	if scrapingSuccessRate < 0.7 {
		t.Errorf("❌ Scraping success rate below threshold: %.1f%% (expected ≥70%%)", scrapingSuccessRate*100)
	} else {
		t.Logf("✅ Scraping success rate: %.1f%%", scrapingSuccessRate*100)
	}

	// Validate classification accuracy
	classificationAccuracy := classificationMetrics["classification_accuracy"].(float64)
	if classificationAccuracy < 0.8 {
		t.Errorf("❌ Classification accuracy below threshold: %.1f%% (expected ≥80%%)", classificationAccuracy*100)
	} else {
		t.Logf("✅ Classification accuracy: %.1f%%", classificationAccuracy*100)
	}

	// Validate code generation
	codeGenRate := codeMetrics["code_generation_rate"].(float64)
	if codeGenRate < 0.9 {
		t.Errorf("❌ Code generation rate below threshold: %.1f%% (expected ≥90%%)", codeGenRate*100)
	} else {
		t.Logf("✅ Code generation rate: %.1f%%", codeGenRate*100)
	}
	
	// Validate code accuracy
	overallCodeAccuracy := codeAccuracyMetrics["overall_code_accuracy"].(float64)
	if overallCodeAccuracy < 0.7 {
		t.Errorf("❌ Overall code accuracy below threshold: %.1f%% (expected ≥70%%)", overallCodeAccuracy*100)
	} else {
		t.Logf("✅ Overall code accuracy: %.1f%%", overallCodeAccuracy*100)
	}
	
	mccTop3Accuracy := codeAccuracyMetrics["mcc_top3_accuracy"].(float64)
	if mccTop3Accuracy < 0.6 {
		t.Errorf("❌ MCC top 3 accuracy below threshold: %.1f%% (expected ≥60%%)", mccTop3Accuracy*100)
	} else {
		t.Logf("✅ MCC top 3 accuracy: %.1f%%", mccTop3Accuracy*100)
	}

	// Validate performance
	avgLatency := performanceMetrics["average_latency_ms"].(int64)
	if avgLatency > 10000 {
		t.Errorf("❌ Average latency too high: %dms (expected <10000ms)", avgLatency)
	} else {
		t.Logf("✅ Average latency: %dms", avgLatency)
	}
}

