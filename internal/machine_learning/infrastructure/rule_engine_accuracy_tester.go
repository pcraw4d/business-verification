package infrastructure

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// RuleEngineAccuracyTester provides comprehensive accuracy testing for rule-based systems
type RuleEngineAccuracyTester struct {
	// Test datasets
	testDatasets map[string]*AccuracyTestDataset

	// Performance metrics
	accuracyMetrics map[string]*AccuracyMetrics

	// Configuration
	config AccuracyTestConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// AccuracyTestDataset represents a test dataset for accuracy testing
type AccuracyTestDataset struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	TestCases   []AccuracyTestCase  `json:"test_cases"`
	Categories  map[string][]string `json:"categories"` // category -> expected labels
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// AccuracyTestCase represents a single test case
type AccuracyTestCase struct {
	ID             string            `json:"id"`
	BusinessName   string            `json:"business_name"`
	Description    string            `json:"description"`
	WebsiteURL     string            `json:"website_url"`
	WebsiteContent string            `json:"website_content"`
	ExpectedLabels []string          `json:"expected_labels"` // Expected classification labels
	ExpectedRisks  []string          `json:"expected_risks"`  // Expected risk categories
	ExpectedMCCs   []string          `json:"expected_mccs"`   // Expected MCC codes
	IsBlacklisted  bool              `json:"is_blacklisted"`  // Expected blacklist status
	Metadata       map[string]string `json:"metadata"`
}

// AccuracyMetrics holds comprehensive accuracy metrics
type AccuracyMetrics struct {
	DatasetName        string                     `json:"dataset_name"`
	TotalTestCases     int                        `json:"total_test_cases"`
	CorrectPredictions int                        `json:"correct_predictions"`
	Accuracy           float64                    `json:"accuracy"`
	Precision          float64                    `json:"precision"`
	Recall             float64                    `json:"recall"`
	F1Score            float64                    `json:"f1_score"`
	PerCategoryMetrics map[string]CategoryMetrics `json:"per_category_metrics"`
	ConfusionMatrix    map[string]map[string]int  `json:"confusion_matrix"`
	ErrorAnalysis      []ErrorAnalysis            `json:"error_analysis"`
	PerformanceMetrics AccuracyPerformanceMetrics `json:"performance_metrics"`
	TestedAt           time.Time                  `json:"tested_at"`
}

// CategoryMetrics holds metrics for a specific category
type CategoryMetrics struct {
	Category       string  `json:"category"`
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	FalseNegatives int     `json:"false_negatives"`
	TrueNegatives  int     `json:"true_negatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1Score        float64 `json:"f1_score"`
	Support        int     `json:"support"`
}

// ErrorAnalysis holds analysis of prediction errors
type ErrorAnalysis struct {
	TestCaseID      string        `json:"test_case_id"`
	BusinessName    string        `json:"business_name"`
	ExpectedLabels  []string      `json:"expected_labels"`
	PredictedLabels []string      `json:"predicted_labels"`
	ErrorType       string        `json:"error_type"` // false_positive, false_negative, misclassification
	Confidence      float64       `json:"confidence"`
	ProcessingTime  time.Duration `json:"processing_time"`
}

// AccuracyPerformanceMetrics holds performance-related metrics for accuracy testing
type AccuracyPerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	MemoryUsageMB       float64       `json:"memory_usage_mb"`
	CPUUsagePercent     float64       `json:"cpu_usage_percent"`
}

// AccuracyTestConfig holds configuration for accuracy testing
type AccuracyTestConfig struct {
	TargetAccuracy           float64       `json:"target_accuracy"`   // 0.90 (90%)
	MaxResponseTime          time.Duration `json:"max_response_time"` // 10ms
	TestTimeout              time.Duration `json:"test_timeout"`
	ConcurrentTests          int           `json:"concurrent_tests"`
	EnableErrorAnalysis      bool          `json:"enable_error_analysis"`
	EnablePerformanceMetrics bool          `json:"enable_performance_metrics"`
}

// NewRuleEngineAccuracyTester creates a new rule engine accuracy tester
func NewRuleEngineAccuracyTester(logger *log.Logger) *RuleEngineAccuracyTester {
	if logger == nil {
		logger = log.Default()
	}

	return &RuleEngineAccuracyTester{
		testDatasets:    make(map[string]*AccuracyTestDataset),
		accuracyMetrics: make(map[string]*AccuracyMetrics),
		config: AccuracyTestConfig{
			TargetAccuracy:           0.90, // 90%
			MaxResponseTime:          10 * time.Millisecond,
			TestTimeout:              30 * time.Second,
			ConcurrentTests:          10,
			EnableErrorAnalysis:      true,
			EnablePerformanceMetrics: true,
		},
		logger: logger,
	}
}

// LoadTestDataset loads a test dataset for accuracy testing
func (reat *RuleEngineAccuracyTester) LoadTestDataset(dataset *AccuracyTestDataset) error {
	reat.mu.Lock()
	defer reat.mu.Unlock()

	reat.logger.Printf("📊 Loading test dataset: %s with %d test cases", dataset.Name, len(dataset.TestCases))
	reat.testDatasets[dataset.Name] = dataset

	return nil
}

// RunAccuracyTest runs comprehensive accuracy testing on the rule engine
func (reat *RuleEngineAccuracyTester) RunAccuracyTest(ctx context.Context, ruleEngine *GoRuleEngine, datasetName string) (*AccuracyMetrics, error) {
	reat.mu.RLock()
	dataset, exists := reat.testDatasets[datasetName]
	reat.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("test dataset '%s' not found", datasetName)
	}

	reat.logger.Printf("🧪 Running accuracy test on dataset: %s", datasetName)

	// Initialize metrics
	metrics := &AccuracyMetrics{
		DatasetName:        datasetName,
		TotalTestCases:     len(dataset.TestCases),
		PerCategoryMetrics: make(map[string]CategoryMetrics),
		ConfusionMatrix:    make(map[string]map[string]int),
		ErrorAnalysis:      []ErrorAnalysis{},
		TestedAt:           time.Now(),
	}

	// Track performance metrics
	var responseTimes []time.Duration
	startTime := time.Now()

	// Run tests concurrently with controlled concurrency
	semaphore := make(chan struct{}, reat.config.ConcurrentTests)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, testCase := range dataset.TestCases {
		wg.Add(1)
		go func(tc AccuracyTestCase) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			// Run classification test
			classificationResult, classificationTime := reat.runClassificationTest(ctx, ruleEngine, tc)

			// Run risk detection test
			riskResult, riskTime := reat.runRiskDetectionTest(ctx, ruleEngine, tc)

			// Run blacklist test
			blacklistResult, blacklistTime := reat.runBlacklistTest(ctx, ruleEngine, tc)

			// Calculate total response time
			totalTime := classificationTime + riskTime + blacklistTime

			mu.Lock()
			responseTimes = append(responseTimes, totalTime)

			// Analyze results
			reat.analyzeClassificationResults(metrics, tc, classificationResult)
			reat.analyzeRiskDetectionResults(metrics, tc, riskResult)
			reat.analyzeBlacklistResults(metrics, tc, blacklistResult)

			// Add error analysis if enabled
			if reat.config.EnableErrorAnalysis {
				reat.addErrorAnalysis(metrics, tc, classificationResult, riskResult, blacklistResult, totalTime)
			}
			mu.Unlock()

		}(testCase)
	}

	wg.Wait()

	// Calculate final metrics
	reat.calculateFinalMetrics(metrics, responseTimes, time.Since(startTime))

	// Store metrics
	reat.mu.Lock()
	reat.accuracyMetrics[datasetName] = metrics
	reat.mu.Unlock()

	reat.logger.Printf("✅ Accuracy test completed for dataset: %s - Accuracy: %.2f%%, Avg Response Time: %v",
		datasetName, metrics.Accuracy*100, metrics.PerformanceMetrics.AverageResponseTime)

	return metrics, nil
}

// runClassificationTest runs a classification test case
func (reat *RuleEngineAccuracyTester) runClassificationTest(ctx context.Context, ruleEngine *GoRuleEngine, testCase AccuracyTestCase) (*RuleEngineClassificationResponse, time.Duration) {
	start := time.Now()

	req := &RuleEngineClassificationRequest{
		BusinessName: testCase.BusinessName,
		Description:  testCase.Description,
		WebsiteURL:   testCase.WebsiteURL,
	}

	result, err := ruleEngine.Classify(ctx, req)
	if err != nil {
		reat.logger.Printf("⚠️ Classification test failed for case %s: %v", testCase.ID, err)
		return nil, time.Since(start)
	}

	return result, time.Since(start)
}

// runRiskDetectionTest runs a risk detection test case
func (reat *RuleEngineAccuracyTester) runRiskDetectionTest(ctx context.Context, ruleEngine *GoRuleEngine, testCase AccuracyTestCase) (*RuleEngineRiskResponse, time.Duration) {
	start := time.Now()

	req := &RuleEngineRiskRequest{
		BusinessName:   testCase.BusinessName,
		Description:    testCase.Description,
		WebsiteURL:     testCase.WebsiteURL,
		WebsiteContent: testCase.WebsiteContent,
	}

	result, err := ruleEngine.DetectRisk(ctx, req)
	if err != nil {
		reat.logger.Printf("⚠️ Risk detection test failed for case %s: %v", testCase.ID, err)
		return nil, time.Since(start)
	}

	return result, time.Since(start)
}

// runBlacklistTest runs a blacklist test case
func (reat *RuleEngineAccuracyTester) runBlacklistTest(ctx context.Context, ruleEngine *GoRuleEngine, testCase AccuracyTestCase) (bool, time.Duration) {
	start := time.Now()

	// This would be implemented as part of the blacklist checker
	// For now, we'll simulate the test
	isBlacklisted := false // This would be determined by the actual blacklist checker

	return isBlacklisted, time.Since(start)
}

// analyzeClassificationResults analyzes classification test results
func (reat *RuleEngineAccuracyTester) analyzeClassificationResults(metrics *AccuracyMetrics, testCase AccuracyTestCase, result *RuleEngineClassificationResponse) {
	if result == nil {
		return
	}

	// Extract predicted labels
	var predictedLabels []string
	for _, prediction := range result.Classifications {
		predictedLabels = append(predictedLabels, prediction.Label)
	}

	// Check for correct predictions
	correct := reat.checkLabelMatch(testCase.ExpectedLabels, predictedLabels)
	if correct {
		metrics.CorrectPredictions++
	}

	// Update confusion matrix
	reat.updateConfusionMatrix(metrics, testCase.ExpectedLabels, predictedLabels)
}

// analyzeRiskDetectionResults analyzes risk detection test results
func (reat *RuleEngineAccuracyTester) analyzeRiskDetectionResults(metrics *AccuracyMetrics, testCase AccuracyTestCase, result *RuleEngineRiskResponse) {
	if result == nil {
		return
	}

	// Extract predicted risk categories
	var predictedRisks []string
	for _, risk := range result.DetectedRisks {
		predictedRisks = append(predictedRisks, risk.Category)
	}

	// Check for correct risk predictions
	correct := reat.checkLabelMatch(testCase.ExpectedRisks, predictedRisks)
	if correct {
		// This would be tracked separately for risk detection accuracy
	}
}

// analyzeBlacklistResults analyzes blacklist test results
func (reat *RuleEngineAccuracyTester) analyzeBlacklistResults(metrics *AccuracyMetrics, testCase AccuracyTestCase, isBlacklisted bool) {
	// Check if blacklist prediction matches expected
	if isBlacklisted == testCase.IsBlacklisted {
		// Correct blacklist prediction
	}
}

// checkLabelMatch checks if predicted labels match expected labels
func (reat *RuleEngineAccuracyTester) checkLabelMatch(expected, predicted []string) bool {
	if len(expected) == 0 && len(predicted) == 0 {
		return true
	}

	// Convert to sets for comparison
	expectedSet := make(map[string]bool)
	for _, label := range expected {
		expectedSet[label] = true
	}

	predictedSet := make(map[string]bool)
	for _, label := range predicted {
		predictedSet[label] = true
	}

	// Check if sets are equal
	if len(expectedSet) != len(predictedSet) {
		return false
	}

	for label := range expectedSet {
		if !predictedSet[label] {
			return false
		}
	}

	return true
}

// updateConfusionMatrix updates the confusion matrix
func (reat *RuleEngineAccuracyTester) updateConfusionMatrix(metrics *AccuracyMetrics, expected, predicted []string) {
	// Initialize confusion matrix if needed
	if metrics.ConfusionMatrix == nil {
		metrics.ConfusionMatrix = make(map[string]map[string]int)
	}

	// Update matrix for each expected-predicted pair
	for _, exp := range expected {
		if metrics.ConfusionMatrix[exp] == nil {
			metrics.ConfusionMatrix[exp] = make(map[string]int)
		}
		for _, pred := range predicted {
			metrics.ConfusionMatrix[exp][pred]++
		}
	}
}

// addErrorAnalysis adds error analysis for a test case
func (reat *RuleEngineAccuracyTester) addErrorAnalysis(metrics *AccuracyMetrics, testCase AccuracyTestCase,
	classificationResult *RuleEngineClassificationResponse, riskResult *RuleEngineRiskResponse,
	blacklistResult bool, processingTime time.Duration) {

	var predictedLabels []string
	if classificationResult != nil {
		for _, prediction := range classificationResult.Classifications {
			predictedLabels = append(predictedLabels, prediction.Label)
		}
	}

	// Determine error type
	errorType := "none"
	if !reat.checkLabelMatch(testCase.ExpectedLabels, predictedLabels) {
		if len(predictedLabels) > len(testCase.ExpectedLabels) {
			errorType = "false_positive"
		} else if len(predictedLabels) < len(testCase.ExpectedLabels) {
			errorType = "false_negative"
		} else {
			errorType = "misclassification"
		}
	}

	if errorType != "none" {
		errorAnalysis := ErrorAnalysis{
			TestCaseID:      testCase.ID,
			BusinessName:    testCase.BusinessName,
			ExpectedLabels:  testCase.ExpectedLabels,
			PredictedLabels: predictedLabels,
			ErrorType:       errorType,
			Confidence:      reat.getConfidence(classificationResult),
			ProcessingTime:  processingTime,
		}
		metrics.ErrorAnalysis = append(metrics.ErrorAnalysis, errorAnalysis)
	}
}

// getConfidence extracts confidence from classification result
func (reat *RuleEngineAccuracyTester) getConfidence(result *RuleEngineClassificationResponse) float64 {
	if result == nil || len(result.Classifications) == 0 {
		return 0.0
	}
	return result.Confidence
}

// calculateFinalMetrics calculates final accuracy metrics
func (reat *RuleEngineAccuracyTester) calculateFinalMetrics(metrics *AccuracyMetrics, responseTimes []time.Duration, totalTime time.Duration) {
	// Calculate accuracy
	metrics.Accuracy = float64(metrics.CorrectPredictions) / float64(metrics.TotalTestCases)

	// Calculate precision, recall, and F1 score
	reat.calculatePrecisionRecallF1(metrics)

	// Calculate performance metrics
	reat.calculatePerformanceMetrics(metrics, responseTimes, totalTime)

	// Calculate per-category metrics
	reat.calculatePerCategoryMetrics(metrics)
}

// calculatePrecisionRecallF1 calculates precision, recall, and F1 score
func (reat *RuleEngineAccuracyTester) calculatePrecisionRecallF1(metrics *AccuracyMetrics) {
	// This is a simplified calculation - in practice, you'd calculate these per category
	// and then average them appropriately

	totalTruePositives := 0
	totalFalsePositives := 0
	totalFalseNegatives := 0

	for _, categoryMetrics := range metrics.PerCategoryMetrics {
		totalTruePositives += categoryMetrics.TruePositives
		totalFalsePositives += categoryMetrics.FalsePositives
		totalFalseNegatives += categoryMetrics.FalseNegatives
	}

	if totalTruePositives+totalFalsePositives > 0 {
		metrics.Precision = float64(totalTruePositives) / float64(totalTruePositives+totalFalsePositives)
	}

	if totalTruePositives+totalFalseNegatives > 0 {
		metrics.Recall = float64(totalTruePositives) / float64(totalTruePositives+totalFalseNegatives)
	}

	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
	}
}

// calculatePerformanceMetrics calculates performance-related metrics
func (reat *RuleEngineAccuracyTester) calculatePerformanceMetrics(metrics *AccuracyMetrics, responseTimes []time.Duration, totalTime time.Duration) {
	if len(responseTimes) == 0 {
		return
	}

	// Sort response times for percentile calculations
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	metrics.PerformanceMetrics = AccuracyPerformanceMetrics{
		AverageResponseTime: reat.calculateAverage(responseTimes),
		MinResponseTime:     responseTimes[0],
		MaxResponseTime:     responseTimes[len(responseTimes)-1],
		P95ResponseTime:     reat.calculatePercentile(responseTimes, 95),
		P99ResponseTime:     reat.calculatePercentile(responseTimes, 99),
		ThroughputPerSecond: float64(len(responseTimes)) / totalTime.Seconds(),
		// Memory and CPU usage would be measured separately
		MemoryUsageMB:   0.0, // Placeholder
		CPUUsagePercent: 0.0, // Placeholder
	}
}

// calculatePerCategoryMetrics calculates metrics for each category
func (reat *RuleEngineAccuracyTester) calculatePerCategoryMetrics(metrics *AccuracyMetrics) {
	// This would be implemented based on the confusion matrix
	// For now, we'll create placeholder metrics
}

// calculateAverage calculates the average of a slice of durations
func (reat *RuleEngineAccuracyTester) calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

// calculatePercentile calculates the nth percentile of a sorted slice of durations
func (reat *RuleEngineAccuracyTester) calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	index := int(math.Ceil(float64(len(durations)) * float64(percentile) / 100.0))
	if index >= len(durations) {
		index = len(durations) - 1
	}
	return durations[index]
}

// ValidateAccuracyTarget validates if the accuracy meets the target
func (reat *RuleEngineAccuracyTester) ValidateAccuracyTarget(metrics *AccuracyMetrics) bool {
	return metrics.Accuracy >= reat.config.TargetAccuracy
}

// ValidatePerformanceTarget validates if the performance meets the target
func (reat *RuleEngineAccuracyTester) ValidatePerformanceTarget(metrics *AccuracyMetrics) bool {
	return metrics.PerformanceMetrics.AverageResponseTime <= reat.config.MaxResponseTime
}

// GetAccuracyReport generates a comprehensive accuracy report
func (reat *RuleEngineAccuracyTester) GetAccuracyReport(datasetName string) (*AccuracyMetrics, error) {
	reat.mu.RLock()
	defer reat.mu.RUnlock()

	metrics, exists := reat.accuracyMetrics[datasetName]
	if !exists {
		return nil, fmt.Errorf("accuracy metrics for dataset '%s' not found", datasetName)
	}

	return metrics, nil
}

// CreateStandardTestDataset creates a standard test dataset for rule engine testing
func (reat *RuleEngineAccuracyTester) CreateStandardTestDataset() *AccuracyTestDataset {
	return &AccuracyTestDataset{
		Name:        "standard_rule_engine_test",
		Description: "Standard test dataset for rule engine accuracy validation",
		TestCases: []AccuracyTestCase{
			{
				ID:             "test_001",
				BusinessName:   "Acme Technology Corp",
				Description:    "Software development and technology consulting services",
				WebsiteURL:     "https://acmetech.com",
				WebsiteContent: "We provide cutting-edge software solutions and technology consulting",
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
				ID:             "test_002",
				BusinessName:   "High Risk Casino",
				Description:    "Online gambling and casino services",
				WebsiteURL:     "https://highriskcasino.com",
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
				ID:             "test_003",
				BusinessName:   "Illegal Drug Store",
				Description:    "Pharmaceutical and drug distribution",
				WebsiteURL:     "https://illegaldrugs.com",
				WebsiteContent: "We sell prescription drugs and controlled substances online",
				ExpectedLabels: []string{"pharmaceuticals"},
				ExpectedRisks:  []string{"illegal", "prohibited"},
				ExpectedMCCs:   []string{"5122"},
				IsBlacklisted:  true,
				Metadata: map[string]string{
					"industry":   "pharmaceuticals",
					"risk_level": "critical",
				},
			},
			// Add more test cases as needed
		},
		Categories: map[string][]string{
			"technology":      {"software_development", "technology_consulting", "it_services"},
			"gambling":        {"online_casino", "sports_betting", "poker"},
			"pharmaceuticals": {"drug_distribution", "prescription_drugs"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AutomatedValidationManager manages automated accuracy validation and regression testing
type AutomatedValidationManager struct {
	// Test history for regression detection
	testHistory map[string][]AccuracyMetrics

	// Baseline metrics for comparison
	baselineMetrics map[string]*AccuracyMetrics

	// Validation rules
	validationRules []ValidationRule

	// Configuration
	config AutomatedValidationConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// ValidationRule defines a validation rule for automated testing
type ValidationRule struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Metric      string  `json:"metric"`    // e.g., "precision", "recall", "f1_score"
	Threshold   float64 `json:"threshold"` // Minimum acceptable value
	Severity    string  `json:"severity"`  // "warning", "error", "critical"
	Enabled     bool    `json:"enabled"`
}

// AutomatedValidationConfig holds configuration for automated validation
type AutomatedValidationConfig struct {
	EnableRegressionTesting bool          `json:"enable_regression_testing"`
	BaselineRetentionDays   int           `json:"baseline_retention_days"` // 30 days
	ValidationInterval      time.Duration `json:"validation_interval"`     // 1 hour
	RegressionThreshold     float64       `json:"regression_threshold"`    // 5% decrease
	AlertOnRegression       bool          `json:"alert_on_regression"`
	AutoRollbackOnFailure   bool          `json:"auto_rollback_on_failure"`
}

// RegressionTestResult represents the result of a regression test
type RegressionTestResult struct {
	TestID            string             `json:"test_id"`
	Timestamp         time.Time          `json:"timestamp"`
	CurrentMetrics    *AccuracyMetrics   `json:"current_metrics"`
	BaselineMetrics   *AccuracyMetrics   `json:"baseline_metrics"`
	RegressionFound   bool               `json:"regression_found"`
	RegressionDetails []RegressionDetail `json:"regression_details"`
	Severity          string             `json:"severity"`
	Recommendation    string             `json:"recommendation"`
}

// RegressionDetail provides details about a specific regression
type RegressionDetail struct {
	Metric        string  `json:"metric"`
	CurrentValue  float64 `json:"current_value"`
	BaselineValue float64 `json:"baseline_value"`
	ChangePercent float64 `json:"change_percent"`
	Threshold     float64 `json:"threshold"`
	Exceeded      bool    `json:"exceeded"`
}

// NewAutomatedValidationManager creates a new automated validation manager
func NewAutomatedValidationManager(logger *log.Logger) *AutomatedValidationManager {
	if logger == nil {
		logger = log.Default()
	}

	manager := &AutomatedValidationManager{
		testHistory:     make(map[string][]AccuracyMetrics),
		baselineMetrics: make(map[string]*AccuracyMetrics),
		validationRules: getDefaultValidationRules(),
		config: AutomatedValidationConfig{
			EnableRegressionTesting: true,
			BaselineRetentionDays:   30,
			ValidationInterval:      1 * time.Hour,
			RegressionThreshold:     0.05, // 5%
			AlertOnRegression:       true,
			AutoRollbackOnFailure:   false,
		},
		logger: logger,
	}

	return manager
}

// getDefaultValidationRules returns default validation rules
func getDefaultValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Name:        "Minimum Precision",
			Description: "Ensures minimum precision threshold is met",
			Metric:      "precision",
			Threshold:   0.85, // 85%
			Severity:    "error",
			Enabled:     true,
		},
		{
			Name:        "Minimum Recall",
			Description: "Ensures minimum recall threshold is met",
			Metric:      "recall",
			Threshold:   0.80, // 80%
			Severity:    "error",
			Enabled:     true,
		},
		{
			Name:        "Minimum F1 Score",
			Description: "Ensures minimum F1 score threshold is met",
			Metric:      "f1_score",
			Threshold:   0.82, // 82%
			Severity:    "error",
			Enabled:     true,
		},
		{
			Name:        "Maximum Error Rate",
			Description: "Ensures error rate stays below threshold",
			Metric:      "error_rate",
			Threshold:   0.10, // 10%
			Severity:    "warning",
			Enabled:     true,
		},
	}
}

// StartAutomatedValidation starts the automated validation process
func (avm *AutomatedValidationManager) StartAutomatedValidation(ctx context.Context, tester *RuleEngineAccuracyTester, ruleEngine *GoRuleEngine) error {
	avm.logger.Printf("🤖 Starting automated validation system")

	// Start validation loop
	go avm.validationLoop(ctx, tester, ruleEngine)

	// Start regression testing loop
	if avm.config.EnableRegressionTesting {
		go avm.regressionTestingLoop(ctx, tester, ruleEngine)
	}

	return nil
}

// validationLoop runs the automated validation loop
func (avm *AutomatedValidationManager) validationLoop(ctx context.Context, tester *RuleEngineAccuracyTester, ruleEngine *GoRuleEngine) {
	ticker := time.NewTicker(avm.config.ValidationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			avm.logger.Printf("🤖 Automated validation loop stopped")
			return
		case <-ticker.C:
			avm.performAutomatedValidation(tester, ruleEngine)
		}
	}
}

// regressionTestingLoop runs the regression testing loop
func (avm *AutomatedValidationManager) regressionTestingLoop(ctx context.Context, tester *RuleEngineAccuracyTester, ruleEngine *GoRuleEngine) {
	ticker := time.NewTicker(avm.config.ValidationInterval * 2) // Run less frequently
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			avm.logger.Printf("📉 Regression testing loop stopped")
			return
		case <-ticker.C:
			avm.performRegressionTesting(tester, ruleEngine)
		}
	}
}

// performAutomatedValidation performs automated accuracy validation
func (avm *AutomatedValidationManager) performAutomatedValidation(tester *RuleEngineAccuracyTester, ruleEngine *GoRuleEngine) {
	avm.logger.Printf("🔍 Performing automated accuracy validation")

	// Run accuracy test
	result, err := tester.RunAccuracyTest(context.Background(), ruleEngine, "standard_rule_engine_test")
	if err != nil {
		avm.logger.Printf("❌ Automated validation failed: %v", err)
		return
	}

	// Validate against rules
	validationResults := avm.validateAgainstRules(result)

	// Store test result
	avm.storeTestResult("automated_validation", result)

	// Check for violations
	violations := avm.checkValidationViolations(validationResults)

	if len(violations) > 0 {
		avm.handleValidationViolations(violations, result)
	} else {
		avm.logger.Printf("✅ Automated validation passed - All rules satisfied")
	}
}

// performRegressionTesting performs regression testing
func (avm *AutomatedValidationManager) performRegressionTesting(tester *RuleEngineAccuracyTester, ruleEngine *GoRuleEngine) {
	avm.logger.Printf("📉 Performing regression testing")

	// Run accuracy test
	result, err := tester.RunAccuracyTest(context.Background(), ruleEngine, "standard_rule_engine_test")
	if err != nil {
		avm.logger.Printf("❌ Regression testing failed: %v", err)
		return
	}

	// Check for regressions
	regressionResult := avm.checkForRegressions("standard_rule_engine_test", result)

	if regressionResult.RegressionFound {
		avm.handleRegression(regressionResult)
	} else {
		avm.logger.Printf("✅ No regressions detected")
	}
}

// validateAgainstRules validates test results against validation rules
func (avm *AutomatedValidationManager) validateAgainstRules(result *AccuracyMetrics) map[string]bool {
	validationResults := make(map[string]bool)

	for _, rule := range avm.validationRules {
		if !rule.Enabled {
			continue
		}

		passed := avm.validateRule(rule, result)
		validationResults[rule.Name] = passed

		status := "✅ PASSED"
		if !passed {
			status = "❌ FAILED"
		}

		avm.logger.Printf("  %s %s: %.2f (threshold: %.2f)",
			status, rule.Name, avm.getMetricValue(rule.Metric, result), rule.Threshold)
	}

	return validationResults
}

// validateRule validates a single rule
func (avm *AutomatedValidationManager) validateRule(rule ValidationRule, result *AccuracyMetrics) bool {
	value := avm.getMetricValue(rule.Metric, result)

	// For error rate, lower is better
	if rule.Metric == "error_rate" {
		return value <= rule.Threshold
	}

	// For other metrics, higher is better
	return value >= rule.Threshold
}

// getMetricValue gets the value of a specific metric from test results
func (avm *AutomatedValidationManager) getMetricValue(metric string, result *AccuracyMetrics) float64 {
	switch metric {
	case "precision":
		return result.Precision
	case "recall":
		return result.Recall
	case "f1_score":
		return result.F1Score
	case "error_rate":
		return 1.0 - result.Precision // Simplified error rate
	default:
		return 0.0
	}
}

// checkValidationViolations checks for validation rule violations
func (avm *AutomatedValidationManager) checkValidationViolations(validationResults map[string]bool) []ValidationRule {
	var violations []ValidationRule

	for _, rule := range avm.validationRules {
		if rule.Enabled && !validationResults[rule.Name] {
			violations = append(violations, rule)
		}
	}

	return violations
}

// handleValidationViolations handles validation rule violations
func (avm *AutomatedValidationManager) handleValidationViolations(violations []ValidationRule, result *AccuracyMetrics) {
	avm.logger.Printf("🚨 Validation violations detected:")

	for _, violation := range violations {
		avm.logger.Printf("  - %s (%s): %.2f < %.2f",
			violation.Name, violation.Severity,
			avm.getMetricValue(violation.Metric, result), violation.Threshold)
	}

	// Send alerts based on severity
	criticalViolations := avm.getViolationsBySeverity(violations, "critical")
	errorViolations := avm.getViolationsBySeverity(violations, "error")
	warningViolations := avm.getViolationsBySeverity(violations, "warning")

	if len(criticalViolations) > 0 {
		avm.sendCriticalAlert(criticalViolations, result)
	}
	if len(errorViolations) > 0 {
		avm.sendErrorAlert(errorViolations, result)
	}
	if len(warningViolations) > 0 {
		avm.sendWarningAlert(warningViolations, result)
	}
}

// getViolationsBySeverity gets violations by severity level
func (avm *AutomatedValidationManager) getViolationsBySeverity(violations []ValidationRule, severity string) []ValidationRule {
	var filtered []ValidationRule
	for _, violation := range violations {
		if violation.Severity == severity {
			filtered = append(filtered, violation)
		}
	}
	return filtered
}

// sendCriticalAlert sends critical alerts
func (avm *AutomatedValidationManager) sendCriticalAlert(violations []ValidationRule, result *AccuracyMetrics) {
	avm.logger.Printf("🚨 CRITICAL ALERT: %d critical validation violations detected", len(violations))
	// In a real implementation, this would send alerts to:
	// - PagerDuty for immediate response
	// - Slack/Discord for team notification
	// - Email for detailed reports
}

// sendErrorAlert sends error alerts
func (avm *AutomatedValidationManager) sendErrorAlert(violations []ValidationRule, result *AccuracyMetrics) {
	avm.logger.Printf("⚠️ ERROR ALERT: %d error validation violations detected", len(violations))
	// In a real implementation, this would send alerts to:
	// - Slack/Discord for team notification
	// - Email for detailed reports
}

// sendWarningAlert sends warning alerts
func (avm *AutomatedValidationManager) sendWarningAlert(violations []ValidationRule, result *AccuracyMetrics) {
	avm.logger.Printf("⚠️ WARNING ALERT: %d warning validation violations detected", len(violations))
	// In a real implementation, this would send alerts to:
	// - Slack/Discord for team notification
}

// checkForRegressions checks for performance regressions
func (avm *AutomatedValidationManager) checkForRegressions(testID string, currentResult *AccuracyMetrics) *RegressionTestResult {
	avm.mu.RLock()
	baseline, exists := avm.baselineMetrics[testID]
	avm.mu.RUnlock()

	if !exists {
		// Set current result as baseline
		avm.mu.Lock()
		avm.baselineMetrics[testID] = currentResult
		avm.mu.Unlock()

		return &RegressionTestResult{
			TestID:          testID,
			Timestamp:       time.Now(),
			CurrentMetrics:  currentResult,
			BaselineMetrics: currentResult,
			RegressionFound: false,
			Severity:        "none",
			Recommendation:  "Baseline established",
		}
	}

	// Compare current metrics with baseline
	regressionDetails := avm.compareMetrics(baseline, currentResult)
	regressionFound := avm.hasSignificantRegression(regressionDetails)

	severity := "none"
	recommendation := "No action required"

	if regressionFound {
		severity = avm.determineRegressionSeverity(regressionDetails)
		recommendation = avm.generateRegressionRecommendation(regressionDetails)
	}

	return &RegressionTestResult{
		TestID:            testID,
		Timestamp:         time.Now(),
		CurrentMetrics:    currentResult,
		BaselineMetrics:   baseline,
		RegressionFound:   regressionFound,
		RegressionDetails: regressionDetails,
		Severity:          severity,
		Recommendation:    recommendation,
	}
}

// compareMetrics compares current metrics with baseline
func (avm *AutomatedValidationManager) compareMetrics(baseline, current *AccuracyMetrics) []RegressionDetail {
	var details []RegressionDetail

	metrics := []struct {
		name     string
		baseline float64
		current  float64
	}{
		{"precision", baseline.Precision, current.Precision},
		{"recall", baseline.Recall, current.Recall},
		{"f1_score", baseline.F1Score, current.F1Score},
	}

	for _, metric := range metrics {
		changePercent := ((metric.current - metric.baseline) / metric.baseline) * 100
		exceeded := math.Abs(changePercent) > (avm.config.RegressionThreshold * 100)

		details = append(details, RegressionDetail{
			Metric:        metric.name,
			CurrentValue:  metric.current,
			BaselineValue: metric.baseline,
			ChangePercent: changePercent,
			Threshold:     avm.config.RegressionThreshold * 100,
			Exceeded:      exceeded,
		})
	}

	return details
}

// hasSignificantRegression checks if there are significant regressions
func (avm *AutomatedValidationManager) hasSignificantRegression(details []RegressionDetail) bool {
	for _, detail := range details {
		if detail.Exceeded && detail.ChangePercent < 0 {
			return true
		}
	}
	return false
}

// determineRegressionSeverity determines the severity of a regression
func (avm *AutomatedValidationManager) determineRegressionSeverity(details []RegressionDetail) string {
	maxDegradation := 0.0

	for _, detail := range details {
		if detail.ChangePercent < 0 {
			degradation := math.Abs(detail.ChangePercent)
			if degradation > maxDegradation {
				maxDegradation = degradation
			}
		}
	}

	if maxDegradation > 20 {
		return "critical"
	} else if maxDegradation > 10 {
		return "error"
	} else {
		return "warning"
	}
}

// generateRegressionRecommendation generates a recommendation for handling regression
func (avm *AutomatedValidationManager) generateRegressionRecommendation(details []RegressionDetail) string {
	var worstMetric string
	maxDegradation := 0.0

	for _, detail := range details {
		if detail.ChangePercent < 0 {
			degradation := math.Abs(detail.ChangePercent)
			if degradation > maxDegradation {
				maxDegradation = degradation
				worstMetric = detail.Metric
			}
		}
	}

	return fmt.Sprintf("Investigate %s degradation (%.1f%% decrease). Consider rule adjustments or model retraining.",
		worstMetric, maxDegradation)
}

// handleRegression handles detected regressions
func (avm *AutomatedValidationManager) handleRegression(result *RegressionTestResult) {
	avm.logger.Printf("📉 Regression detected: %s", result.Severity)
	avm.logger.Printf("  Recommendation: %s", result.Recommendation)

	for _, detail := range result.RegressionDetails {
		if detail.Exceeded && detail.ChangePercent < 0 {
			avm.logger.Printf("  - %s: %.2f -> %.2f (%.1f%% decrease)",
				detail.Metric, detail.BaselineValue, detail.CurrentValue, math.Abs(detail.ChangePercent))
		}
	}

	// Store regression result
	avm.mu.Lock()
	avm.testHistory[result.TestID] = append(avm.testHistory[result.TestID], *result.CurrentMetrics)
	avm.mu.Unlock()

	// Send regression alerts
	if avm.config.AlertOnRegression {
		avm.sendRegressionAlert(result)
	}
}

// sendRegressionAlert sends regression alerts
func (avm *AutomatedValidationManager) sendRegressionAlert(result *RegressionTestResult) {
	avm.logger.Printf("🚨 REGRESSION ALERT: %s regression detected in %s",
		result.Severity, result.TestID)

	// In a real implementation, this would send alerts to:
	// - PagerDuty for critical regressions
	// - Slack/Discord for team notification
	// - Email for detailed reports
}

// storeTestResult stores a test result in history
func (avm *AutomatedValidationManager) storeTestResult(testID string, result *AccuracyMetrics) {
	avm.mu.Lock()
	defer avm.mu.Unlock()

	avm.testHistory[testID] = append(avm.testHistory[testID], *result)

	// Trim old results based on retention policy
	avm.trimOldResults(testID)
}

// trimOldResults trims old test results based on retention policy
func (avm *AutomatedValidationManager) trimOldResults(testID string) {
	cutoff := time.Now().AddDate(0, 0, -avm.config.BaselineRetentionDays)

	var trimmed []AccuracyMetrics
	for _, result := range avm.testHistory[testID] {
		if result.TestedAt.After(cutoff) {
			trimmed = append(trimmed, result)
		}
	}

	avm.testHistory[testID] = trimmed
}

// GetValidationReport generates a validation report
func (avm *AutomatedValidationManager) GetValidationReport() map[string]interface{} {
	avm.mu.RLock()
	defer avm.mu.RUnlock()

	report := map[string]interface{}{
		"timestamp": time.Now(),
		"config":    avm.config,
		"rules":     avm.validationRules,
		"baselines": avm.baselineMetrics,
		"history":   avm.testHistory,
	}

	return report
}
