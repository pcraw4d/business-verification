package classification

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

// NewABTestManager creates a new A/B test manager
func NewABTestManager(config PerformanceWeightConfig, logger *log.Logger) *ABTestManager {
	if logger == nil {
		logger = log.Default()
	}

	return &ABTestManager{
		activeTests: make(map[string]*ABTest),
		config:      config,
		logger:      logger,
	}
}

// StartABTest starts a new A/B test for weight optimization
func (abtm *ABTestManager) StartABTest(
	methodName string,
	controlWeight float64,
	treatmentWeight float64,
	duration time.Duration,
) (*ABTest, error) {
	abtm.mutex.Lock()
	defer abtm.mutex.Unlock()

	// Check if there's already an active test for this method
	if existingTest, exists := abtm.activeTests[methodName]; exists {
		if existingTest.Status == "active" {
			return nil, fmt.Errorf("active A/B test already exists for method '%s'", methodName)
		}
	}

	// Generate test ID
	testID := fmt.Sprintf("ab_test_%s_%d", methodName, time.Now().Unix())

	// Create new A/B test
	test := &ABTest{
		TestID:            testID,
		MethodName:        methodName,
		ControlWeight:     controlWeight,
		TreatmentWeight:   treatmentWeight,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(duration),
		TrafficSplit:      abtm.config.ABTestTrafficSplit,
		Status:            "active",
		ControlResults:    &ABTestResults{},
		TreatmentResults:  &ABTestResults{},
		SignificanceLevel: abtm.config.ABTestSignificanceLevel,
	}

	// Store the test
	abtm.activeTests[methodName] = test

	abtm.logger.Printf("🧪 Started A/B test '%s' for method '%s': control=%.3f, treatment=%.3f, duration=%v",
		testID, methodName, controlWeight, treatmentWeight, duration)

	return test, nil
}

// GetTestWeight returns the appropriate weight for a method based on A/B testing
func (abtm *ABTestManager) GetTestWeight(methodName string, userID string) (float64, string, error) {
	abtm.mutex.RLock()
	defer abtm.mutex.RUnlock()

	test, exists := abtm.activeTests[methodName]
	if !exists || test.Status != "active" {
		// No active test, return default weight
		return 0.0, "no_test", nil
	}

	// Check if test has expired
	if time.Now().After(test.EndTime) {
		abtm.logger.Printf("⏰ A/B test '%s' has expired, stopping", test.TestID)
		go abtm.completeTest(test.TestID)
		return 0.0, "expired", nil
	}

	// Determine which variant to use based on user ID hash
	variant := abtm.getVariantForUser(userID, test.TrafficSplit)

	var weight float64
	var variantName string

	if variant == "control" {
		weight = test.ControlWeight
		variantName = "control"
	} else {
		weight = test.TreatmentWeight
		variantName = "treatment"
	}

	return weight, variantName, nil
}

// RecordTestResult records a result for an A/B test
func (abtm *ABTestManager) RecordTestResult(
	methodName string,
	variant string,
	accuracy float64,
	latency time.Duration,
	success bool,
) error {
	abtm.mutex.Lock()
	defer abtm.mutex.Unlock()

	test, exists := abtm.activeTests[methodName]
	if !exists || test.Status != "active" {
		return fmt.Errorf("no active A/B test found for method '%s'", methodName)
	}

	// Update results based on variant
	if variant == "control" {
		abtm.updateTestResults(test.ControlResults, accuracy, latency, success)
	} else if variant == "treatment" {
		abtm.updateTestResults(test.TreatmentResults, accuracy, latency, success)
	} else {
		return fmt.Errorf("invalid variant '%s'", variant)
	}

	// Check if we have enough samples to complete the test
	if abtm.shouldCompleteTest(test) {
		go abtm.completeTest(test.TestID)
	}

	return nil
}

// updateTestResults updates the results for a test variant
func (abtm *ABTestManager) updateTestResults(
	results *ABTestResults,
	accuracy float64,
	latency time.Duration,
	success bool,
) {
	results.SampleSize++

	// Update average accuracy
	if results.SampleSize == 1 {
		results.AverageAccuracy = accuracy
	} else {
		// Exponential moving average
		alpha := 0.1
		results.AverageAccuracy = (alpha * accuracy) + ((1 - alpha) * results.AverageAccuracy)
	}

	// Update average latency
	if results.SampleSize == 1 {
		results.AverageLatency = latency
	} else {
		// Exponential moving average
		alpha := 0.1
		results.AverageLatency = time.Duration(
			(alpha * float64(latency)) + ((1 - alpha) * float64(results.AverageLatency)),
		)
	}

	// Update success/error rates
	if success {
		results.SuccessRate = float64(results.SampleSize-int(results.ErrorRate*float64(results.SampleSize-1))) / float64(results.SampleSize)
		results.ErrorRate = 1.0 - results.SuccessRate
	} else {
		results.ErrorRate = float64(results.SampleSize-int(results.SuccessRate*float64(results.SampleSize-1))) / float64(results.SampleSize)
		results.SuccessRate = 1.0 - results.ErrorRate
	}
}

// shouldCompleteTest determines if a test should be completed
func (abtm *ABTestManager) shouldCompleteTest(test *ABTest) bool {
	// Complete if we have enough samples
	minSamples := abtm.config.ABTestMinSampleSize
	if test.ControlResults.SampleSize >= minSamples && test.TreatmentResults.SampleSize >= minSamples {
		return true
	}

	// Complete if test has been running for too long
	maxDuration := 7 * 24 * time.Hour // 7 days max
	if time.Since(test.StartTime) > maxDuration {
		return true
	}

	return false
}

// completeTest completes an A/B test and determines the winner
func (abtm *ABTestManager) completeTest(testID string) error {
	abtm.mutex.Lock()
	defer abtm.mutex.Unlock()

	// Find the test
	var test *ABTest
	var methodName string
	for name, t := range abtm.activeTests {
		if t.TestID == testID {
			test = t
			methodName = name
			break
		}
	}

	if test == nil {
		return fmt.Errorf("test '%s' not found", testID)
	}

	// Mark test as completed
	test.Status = "completed"
	test.EndTime = time.Now()

	// Perform statistical analysis
	isSignificant, winner := abtm.performStatisticalAnalysis(test)
	test.IsSignificant = isSignificant
	test.Winner = winner

	// Log results
	abtm.logger.Printf("🏁 Completed A/B test '%s' for method '%s':", testID, methodName)
	abtm.logger.Printf("   Control: samples=%d, accuracy=%.3f, latency=%v",
		test.ControlResults.SampleSize, test.ControlResults.AverageAccuracy, test.ControlResults.AverageLatency)
	abtm.logger.Printf("   Treatment: samples=%d, accuracy=%.3f, latency=%v",
		test.TreatmentResults.SampleSize, test.TreatmentResults.AverageAccuracy, test.TreatmentResults.AverageLatency)
	abtm.logger.Printf("   Significant: %v, Winner: %s", isSignificant, winner)

	// Save test results
	if err := abtm.saveTestResults(test); err != nil {
		abtm.logger.Printf("⚠️ Failed to save A/B test results: %v", err)
	}

	return nil
}

// performStatisticalAnalysis performs statistical analysis to determine test significance
func (abtm *ABTestManager) performStatisticalAnalysis(test *ABTest) (bool, string) {
	control := test.ControlResults
	treatment := test.TreatmentResults

	// Check if we have enough samples
	if control.SampleSize < 30 || treatment.SampleSize < 30 {
		abtm.logger.Printf("⚠️ Insufficient samples for statistical analysis: control=%d, treatment=%d",
			control.SampleSize, treatment.SampleSize)
		return false, "inconclusive"
	}

	// Perform two-sample t-test for accuracy
	// This is a simplified implementation - in production, you'd use a proper statistical library
	controlMean := control.AverageAccuracy
	treatmentMean := treatment.AverageAccuracy

	// Calculate pooled standard deviation (simplified)
	controlVar := abtm.estimateVariance(control.AverageAccuracy, control.SampleSize)
	treatmentVar := abtm.estimateVariance(treatment.AverageAccuracy, treatment.SampleSize)

	pooledStdErr := math.Sqrt((controlVar / float64(control.SampleSize)) + (treatmentVar / float64(treatment.SampleSize)))

	if pooledStdErr == 0 {
		return false, "inconclusive"
	}

	// Calculate t-statistic
	tStat := (treatmentMean - controlMean) / pooledStdErr

	// Calculate degrees of freedom (simplified)
	df := float64(control.SampleSize + treatment.SampleSize - 2)

	// Calculate p-value (simplified - in production, use proper t-distribution)
	pValue := abtm.calculatePValue(tStat, df)

	// Determine significance
	isSignificant := pValue < test.SignificanceLevel

	// Determine winner
	var winner string
	if !isSignificant {
		winner = "inconclusive"
	} else if treatmentMean > controlMean {
		winner = "treatment"
	} else {
		winner = "control"
	}

	abtm.logger.Printf("📊 Statistical analysis: t=%.3f, p=%.6f, significant=%v, winner=%s",
		tStat, pValue, isSignificant, winner)

	return isSignificant, winner
}

// estimateVariance estimates variance from mean and sample size (simplified)
func (abtm *ABTestManager) estimateVariance(mean float64, sampleSize int) float64 {
	// This is a very simplified estimation
	// In production, you'd track actual variance
	return mean * (1 - mean) / float64(sampleSize)
}

// calculatePValue calculates p-value for t-test (simplified)
func (abtm *ABTestManager) calculatePValue(tStat float64, df float64) float64 {
	// This is a very simplified p-value calculation
	// In production, you'd use a proper statistical library

	// For large degrees of freedom, t-distribution approximates normal
	if df > 30 {
		// Two-tailed test
		absT := math.Abs(tStat)
		if absT > 2.576 { // 99% confidence
			return 0.01
		} else if absT > 1.96 { // 95% confidence
			return 0.05
		} else if absT > 1.645 { // 90% confidence
			return 0.10
		}
	}

	return 0.20 // Default to non-significant
}

// getVariantForUser determines which variant a user should see
func (abtm *ABTestManager) getVariantForUser(userID string, trafficSplit float64) string {
	// Use user ID hash to ensure consistent assignment
	hash := abtm.hashString(userID)

	// Convert hash to float between 0 and 1
	hashFloat := float64(hash) / float64(math.MaxUint64)

	if hashFloat < trafficSplit {
		return "treatment"
	}
	return "control"
}

// hashString creates a hash of a string
func (abtm *ABTestManager) hashString(s string) uint64 {
	hash := uint64(0)
	for _, c := range s {
		hash = hash*31 + uint64(c)
	}
	return hash
}

// GetActiveTests returns all active A/B tests
func (abtm *ABTestManager) GetActiveTests() map[string]*ABTest {
	abtm.mutex.RLock()
	defer abtm.mutex.RUnlock()

	// Return a copy
	result := make(map[string]*ABTest)
	for methodName, test := range abtm.activeTests {
		if test.Status == "active" {
			// Create a copy
			copy := &ABTest{
				TestID:            test.TestID,
				MethodName:        test.MethodName,
				ControlWeight:     test.ControlWeight,
				TreatmentWeight:   test.TreatmentWeight,
				StartTime:         test.StartTime,
				EndTime:           test.EndTime,
				TrafficSplit:      test.TrafficSplit,
				Status:            test.Status,
				SignificanceLevel: test.SignificanceLevel,
			}

			// Copy results
			if test.ControlResults != nil {
				copy.ControlResults = &ABTestResults{
					SampleSize:      test.ControlResults.SampleSize,
					AverageAccuracy: test.ControlResults.AverageAccuracy,
					AverageLatency:  test.ControlResults.AverageLatency,
					SuccessRate:     test.ControlResults.SuccessRate,
					ErrorRate:       test.ControlResults.ErrorRate,
				}
			}

			if test.TreatmentResults != nil {
				copy.TreatmentResults = &ABTestResults{
					SampleSize:      test.TreatmentResults.SampleSize,
					AverageAccuracy: test.TreatmentResults.AverageAccuracy,
					AverageLatency:  test.TreatmentResults.AverageLatency,
					SuccessRate:     test.TreatmentResults.SuccessRate,
					ErrorRate:       test.TreatmentResults.ErrorRate,
				}
			}

			result[methodName] = copy
		}
	}

	return result
}

// saveTestResults saves A/B test results to file
func (abtm *ABTestManager) saveTestResults(test *ABTest) error {
	// Create data directory if it doesn't exist
	dataDir := "data/ab_tests"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Save to file
	filename := filepath.Join(dataDir, fmt.Sprintf("%s.json", test.TestID))
	data, err := json.MarshalIndent(test, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	abtm.logger.Printf("💾 Saved A/B test results to %s", filename)
	return nil
}

// manageABTests manages active A/B tests (checks for completion, etc.)
func (abtm *ABTestManager) manageABTests() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			abtm.checkAndCompleteExpiredTests()
		}
	}
}

// checkAndCompleteExpiredTests checks for expired tests and completes them
func (abtm *ABTestManager) checkAndCompleteExpiredTests() {
	abtm.mutex.Lock()
	defer abtm.mutex.Unlock()

	now := time.Now()
	for methodName, test := range abtm.activeTests {
		if test.Status == "active" && now.After(test.EndTime) {
			abtm.logger.Printf("⏰ A/B test '%s' for method '%s' has expired", test.TestID, methodName)
			go abtm.completeTest(test.TestID)
		}
	}
}

// GetTestSummary returns a summary of A/B testing activity
func (abtm *ABTestManager) GetTestSummary() map[string]interface{} {
	abtm.mutex.RLock()
	defer abtm.mutex.RUnlock()

	summary := make(map[string]interface{})

	activeCount := 0
	completedCount := 0

	for _, test := range abtm.activeTests {
		if test.Status == "active" {
			activeCount++
		} else if test.Status == "completed" {
			completedCount++
		}
	}

	summary["active_tests"] = activeCount
	summary["completed_tests"] = completedCount
	summary["total_tests"] = len(abtm.activeTests)
	summary["config"] = map[string]interface{}{
		"ab_testing_enabled":         abtm.config.ABTestingEnabled,
		"ab_test_duration":           abtm.config.ABTestDuration,
		"ab_test_traffic_split":      abtm.config.ABTestTrafficSplit,
		"ab_test_min_sample_size":    abtm.config.ABTestMinSampleSize,
		"ab_test_significance_level": abtm.config.ABTestSignificanceLevel,
	}

	return summary
}
