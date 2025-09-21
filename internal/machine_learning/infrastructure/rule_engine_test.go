package infrastructure

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

// TestRuleEngineAccuracy tests rule engine accuracy with comprehensive test cases
func TestRuleEngineAccuracy(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Create accuracy tester
	accuracyTester := NewRuleEngineAccuracyTester(logger)

	// Load standard test dataset
	testDataset := accuracyTester.CreateStandardTestDataset()
	if err := accuracyTester.LoadTestDataset(testDataset); err != nil {
		t.Fatalf("Failed to load test dataset: %v", err)
	}

	// Run accuracy test
	metrics, err := accuracyTester.RunAccuracyTest(ctx, ruleEngine, testDataset.Name)
	if err != nil {
		t.Fatalf("Failed to run accuracy test: %v", err)
	}

	// Validate accuracy target (90%+)
	if metrics.Accuracy < 0.90 {
		t.Errorf("Accuracy %.2f%% is below target of 90%%", metrics.Accuracy*100)
	}

	// Validate performance target (sub-10ms average response time)
	if metrics.PerformanceMetrics.AverageResponseTime > 10*time.Millisecond {
		t.Errorf("Average response time %v exceeds target of 10ms", metrics.PerformanceMetrics.AverageResponseTime)
	}

	// Log results
	t.Logf("Accuracy Test Results:")
	t.Logf("  Accuracy: %.2f%%", metrics.Accuracy*100)
	t.Logf("  Precision: %.2f%%", metrics.Precision*100)
	t.Logf("  Recall: %.2f%%", metrics.Recall*100)
	t.Logf("  F1 Score: %.2f%%", metrics.F1Score*100)
	t.Logf("  Average Response Time: %v", metrics.PerformanceMetrics.AverageResponseTime)
	t.Logf("  P95 Response Time: %v", metrics.PerformanceMetrics.P95ResponseTime)
	t.Logf("  P99 Response Time: %v", metrics.PerformanceMetrics.P99ResponseTime)
	t.Logf("  Throughput: %.2f req/s", metrics.PerformanceMetrics.ThroughputPerSecond)
}

// TestRuleEnginePerformance tests rule engine performance optimization
func TestRuleEnginePerformance(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Create performance optimizer
	optimizer := NewRuleEnginePerformanceOptimizer(logger)

	// Optimize rule engine
	if err := optimizer.OptimizeRuleEngine(ruleEngine); err != nil {
		t.Fatalf("Failed to optimize rule engine: %v", err)
	}

	// Benchmark rule engine performance
	benchmarkMetrics, err := optimizer.BenchmarkRuleEnginePerformance(ruleEngine, 100)
	if err != nil {
		t.Fatalf("Failed to benchmark rule engine: %v", err)
	}

	// Validate performance targets
	if benchmarkMetrics.AverageResponseTime > 10*time.Millisecond {
		t.Errorf("Average response time %v exceeds target of 10ms", benchmarkMetrics.AverageResponseTime)
	}

	if benchmarkMetrics.P95ResponseTime > 15*time.Millisecond {
		t.Errorf("P95 response time %v exceeds target of 15ms", benchmarkMetrics.P95ResponseTime)
	}

	if benchmarkMetrics.P99ResponseTime > 20*time.Millisecond {
		t.Errorf("P99 response time %v exceeds target of 20ms", benchmarkMetrics.P99ResponseTime)
	}

	// Log performance results
	t.Logf("Performance Test Results:")
	t.Logf("  Average Response Time: %v", benchmarkMetrics.AverageResponseTime)
	t.Logf("  Min Response Time: %v", benchmarkMetrics.MinResponseTime)
	t.Logf("  Max Response Time: %v", benchmarkMetrics.MaxResponseTime)
	t.Logf("  P95 Response Time: %v", benchmarkMetrics.P95ResponseTime)
	t.Logf("  P99 Response Time: %v", benchmarkMetrics.P99ResponseTime)
	t.Logf("  Throughput: %.2f req/s", benchmarkMetrics.RequestsPerSecond)
	t.Logf("  Success Rate: %.2f%%", float64(benchmarkMetrics.SuccessfulRequests)/float64(benchmarkMetrics.TotalRequests)*100)

	// Test detailed performance report
	detailedReport := optimizer.GetDetailedPerformanceReport()
	if detailedReport == nil {
		t.Error("Detailed performance report should not be nil")
	}

	// Validate report structure
	if _, ok := detailedReport["timestamp"]; !ok {
		t.Error("Detailed report should include timestamp")
	}
	if _, ok := detailedReport["summary"]; !ok {
		t.Error("Detailed report should include summary")
	}
	if _, ok := detailedReport["metrics"]; !ok {
		t.Error("Detailed report should include metrics")
	}
	if _, ok := detailedReport["trends"]; !ok {
		t.Error("Detailed report should include trends")
	}
	if _, ok := detailedReport["alerts"]; !ok {
		t.Error("Detailed report should include alerts")
	}
	if _, ok := detailedReport["system"]; !ok {
		t.Error("Detailed report should include system info")
	}

	t.Logf("Detailed Performance Report: %+v", detailedReport)
}

// TestRuleEngineMonitoring tests performance monitoring functionality
func TestRuleEngineMonitoring(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[MONITORING_TEST] ", log.LstdFlags)

	// Create performance optimizer
	optimizer := NewRuleEnginePerformanceOptimizer(logger)

	// Test monitoring initialization
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start monitoring
	if err := optimizer.StartMonitoring(ctx); err != nil {
		t.Fatalf("Failed to start monitoring: %v", err)
	}

	// Wait a bit for monitoring to collect some data
	time.Sleep(2 * time.Second)

	// Test detailed performance report
	report := optimizer.GetDetailedPerformanceReport()
	if report == nil {
		t.Error("Performance report should not be nil")
	}

	// Validate report structure
	requiredKeys := []string{"timestamp", "summary", "metrics", "trends", "alerts", "system"}
	for _, key := range requiredKeys {
		if _, ok := report[key]; !ok {
			t.Errorf("Report should include %s", key)
		}
	}

	// Test summary structure
	summary, ok := report["summary"].(map[string]interface{})
	if !ok {
		t.Error("Summary should be a map")
	}

	// Test system info structure
	system, ok := report["system"].(map[string]interface{})
	if !ok {
		t.Error("System info should be a map")
	}

	// Validate system info contains expected fields
	systemKeys := []string{"go_version", "num_cpu", "num_goroutines", "memory_alloc_mb"}
	for _, key := range systemKeys {
		if _, ok := system[key]; !ok {
			t.Errorf("System info should include %s", key)
		}
	}

	t.Logf("Monitoring Test Results:")
	t.Logf("  Report Keys: %v", getMapKeys(report))
	t.Logf("  Summary: %+v", summary)
	t.Logf("  System Info: %+v", system)
}

// getMapKeys returns the keys of a map
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestAutomatedValidation tests automated validation and regression testing
func TestAutomatedValidation(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[AUTOMATED_VALIDATION_TEST] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("test_engine", logger)

	// Create accuracy tester
	tester := NewRuleEngineAccuracyTester(logger)

	// Create automated validation manager
	validationManager := NewAutomatedValidationManager(logger)

	// Test validation manager initialization
	if validationManager == nil {
		t.Fatal("Validation manager should not be nil")
	}

	// Test validation report generation
	report := validationManager.GetValidationReport()
	if report == nil {
		t.Error("Validation report should not be nil")
	}

	// Validate report structure
	requiredKeys := []string{"timestamp", "config", "rules", "baselines", "history"}
	for _, key := range requiredKeys {
		if _, ok := report[key]; !ok {
			t.Errorf("Report should include %s", key)
		}
	}

	// Test validation rules
	rules, ok := report["rules"].([]ValidationRule)
	if !ok {
		t.Error("Rules should be a slice of ValidationRule")
	}

	if len(rules) == 0 {
		t.Error("Should have default validation rules")
	}

	// Test configuration
	config, ok := report["config"].(AutomatedValidationConfig)
	if !ok {
		t.Error("Config should be of type AutomatedValidationConfig")
	}

	if !config.EnableRegressionTesting {
		t.Error("Regression testing should be enabled by default")
	}

	// Test manual validation
	ctx := context.Background()
	result, err := tester.RunAccuracyTest(ctx, ruleEngine, "standard_rule_engine_test")
	if err != nil {
		t.Fatalf("Accuracy test failed: %v", err)
	}

	// Test validation against rules
	validationResults := validationManager.validateAgainstRules(result)
	if len(validationResults) == 0 {
		t.Error("Should have validation results")
	}

	// Test regression detection
	regressionResult := validationManager.checkForRegressions("test_validation", result)
	if regressionResult == nil {
		t.Error("Regression result should not be nil")
	}

	// First run should establish baseline
	if regressionResult.RegressionFound {
		t.Error("First run should not detect regression")
	}

	// Test storing test result
	validationManager.storeTestResult("test_validation", result)

	// Test getting validation report after storing result
	updatedReport := validationManager.GetValidationReport()
	if updatedReport == nil {
		t.Error("Updated report should not be nil")
	}

	t.Logf("Automated Validation Test Results:")
	t.Logf("  Validation Rules: %d", len(rules))
	t.Logf("  Config: %+v", config)
	t.Logf("  Validation Results: %+v", validationResults)
	t.Logf("  Regression Result: %+v", regressionResult)
}

// TestRuleEngineKeywordMatching tests keyword matching accuracy
func TestRuleEngineKeywordMatching(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create keyword matcher
	keywordMatcher := NewKeywordMatcher(logger)

	// Initialize keyword matcher
	ctx := context.Background()
	if err := keywordMatcher.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize keyword matcher: %v", err)
	}

	// Test cases for keyword matching
	testCases := []struct {
		name           string
		businessName   string
		description    string
		expectedLabels []string
	}{
		{
			name:           "Technology Company",
			businessName:   "TechCorp Solutions",
			description:    "Software development and technology consulting",
			expectedLabels: []string{"technology", "software_development"},
		},
		{
			name:           "Financial Services",
			businessName:   "FinanceFirst Bank",
			description:    "Banking and financial services for businesses",
			expectedLabels: []string{"financial_services", "banking"},
		},
		{
			name:           "Healthcare Provider",
			businessName:   "HealthCare Plus",
			description:    "Medical services and healthcare consulting",
			expectedLabels: []string{"healthcare", "medical_services"},
		},
	}

	// Test keyword matching
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			predictions, err := keywordMatcher.ClassifyByKeywords(ctx, tc.businessName, tc.description)
			if err != nil {
				t.Fatalf("Keyword matching failed: %v", err)
			}

			if len(predictions) == 0 {
				t.Error("Expected predictions, got none")
			}

			// Check if expected labels are in predictions
			predictedLabels := make(map[string]bool)
			for _, pred := range predictions {
				predictedLabels[pred.Label] = true
			}

			for _, expectedLabel := range tc.expectedLabels {
				if !predictedLabels[expectedLabel] {
					t.Errorf("Expected label '%s' not found in predictions", expectedLabel)
				}
			}

			// Check confidence scores
			for _, pred := range predictions {
				if pred.Confidence <= 0 || pred.Confidence > 1 {
					t.Errorf("Invalid confidence score: %f", pred.Confidence)
				}
			}
		})
	}
}

// TestRuleEngineRiskDetection tests risk detection accuracy
func TestRuleEngineRiskDetection(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create keyword matcher
	keywordMatcher := NewKeywordMatcher(logger)

	// Initialize keyword matcher
	ctx := context.Background()
	if err := keywordMatcher.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize keyword matcher: %v", err)
	}

	// Test cases for risk detection
	testCases := []struct {
		name           string
		businessName   string
		description    string
		websiteContent string
		expectedRisks  []string
	}{
		{
			name:           "Low Risk Business",
			businessName:   "Safe Tech Corp",
			description:    "Legitimate software development company",
			websiteContent: "We provide secure and reliable software solutions",
			expectedRisks:  []string{},
		},
		{
			name:           "High Risk Gambling",
			businessName:   "Lucky Casino",
			description:    "Online casino and gambling services",
			websiteContent: "Welcome to our online casino with slots and poker",
			expectedRisks:  []string{"prohibited", "high_risk"},
		},
		{
			name:           "Illegal Activities",
			businessName:   "Drug Store Online",
			description:    "Pharmaceutical distribution",
			websiteContent: "We sell prescription drugs and controlled substances",
			expectedRisks:  []string{"illegal", "prohibited"},
		},
	}

	// Test risk detection
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			risks, err := keywordMatcher.DetectRiskKeywords(ctx, tc.businessName, tc.description, tc.websiteContent)
			if err != nil {
				t.Fatalf("Risk detection failed: %v", err)
			}

			// Check if expected risks are detected
			detectedCategories := make(map[string]bool)
			for _, risk := range risks {
				detectedCategories[risk.Category] = true
			}

			for _, expectedRisk := range tc.expectedRisks {
				if !detectedCategories[expectedRisk] {
					t.Errorf("Expected risk category '%s' not detected", expectedRisk)
				}
			}

			// Check risk confidence scores
			for _, risk := range risks {
				if risk.Confidence <= 0 || risk.Confidence > 1 {
					t.Errorf("Invalid risk confidence score: %f", risk.Confidence)
				}
			}
		})
	}
}

// TestRuleEngineMCCLookup tests MCC code lookup functionality
func TestRuleEngineMCCLookup(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create MCC code lookup
	mccLookup := NewMCCCodeLookup(logger)

	// Initialize MCC code lookup
	ctx := context.Background()
	if err := mccLookup.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize MCC code lookup: %v", err)
	}

	// Test cases for MCC lookup
	testCases := []struct {
		name               string
		businessName       string
		description        string
		expectedMCCs       []string
		expectedProhibited bool
	}{
		{
			name:               "Software Development",
			businessName:       "Software Solutions Inc",
			description:        "Custom software development and consulting",
			expectedMCCs:       []string{"7372"},
			expectedProhibited: false,
		},
		{
			name:               "Gambling Services",
			businessName:       "Online Casino",
			description:        "Online gambling and casino services",
			expectedMCCs:       []string{"7995"},
			expectedProhibited: true,
		},
		{
			name:               "Financial Services",
			businessName:       "Business Bank",
			description:        "Commercial banking and financial services",
			expectedMCCs:       []string{"6012"},
			expectedProhibited: false,
		},
	}

	// Test MCC lookup
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test classification by MCC
			classifications, err := mccLookup.ClassifyByMCC(ctx, tc.businessName, tc.description)
			if err != nil {
				t.Fatalf("MCC classification failed: %v", err)
			}

			if len(classifications) == 0 {
				t.Error("Expected MCC classifications, got none")
			}

			// Check if expected MCCs are found
			foundMCCs := make(map[string]bool)
			for _, classification := range classifications {
				foundMCCs[classification.Label] = true
			}

			for _, expectedMCC := range tc.expectedMCCs {
				if !foundMCCs[expectedMCC] {
					t.Errorf("Expected MCC '%s' not found in classifications", expectedMCC)
				}
			}

			// Test MCC restrictions
			restrictions, err := mccLookup.CheckMCCRestrictions(ctx, tc.businessName, tc.description)
			if err != nil {
				t.Fatalf("MCC restriction check failed: %v", err)
			}

			hasProhibited := len(restrictions) > 0
			if hasProhibited != tc.expectedProhibited {
				t.Errorf("Expected prohibited status %v, got %v", tc.expectedProhibited, hasProhibited)
			}
		})
	}
}

// TestRuleEngineBlacklistChecker tests blacklist checking functionality
func TestRuleEngineBlacklistChecker(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create blacklist checker
	blacklistChecker := NewBlacklistChecker(logger)

	// Initialize blacklist checker
	ctx := context.Background()
	if err := blacklistChecker.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize blacklist checker: %v", err)
	}

	// Test cases for blacklist checking
	testCases := []struct {
		name                string
		businessName        string
		websiteURL          string
		expectedBlacklisted bool
	}{
		{
			name:                "Legitimate Business",
			businessName:        "Clean Business Corp",
			websiteURL:          "https://cleanbusiness.com",
			expectedBlacklisted: false,
		},
		{
			name:                "Known Bad Actor",
			businessName:        "Fraudulent Services",
			websiteURL:          "https://fraudulent.com",
			expectedBlacklisted: true,
		},
		{
			name:                "Suspicious Domain",
			businessName:        "Questionable Corp",
			websiteURL:          "https://suspicious-domain.com",
			expectedBlacklisted: true,
		},
	}

	// Test blacklist checking
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			risks, err := blacklistChecker.CheckBlacklist(ctx, tc.businessName, tc.websiteURL)
			if err != nil {
				t.Fatalf("Blacklist check failed: %v", err)
			}

			isBlacklisted := len(risks) > 0
			if isBlacklisted != tc.expectedBlacklisted {
				t.Errorf("Expected blacklisted status %v, got %v", tc.expectedBlacklisted, isBlacklisted)
			}

			// Check risk details if blacklisted
			if isBlacklisted {
				for _, risk := range risks {
					if risk.Category != "blacklist" {
						t.Errorf("Expected blacklist category, got %s", risk.Category)
					}
					if risk.Severity != "critical" {
						t.Errorf("Expected critical severity for blacklisted entity, got %s", risk.Severity)
					}
				}
			}
		})
	}
}

// TestRuleEngineCache tests caching functionality
func TestRuleEngineCache(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create cache
	cache := NewRuleEngineCache(logger)

	// Initialize cache
	ctx := context.Background()
	if err := cache.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test cache operations
	testKey := "test_key"
	testResult := &RuleEngineClassificationResponse{
		RequestID: "test_request",
		Classifications: []ClassificationPrediction{
			{Label: "technology", Confidence: 0.95, Probability: 0.95, Rank: 1},
		},
		Confidence:     0.95,
		ProcessingTime: 5 * time.Millisecond,
		Timestamp:      time.Now(),
		Success:        true,
		Method:         "keyword_matching",
	}

	// Test cache set and get
	cache.SetClassification(testKey, testResult, 1*time.Hour)

	cached, found := cache.GetClassification(testKey)
	if !found {
		t.Error("Expected to find cached result")
	}

	if cached.Result.RequestID != testResult.RequestID {
		t.Errorf("Expected request ID %s, got %s", testResult.RequestID, cached.Result.RequestID)
	}

	// Test cache expiration
	cache.SetClassification("expired_key", testResult, 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond)

	_, found = cache.GetClassification("expired_key")
	if found {
		t.Error("Expected expired cache entry to not be found")
	}

	// Test cache statistics
	stats := cache.GetStats()
	if stats.TotalEntries == 0 {
		t.Error("Expected cache to have entries")
	}
}

// TestRuleEngineConcurrentAccess tests concurrent access to rule engine
func TestRuleEngineConcurrentAccess(t *testing.T) {
	// Create test logger
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Test concurrent requests
	numGoroutines := 10
	numRequestsPerGoroutine := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numRequestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numRequestsPerGoroutine; j++ {
				req := &RuleEngineClassificationRequest{
					BusinessName: "Concurrent Test Business",
					Description:  "Testing concurrent access to rule engine",
					WebsiteURL:   "https://concurrent-test.com",
				}

				_, err := ruleEngine.Classify(ctx, req)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Expected no errors, got %d errors", errorCount)
	}
}

// BenchmarkRuleEngineClassification benchmarks classification performance
func BenchmarkRuleEngineClassification(b *testing.B) {
	// Create test logger
	logger := log.New(log.Writer(), "[BENCHMARK] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		b.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Create test request
	req := &RuleEngineClassificationRequest{
		BusinessName: "Benchmark Technology Corp",
		Description:  "High-performance software development and consulting services",
		WebsiteURL:   "https://benchmark-tech.com",
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ruleEngine.Classify(ctx, req)
		if err != nil {
			b.Fatalf("Classification failed: %v", err)
		}
	}
}

// BenchmarkRuleEngineRiskDetection benchmarks risk detection performance
func BenchmarkRuleEngineRiskDetection(b *testing.B) {
	// Create test logger
	logger := log.New(log.Writer(), "[BENCHMARK] ", log.LstdFlags)

	// Create rule engine
	ruleEngine := NewGoRuleEngine("localhost:8080", logger)

	// Initialize rule engine
	ctx := context.Background()
	if err := ruleEngine.Initialize(ctx); err != nil {
		b.Fatalf("Failed to initialize rule engine: %v", err)
	}

	// Create test request
	req := &RuleEngineRiskRequest{
		BusinessName:   "Benchmark Risk Assessment Corp",
		Description:    "Risk assessment and compliance services",
		WebsiteURL:     "https://benchmark-risk.com",
		WebsiteContent: "We provide comprehensive risk assessment and compliance monitoring services",
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ruleEngine.DetectRisk(ctx, req)
		if err != nil {
			b.Fatalf("Risk detection failed: %v", err)
		}
	}
}
