package infrastructure

import (
	"context"
	"log"
	"testing"
	"time"
)

// TestRuleEngineIntegration tests the complete rule engine integration
func TestRuleEngineIntegration(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[INTEGRATION_TEST] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Start rule engine
	if err := ruleEngine.Start(ctx); err != nil {
		t.Fatalf("Failed to start rule engine: %v", err)
	}
	defer ruleEngine.Stop()

	// Create performance optimizer
	optimizer := NewRuleEnginePerformanceOptimizer(logger)

	// Start performance monitoring
	if err := optimizer.StartMonitoring(ctx); err != nil {
		t.Fatalf("Failed to start performance monitoring: %v", err)
	}

	// Optimize rule engine for performance
	if err := optimizer.OptimizeRuleEngine(ruleEngine); err != nil {
		t.Fatalf("Failed to optimize rule engine: %v", err)
	}

	// Create accuracy tester
	accuracyTester := NewRuleEngineAccuracyTester(logger)

	// Load comprehensive test dataset
	testDataset := createComprehensiveTestDataset()
	if err := accuracyTester.LoadTestDataset(testDataset); err != nil {
		t.Fatalf("Failed to load test dataset: %v", err)
	}

	// Run comprehensive accuracy test
	metrics, err := accuracyTester.RunAccuracyTest(ctx, ruleEngine, testDataset.Name)
	if err != nil {
		t.Fatalf("Failed to run accuracy test: %v", err)
	}

	// Validate accuracy target (90%+)
	if metrics.Accuracy < 0.90 {
		t.Errorf("❌ Accuracy %.2f%% is below target of 90%%", metrics.Accuracy*100)
	} else {
		t.Logf("✅ Accuracy %.2f%% meets target of 90%%", metrics.Accuracy*100)
	}

	// Validate performance target (sub-10ms average response time)
	if metrics.PerformanceMetrics.AverageResponseTime > 10*time.Millisecond {
		t.Errorf("❌ Average response time %v exceeds target of 10ms", metrics.PerformanceMetrics.AverageResponseTime)
	} else {
		t.Logf("✅ Average response time %v meets target of 10ms", metrics.PerformanceMetrics.AverageResponseTime)
	}

	// Validate P95 response time (sub-15ms)
	if metrics.PerformanceMetrics.P95ResponseTime > 15*time.Millisecond {
		t.Errorf("❌ P95 response time %v exceeds target of 15ms", metrics.PerformanceMetrics.P95ResponseTime)
	} else {
		t.Logf("✅ P95 response time %v meets target of 15ms", metrics.PerformanceMetrics.P95ResponseTime)
	}

	// Validate P99 response time (sub-20ms)
	if metrics.PerformanceMetrics.P99ResponseTime > 20*time.Millisecond {
		t.Errorf("❌ P99 response time %v exceeds target of 20ms", metrics.PerformanceMetrics.P99ResponseTime)
	} else {
		t.Logf("✅ P99 response time %v meets target of 20ms", metrics.PerformanceMetrics.P99ResponseTime)
	}

	// Validate throughput (minimum 100 req/s)
	if metrics.PerformanceMetrics.ThroughputPerSecond < 100 {
		t.Errorf("❌ Throughput %.2f req/s is below target of 100 req/s", metrics.PerformanceMetrics.ThroughputPerSecond)
	} else {
		t.Logf("✅ Throughput %.2f req/s meets target of 100 req/s", metrics.PerformanceMetrics.ThroughputPerSecond)
	}

	// Log comprehensive results
	t.Logf("\n📊 Comprehensive Integration Test Results:")
	t.Logf("  📈 Accuracy Metrics:")
	t.Logf("    - Overall Accuracy: %.2f%%", metrics.Accuracy*100)
	t.Logf("    - Precision: %.2f%%", metrics.Precision*100)
	t.Logf("    - Recall: %.2f%%", metrics.Recall*100)
	t.Logf("    - F1 Score: %.2f%%", metrics.F1Score*100)
	t.Logf("    - Correct Predictions: %d/%d", metrics.CorrectPredictions, metrics.TotalTestCases)

	t.Logf("  ⚡ Performance Metrics:")
	t.Logf("    - Average Response Time: %v", metrics.PerformanceMetrics.AverageResponseTime)
	t.Logf("    - Min Response Time: %v", metrics.PerformanceMetrics.MinResponseTime)
	t.Logf("    - Max Response Time: %v", metrics.PerformanceMetrics.MaxResponseTime)
	t.Logf("    - P95 Response Time: %v", metrics.PerformanceMetrics.P95ResponseTime)
	t.Logf("    - P99 Response Time: %v", metrics.PerformanceMetrics.P99ResponseTime)
	t.Logf("    - Throughput: %.2f req/s", metrics.PerformanceMetrics.ThroughputPerSecond)

	t.Logf("  🔍 Error Analysis:")
	t.Logf("    - Total Errors: %d", len(metrics.ErrorAnalysis))
	if len(metrics.ErrorAnalysis) > 0 {
		errorTypes := make(map[string]int)
		for _, error := range metrics.ErrorAnalysis {
			errorTypes[error.ErrorType]++
		}
		for errorType, count := range errorTypes {
			t.Logf("    - %s: %d", errorType, count)
		}
	}

	// Test individual components
	t.Run("KeywordMatching", func(t *testing.T) {
		testKeywordMatchingComponent(t, ruleEngine)
	})

	t.Run("MCCLookup", func(t *testing.T) {
		testMCCLookupComponent(t, ruleEngine)
	})

	t.Run("BlacklistChecker", func(t *testing.T) {
		testBlacklistCheckerComponent(t, ruleEngine)
	})

	t.Run("CachePerformance", func(t *testing.T) {
		testCachePerformance(t, ruleEngine)
	})

	// Test end-to-end scenarios
	t.Run("EndToEndScenarios", func(t *testing.T) {
		testEndToEndScenarios(t, ruleEngine)
	})
}

// testKeywordMatchingComponent tests keyword matching component
func testKeywordMatchingComponent(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		businessName   string
		description    string
		expectedLabels []string
		minConfidence  float64
	}{
		{
			name:           "Technology Company",
			businessName:   "TechInnovate Solutions",
			description:    "Cutting-edge software development and AI consulting",
			expectedLabels: []string{"technology", "software_development"},
			minConfidence:  0.8,
		},
		{
			name:           "Financial Services",
			businessName:   "SecureBank Financial",
			description:    "Commercial banking and investment services",
			expectedLabels: []string{"financial_services", "banking"},
			minConfidence:  0.8,
		},
		{
			name:           "Healthcare Provider",
			businessName:   "MedCare Plus",
			description:    "Comprehensive medical services and healthcare consulting",
			expectedLabels: []string{"healthcare", "medical_services"},
			minConfidence:  0.8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &RuleEngineClassificationRequest{
				BusinessName: tc.businessName,
				Description:  tc.description,
				WebsiteURL:   "https://test-" + tc.name + ".com",
			}

			result, err := ruleEngine.Classify(ctx, req)
			if err != nil {
				t.Fatalf("Classification failed: %v", err)
			}

			if result.Confidence < tc.minConfidence {
				t.Errorf("Confidence %.2f is below minimum %.2f", result.Confidence, tc.minConfidence)
			}

			// Check if expected labels are found
			foundLabels := make(map[string]bool)
			for _, pred := range result.Classifications {
				foundLabels[pred.Label] = true
			}

			for _, expectedLabel := range tc.expectedLabels {
				if !foundLabels[expectedLabel] {
					t.Errorf("Expected label '%s' not found in predictions", expectedLabel)
				}
			}
		})
	}
}

// testMCCLookupComponent tests MCC lookup component
func testMCCLookupComponent(t *testing.T, ruleEngine *GoRuleEngine) {

	testCases := []struct {
		name               string
		businessName       string
		description        string
		expectedMCCs       []string
		expectedProhibited bool
	}{
		{
			name:               "Software Development",
			businessName:       "CodeCraft Solutions",
			description:        "Custom software development and system integration",
			expectedMCCs:       []string{"7372"},
			expectedProhibited: false,
		},
		{
			name:               "Gambling Services",
			businessName:       "Lucky Slots Casino",
			description:        "Online casino with slots, poker, and sports betting",
			expectedMCCs:       []string{"7995"},
			expectedProhibited: true,
		},
		{
			name:               "Financial Services",
			businessName:       "Business Credit Union",
			description:        "Commercial banking and business lending services",
			expectedMCCs:       []string{"6012"},
			expectedProhibited: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			req := &RuleEngineClassificationRequest{
				BusinessName: tc.businessName,
				Description:  tc.description,
				WebsiteURL:   "https://test-" + tc.name + ".com",
			}

			result, err := ruleEngine.Classify(ctx, req)
			if err != nil {
				t.Fatalf("MCC classification failed: %v", err)
			}

			// Check if expected MCCs are found
			foundMCCs := make(map[string]bool)
			for _, pred := range result.Classifications {
				foundMCCs[pred.Label] = true
			}

			for _, expectedMCC := range tc.expectedMCCs {
				if !foundMCCs[expectedMCC] {
					t.Errorf("Expected MCC '%s' not found in classifications", expectedMCC)
				}
			}
		})
	}
}

// testBlacklistCheckerComponent tests blacklist checker component
func testBlacklistCheckerComponent(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	testCases := []struct {
		name                string
		businessName        string
		websiteURL          string
		expectedBlacklisted bool
	}{
		{
			name:                "Legitimate Business",
			businessName:        "Trustworthy Corp",
			websiteURL:          "https://trustworthy-corp.com",
			expectedBlacklisted: false,
		},
		{
			name:                "Known Bad Actor",
			businessName:        "Fraudulent Services Inc",
			websiteURL:          "https://fraudulent-services.com",
			expectedBlacklisted: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &RuleEngineRiskRequest{
				BusinessName:   tc.businessName,
				Description:    "Test business for blacklist checking",
				WebsiteURL:     tc.websiteURL,
				WebsiteContent: "Test website content",
			}

			result, err := ruleEngine.DetectRisk(ctx, req)
			if err != nil {
				t.Fatalf("Risk detection failed: %v", err)
			}

			isBlacklisted := false
			for _, risk := range result.DetectedRisks {
				if risk.Category == "blacklist" {
					isBlacklisted = true
					break
				}
			}

			if isBlacklisted != tc.expectedBlacklisted {
				t.Errorf("Expected blacklisted status %v, got %v", tc.expectedBlacklisted, isBlacklisted)
			}
		})
	}
}

// testCachePerformance tests cache performance
func testCachePerformance(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	// Test cache hit performance
	req := &RuleEngineClassificationRequest{
		BusinessName: "Cache Test Business",
		Description:  "Testing cache performance",
		WebsiteURL:   "https://cache-test.com",
	}

	// First request (cache miss)
	start := time.Now()
	result1, err := ruleEngine.Classify(ctx, req)
	if err != nil {
		t.Fatalf("First classification failed: %v", err)
	}
	firstRequestTime := time.Since(start)

	// Second request (cache hit)
	start = time.Now()
	result2, err := ruleEngine.Classify(ctx, req)
	if err != nil {
		t.Fatalf("Second classification failed: %v", err)
	}
	secondRequestTime := time.Since(start)

	// Cache hit should be significantly faster
	if secondRequestTime >= firstRequestTime {
		t.Errorf("Cache hit time %v should be less than cache miss time %v", secondRequestTime, firstRequestTime)
	}

	// Results should be identical
	if result1.RequestID != result2.RequestID {
		t.Error("Cached result should be identical to original result")
	}

	t.Logf("Cache Performance: Miss=%v, Hit=%v, Speedup=%.2fx",
		firstRequestTime, secondRequestTime, float64(firstRequestTime)/float64(secondRequestTime))
}

// testEndToEndScenarios tests end-to-end scenarios
func testEndToEndScenarios(t *testing.T, ruleEngine *GoRuleEngine) {

	scenarios := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, ruleEngine *GoRuleEngine)
	}{
		{
			name:        "Low Risk Business Verification",
			description: "Complete verification of a low-risk legitimate business",
			testFunc:    testLowRiskBusinessScenario,
		},
		{
			name:        "High Risk Business Detection",
			description: "Detection and classification of a high-risk business",
			testFunc:    testHighRiskBusinessScenario,
		},
		{
			name:        "Prohibited Business Blocking",
			description: "Blocking of prohibited business activities",
			testFunc:    testProhibitedBusinessScenario,
		},
		{
			name:        "Complex Business Classification",
			description: "Classification of a complex multi-industry business",
			testFunc:    testComplexBusinessScenario,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Testing scenario: %s - %s", scenario.name, scenario.description)
			scenario.testFunc(t, ruleEngine)
		})
	}
}

// testLowRiskBusinessScenario tests low-risk business scenario
func testLowRiskBusinessScenario(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	// Low-risk technology company
	classificationReq := &RuleEngineClassificationRequest{
		BusinessName: "Innovative Tech Solutions",
		Description:  "Software development and technology consulting services",
		WebsiteURL:   "https://innovative-tech.com",
	}

	riskReq := &RuleEngineRiskRequest{
		BusinessName:   "Innovative Tech Solutions",
		Description:    "Software development and technology consulting services",
		WebsiteURL:     "https://innovative-tech.com",
		WebsiteContent: "We provide innovative software solutions and technology consulting",
	}

	// Test classification
	classificationResult, err := ruleEngine.Classify(ctx, classificationReq)
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}

	// Test risk detection
	riskResult, err := ruleEngine.DetectRisk(ctx, riskReq)
	if err != nil {
		t.Fatalf("Risk detection failed: %v", err)
	}

	// Validate low-risk classification
	if classificationResult.Confidence < 0.8 {
		t.Errorf("Expected high confidence for low-risk business, got %.2f", classificationResult.Confidence)
	}

	// Validate low risk score
	if riskResult.RiskScore > 0.3 {
		t.Errorf("Expected low risk score, got %.2f", riskResult.RiskScore)
	}

	// Validate no high-risk categories detected
	for _, risk := range riskResult.DetectedRisks {
		if risk.Severity == "critical" || risk.Severity == "high" {
			t.Errorf("Unexpected high-risk category detected: %s", risk.Category)
		}
	}
}

// testHighRiskBusinessScenario tests high-risk business scenario
func testHighRiskBusinessScenario(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	// High-risk gambling business
	classificationReq := &RuleEngineClassificationRequest{
		BusinessName: "High Stakes Casino",
		Description:  "Online casino and gambling services",
		WebsiteURL:   "https://highstakes-casino.com",
	}

	riskReq := &RuleEngineRiskRequest{
		BusinessName:   "High Stakes Casino",
		Description:    "Online casino and gambling services",
		WebsiteURL:     "https://highstakes-casino.com",
		WebsiteContent: "Welcome to our online casino with slots, poker, and sports betting",
	}

	// Test classification
	_, err := ruleEngine.Classify(ctx, classificationReq)
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}

	// Test risk detection
	riskResult, err := ruleEngine.DetectRisk(ctx, riskReq)
	if err != nil {
		t.Fatalf("Risk detection failed: %v", err)
	}

	// Validate high-risk detection
	if riskResult.RiskScore < 0.7 {
		t.Errorf("Expected high risk score, got %.2f", riskResult.RiskScore)
	}

	// Validate high-risk categories detected
	hasHighRisk := false
	for _, risk := range riskResult.DetectedRisks {
		if risk.Severity == "high" || risk.Severity == "critical" {
			hasHighRisk = true
			break
		}
	}

	if !hasHighRisk {
		t.Error("Expected high-risk categories to be detected")
	}
}

// testProhibitedBusinessScenario tests prohibited business scenario
func testProhibitedBusinessScenario(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	// Prohibited illegal business
	classificationReq := &RuleEngineClassificationRequest{
		BusinessName: "Illegal Drug Store",
		Description:  "Online pharmaceutical and controlled substance sales",
		WebsiteURL:   "https://illegal-drugs.com",
	}

	riskReq := &RuleEngineRiskRequest{
		BusinessName:   "Illegal Drug Store",
		Description:    "Online pharmaceutical and controlled substance sales",
		WebsiteURL:     "https://illegal-drugs.com",
		WebsiteContent: "We sell prescription drugs and controlled substances online without prescription",
	}

	// Test classification
	_, err := ruleEngine.Classify(ctx, classificationReq)
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}

	// Test risk detection
	riskResult, err := ruleEngine.DetectRisk(ctx, riskReq)
	if err != nil {
		t.Fatalf("Risk detection failed: %v", err)
	}

	// Validate critical risk detection
	if riskResult.RiskScore < 0.9 {
		t.Errorf("Expected critical risk score, got %.2f", riskResult.RiskScore)
	}

	// Validate prohibited categories detected
	hasProhibited := false
	for _, risk := range riskResult.DetectedRisks {
		if risk.Category == "illegal" || risk.Category == "prohibited" {
			hasProhibited = true
			break
		}
	}

	if !hasProhibited {
		t.Error("Expected prohibited categories to be detected")
	}
}

// testComplexBusinessScenario tests complex business scenario
func testComplexBusinessScenario(t *testing.T, ruleEngine *GoRuleEngine) {
	ctx := context.Background()

	// Complex multi-industry business
	classificationReq := &RuleEngineClassificationRequest{
		BusinessName: "Global Enterprise Solutions",
		Description:  "Technology consulting, financial services, and healthcare solutions",
		WebsiteURL:   "https://global-enterprise.com",
	}

	riskReq := &RuleEngineRiskRequest{
		BusinessName:   "Global Enterprise Solutions",
		Description:    "Technology consulting, financial services, and healthcare solutions",
		WebsiteURL:     "https://global-enterprise.com",
		WebsiteContent: "We provide comprehensive technology consulting, financial advisory, and healthcare technology solutions",
	}

	// Test classification
	classificationResult, err := ruleEngine.Classify(ctx, classificationReq)
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}

	// Test risk detection
	riskResult, err := ruleEngine.DetectRisk(ctx, riskReq)
	if err != nil {
		t.Fatalf("Risk detection failed: %v", err)
	}

	// Validate multiple industry classifications
	if len(classificationResult.Classifications) < 2 {
		t.Error("Expected multiple industry classifications for complex business")
	}

	// Validate reasonable risk assessment
	if riskResult.RiskScore > 0.5 {
		t.Errorf("Expected moderate risk score for complex business, got %.2f", riskResult.RiskScore)
	}
}

// createComprehensiveTestDataset creates a comprehensive test dataset
func createComprehensiveTestDataset() *AccuracyTestDataset {
	return &AccuracyTestDataset{
		Name:        "comprehensive_rule_engine_test",
		Description: "Comprehensive test dataset for rule engine accuracy and performance validation",
		TestCases: []AccuracyTestCase{
			// Technology companies
			{
				ID:             "tech_001",
				BusinessName:   "TechInnovate Solutions",
				Description:    "Software development and AI consulting services",
				WebsiteURL:     "https://techinnovate.com",
				WebsiteContent: "We provide cutting-edge software solutions and AI consulting",
				ExpectedLabels: []string{"technology", "software_development"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"7372"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "technology",
					"risk_level": "low",
				},
			},
			{
				ID:             "tech_002",
				BusinessName:   "CloudTech Services",
				Description:    "Cloud computing and infrastructure services",
				WebsiteURL:     "https://cloudtech.com",
				WebsiteContent: "We provide cloud computing and infrastructure solutions",
				ExpectedLabels: []string{"technology", "cloud_services"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"7372"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "technology",
					"risk_level": "low",
				},
			},
			// Financial services
			{
				ID:             "finance_001",
				BusinessName:   "SecureBank Financial",
				Description:    "Commercial banking and investment services",
				WebsiteURL:     "https://securebank.com",
				WebsiteContent: "We provide commercial banking and investment services",
				ExpectedLabels: []string{"financial_services", "banking"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"6012"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "financial_services",
					"risk_level": "low",
				},
			},
			{
				ID:             "finance_002",
				BusinessName:   "Investment Partners LLC",
				Description:    "Investment advisory and wealth management",
				WebsiteURL:     "https://investment-partners.com",
				WebsiteContent: "We provide investment advisory and wealth management services",
				ExpectedLabels: []string{"financial_services", "investment"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"6012"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "financial_services",
					"risk_level": "low",
				},
			},
			// Healthcare
			{
				ID:             "healthcare_001",
				BusinessName:   "MedCare Plus",
				Description:    "Medical services and healthcare consulting",
				WebsiteURL:     "https://medcare-plus.com",
				WebsiteContent: "We provide comprehensive medical services and healthcare consulting",
				ExpectedLabels: []string{"healthcare", "medical_services"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"8011"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "healthcare",
					"risk_level": "low",
				},
			},
			// High-risk businesses
			{
				ID:             "gambling_001",
				BusinessName:   "Lucky Slots Casino",
				Description:    "Online casino and gambling services",
				WebsiteURL:     "https://luckyslots-casino.com",
				WebsiteContent: "Welcome to our online casino with slots, poker, and sports betting",
				ExpectedLabels: []string{"gambling", "entertainment"},
				ExpectedRisks:  []string{"prohibited", "high_risk"},
				ExpectedMCCs:   []string{"7995"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "gambling",
					"risk_level": "high",
				},
			},
			{
				ID:             "adult_001",
				BusinessName:   "Adult Entertainment Corp",
				Description:    "Adult entertainment and media services",
				WebsiteURL:     "https://adult-entertainment.com",
				WebsiteContent: "We provide adult entertainment and media services",
				ExpectedLabels: []string{"adult_entertainment"},
				ExpectedRisks:  []string{"prohibited", "high_risk"},
				ExpectedMCCs:   []string{"7273"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "adult_entertainment",
					"risk_level": "high",
				},
			},
			// Prohibited businesses
			{
				ID:             "illegal_001",
				BusinessName:   "Drug Store Online",
				Description:    "Online pharmaceutical and controlled substance sales",
				WebsiteURL:     "https://drugstore-online.com",
				WebsiteContent: "We sell prescription drugs and controlled substances online without prescription",
				ExpectedLabels: []string{"pharmaceuticals"},
				ExpectedRisks:  []string{"illegal", "prohibited"},
				ExpectedMCCs:   []string{"5122"},
				IsBlacklisted:  true,
				Metadata: map[string]string{
					"industry":   "pharmaceuticals",
					"risk_level": "critical",
				},
			},
			{
				ID:             "fraud_001",
				BusinessName:   "Fraudulent Services Inc",
				Description:    "Financial services and investment opportunities",
				WebsiteURL:     "https://fraudulent-services.com",
				WebsiteContent: "We provide guaranteed high-return investment opportunities",
				ExpectedLabels: []string{"financial_services"},
				ExpectedRisks:  []string{"fraud", "high_risk"},
				ExpectedMCCs:   []string{"6012"},
				IsBlacklisted:  true,
				Metadata: map[string]string{
					"industry":   "financial_services",
					"risk_level": "critical",
				},
			},
			// Complex businesses
			{
				ID:             "complex_001",
				BusinessName:   "Global Enterprise Solutions",
				Description:    "Technology consulting, financial services, and healthcare solutions",
				WebsiteURL:     "https://global-enterprise.com",
				WebsiteContent: "We provide comprehensive technology consulting, financial advisory, and healthcare technology solutions",
				ExpectedLabels: []string{"technology", "financial_services", "healthcare"},
				ExpectedRisks:  []string{},
				ExpectedMCCs:   []string{"7372", "6012", "8011"},
				IsBlacklisted:  false,
				Metadata: map[string]string{
					"industry":   "multi_industry",
					"risk_level": "low",
				},
			},
		},
		Categories: map[string][]string{
			"technology":          {"software_development", "cloud_services", "ai_consulting"},
			"financial_services":  {"banking", "investment", "wealth_management"},
			"healthcare":          {"medical_services", "healthcare_consulting"},
			"gambling":            {"online_casino", "sports_betting", "poker"},
			"adult_entertainment": {"adult_media", "adult_services"},
			"pharmaceuticals":     {"drug_distribution", "prescription_drugs"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
