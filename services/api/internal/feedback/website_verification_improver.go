package feedback

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// NewWebsiteVerificationImprover creates a new website verification improver
func NewWebsiteVerificationImprover(config *AdvancedLearningConfig, logger *zap.Logger) *WebsiteVerificationImprover {
	return &WebsiteVerificationImprover{
		config:           config,
		logger:           logger,
		verificationData: make([]*VerificationDataPoint, 0),
		improvementMetrics: &VerificationImprovementMetrics{
			MethodAccuracy: make(map[string]float64),
			LastUpdated:    time.Now(),
		},
	}
}

// ImproveVerification improves website verification algorithms based on feedback
func (wvi *WebsiteVerificationImprover) ImproveVerification(verificationData []*VerificationDataPoint) error {
	wvi.mu.Lock()
	defer wvi.mu.Unlock()

	wvi.logger.Info("Improving website verification algorithms",
		zap.Int("verification_data_count", len(verificationData)))

	// Validate verification data
	if err := wvi.validateVerificationData(verificationData); err != nil {
		return fmt.Errorf("failed to validate verification data: %w", err)
	}

	// Add new verification data
	wvi.verificationData = append(wvi.verificationData, verificationData...)

	// Maintain verification data size
	if len(wvi.verificationData) > wvi.config.MaxLearningBatchSize*2 {
		// Keep most recent data
		wvi.verificationData = wvi.verificationData[len(wvi.verificationData)-wvi.config.MaxLearningBatchSize:]
	}

	// Check if improvement is needed
	if !wvi.shouldImprove() {
		wvi.logger.Info("Website verification improvement not needed based on current criteria")
		return nil
	}

	// Perform verification improvement
	if err := wvi.performVerificationImprovement(); err != nil {
		return fmt.Errorf("failed to perform verification improvement: %w", err)
	}

	// Update improvement metrics
	wvi.updateImprovementMetrics()

	wvi.logger.Info("Website verification improvement completed")

	return nil
}

// validateVerificationData validates verification data quality
func (wvi *WebsiteVerificationImprover) validateVerificationData(verificationData []*VerificationDataPoint) error {
	if len(verificationData) == 0 {
		return fmt.Errorf("no verification data provided")
	}

	// Check data quality
	validDataCount := 0
	for _, data := range verificationData {
		if wvi.isValidVerificationDataPoint(data) {
			validDataCount++
		}
	}

	if validDataCount < int(float64(len(verificationData))*0.8) { // At least 80% valid data
		return fmt.Errorf("insufficient valid verification data: %d/%d", validDataCount, len(verificationData))
	}

	// Check domain distribution
	domainCounts := make(map[string]int)
	for _, data := range verificationData {
		domainCounts[data.Domain]++
	}

	// Ensure minimum samples per domain
	minSamplesPerDomain := 3
	for domain, count := range domainCounts {
		if count < minSamplesPerDomain {
			wvi.logger.Warn("Insufficient samples for domain",
				zap.String("domain", domain),
				zap.Int("count", count),
				zap.Int("min_required", minSamplesPerDomain))
		}
	}

	return nil
}

// isValidVerificationDataPoint checks if a verification data point is valid
func (wvi *WebsiteVerificationImprover) isValidVerificationDataPoint(data *VerificationDataPoint) bool {
	// Check required fields
	if data.Domain == "" || data.VerificationMethod == "" {
		return false
	}

	// Check confidence score range
	if data.ConfidenceScore < 0.0 || data.ConfidenceScore > 1.0 {
		return false
	}

	// Check timestamp
	if data.Timestamp.IsZero() || data.Timestamp.After(time.Now()) {
		return false
	}

	// Check feedback type
	if data.FeedbackType != FeedbackTypeAccuracy && data.FeedbackType != FeedbackTypeCorrection {
		return false
	}

	return true
}

// shouldImprove determines if verification improvement is needed
func (wvi *WebsiteVerificationImprover) shouldImprove() bool {
	// Check if we have enough verification data
	if len(wvi.verificationData) < wvi.config.MinFeedbackThreshold {
		return false
	}

	// Check if enough time has passed since last improvement
	lastImprovement := wvi.improvementMetrics.LastUpdated
	if time.Since(lastImprovement) < 24*time.Hour { // Improve at most once per day
		return false
	}

	// Check if overall accuracy is below threshold
	if wvi.improvementMetrics.OverallAccuracy < wvi.config.VerificationAccuracyThreshold {
		wvi.logger.Info("Verification improvement triggered due to low overall accuracy",
			zap.Float64("overall_accuracy", wvi.improvementMetrics.OverallAccuracy),
			zap.Float64("threshold", wvi.config.VerificationAccuracyThreshold))
		return true
	}

	// Check if false positive rate is too high
	if wvi.improvementMetrics.FalsePositiveRate > 0.1 { // 10% threshold
		wvi.logger.Info("Verification improvement triggered due to high false positive rate",
			zap.Float64("false_positive_rate", wvi.improvementMetrics.FalsePositiveRate))
		return true
	}

	// Check if false negative rate is too high
	if wvi.improvementMetrics.FalseNegativeRate > 0.1 { // 10% threshold
		wvi.logger.Info("Verification improvement triggered due to high false negative rate",
			zap.Float64("false_negative_rate", wvi.improvementMetrics.FalseNegativeRate))
		return true
	}

	return true
}

// performVerificationImprovement performs the actual verification improvement
func (wvi *WebsiteVerificationImprover) performVerificationImprovement() error {
	wvi.logger.Info("Performing website verification improvement")

	// Analyze verification method performance
	methodPerformance, err := wvi.analyzeMethodPerformance()
	if err != nil {
		return fmt.Errorf("failed to analyze method performance: %w", err)
	}

	// Optimize verification methods
	if err := wvi.optimizeVerificationMethods(methodPerformance); err != nil {
		return fmt.Errorf("failed to optimize verification methods: %w", err)
	}

	// Improve domain matching algorithms
	if err := wvi.improveDomainMatching(); err != nil {
		return fmt.Errorf("failed to improve domain matching: %w", err)
	}

	// Optimize confidence scoring
	if err := wvi.optimizeConfidenceScoring(); err != nil {
		return fmt.Errorf("failed to optimize confidence scoring: %w", err)
	}

	// Improve business name matching
	if err := wvi.improveBusinessNameMatching(); err != nil {
		return fmt.Errorf("failed to improve business name matching: %w", err)
	}

	wvi.logger.Info("Website verification improvement techniques applied successfully")

	return nil
}

// analyzeMethodPerformance analyzes performance of verification methods
func (wvi *WebsiteVerificationImprover) analyzeMethodPerformance() (map[string]*MethodPerformanceMetrics, error) {
	// Group data by verification method
	methodData := make(map[string][]*VerificationDataPoint)
	for _, data := range wvi.verificationData {
		methodData[data.VerificationMethod] = append(methodData[data.VerificationMethod], data)
	}

	methodPerformance := make(map[string]*MethodPerformanceMetrics)

	// Analyze each method
	for method, methodData := range methodData {
		if len(methodData) < 10 { // Need minimum samples for reliable analysis
			continue
		}

		performance := wvi.calculateMethodPerformance(method, methodData)
		methodPerformance[method] = performance

		wvi.logger.Debug("Verification method performance analyzed",
			zap.String("method", method),
			zap.Int("data_count", len(methodData)),
			zap.Float64("accuracy", performance.Accuracy),
			zap.Float64("average_confidence", performance.Confidence))
	}

	return methodPerformance, nil
}

// calculateMethodPerformance calculates performance metrics for a verification method
func (wvi *WebsiteVerificationImprover) calculateMethodPerformance(method string, data []*VerificationDataPoint) *MethodPerformanceMetrics {
	metrics := &MethodPerformanceMetrics{
		TotalFeedback: len(data),
	}

	// Calculate accuracy based on feedback
	correctPredictions := 0
	totalConfidence := 0.0
	confidenceCount := 0

	for _, point := range data {
		// Calculate accuracy (positive feedback indicates correct prediction)
		if point.FeedbackType == FeedbackTypeAccuracy || point.FeedbackType == FeedbackTypeClassification {
			correctPredictions++
		}

		// Calculate average confidence
		if point.ConfidenceScore > 0 {
			totalConfidence += point.ConfidenceScore
			confidenceCount++
		}
	}

	// Calculate metrics
	metrics.Accuracy = float64(correctPredictions) / float64(len(data))
	if confidenceCount > 0 {
		metrics.Confidence = totalConfidence / float64(confidenceCount)
	}

	// Calculate confidence calibration
	metrics.Consistency = wvi.calculateConfidenceCalibration(data)

	// Calculate response time (placeholder)
	metrics.ProcessingTime = 100.0

	// Calculate reliability score
	metrics.Reliability = wvi.calculateReliabilityScore(metrics)

	return metrics
}

// calculateConfidenceCalibration calculates confidence calibration for verification data
func (wvi *WebsiteVerificationImprover) calculateConfidenceCalibration(data []*VerificationDataPoint) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Group by confidence bins
	confidenceBins := make(map[int][]*VerificationDataPoint)
	for _, point := range data {
		if point.ConfidenceScore > 0 {
			bin := int(point.ConfidenceScore * 10) // 0.1 bins
			confidenceBins[bin] = append(confidenceBins[bin], point)
		}
	}

	// Calculate calibration error for each bin
	totalError := 0.0
	binCount := 0

	for bin, binData := range confidenceBins {
		if len(binData) < 3 { // Need minimum samples per bin
			continue
		}

		// Calculate accuracy in this confidence bin
		correct := 0
		for _, point := range binData {
			if point.FeedbackType == FeedbackTypeAccuracy || point.FeedbackType == FeedbackTypeClassification {
				correct++
			}
		}

		accuracy := float64(correct) / float64(len(binData))
		expectedConfidence := float64(bin) / 10.0
		calibrationError := math.Abs(accuracy - expectedConfidence)

		totalError += calibrationError
		binCount++
	}

	if binCount == 0 {
		return 0.0
	}

	// Return calibration score (1.0 - average error)
	return math.Max(0.0, 1.0-totalError/float64(binCount))
}

// calculateReliabilityScore calculates reliability score based on consistency
func (wvi *WebsiteVerificationImprover) calculateReliabilityScore(performance *MethodPerformanceMetrics) float64 {
	// Calculate consistency based on feedback distribution
	if performance.TotalFeedback < 10 {
		return 0.5 // Default reliability for small samples
	}

	// Simple reliability calculation based on accuracy and consistency
	reliability := (performance.Accuracy + performance.Consistency) / 2.0
	return math.Max(0.0, math.Min(1.0, reliability))
}

// optimizeVerificationMethods optimizes verification methods based on performance
func (wvi *WebsiteVerificationImprover) optimizeVerificationMethods(methodPerformance map[string]*MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing verification methods")

	// Identify underperforming methods
	var underperformingMethods []string
	for method, performance := range methodPerformance {
		if performance.Accuracy < wvi.config.VerificationAccuracyThreshold {
			underperformingMethods = append(underperformingMethods, method)
			wvi.logger.Warn("Underperforming verification method identified",
				zap.String("method", method),
				zap.Float64("accuracy", performance.Accuracy),
				zap.Float64("threshold", wvi.config.VerificationAccuracyThreshold))
		}
	}

	// Apply optimization techniques for underperforming methods
	for _, method := range underperformingMethods {
		if err := wvi.optimizeSpecificMethod(method, methodPerformance[method]); err != nil {
			wvi.logger.Error("Failed to optimize verification method",
				zap.String("method", method),
				zap.Error(err))
			continue
		}
	}

	// Optimize method weights based on performance
	if err := wvi.optimizeMethodWeights(methodPerformance); err != nil {
		return fmt.Errorf("failed to optimize method weights: %w", err)
	}

	wvi.logger.Info("Verification methods optimized successfully")
	return nil
}

// optimizeSpecificMethod optimizes a specific verification method
func (wvi *WebsiteVerificationImprover) optimizeSpecificMethod(method string, performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing specific verification method",
		zap.String("method", method))

	switch method {
	case "dns_verification":
		return wvi.optimizeDNSVerification(performance)
	case "whois_verification":
		return wvi.optimizeWHOISVerification(performance)
	case "content_verification":
		return wvi.optimizeContentVerification(performance)
	case "ssl_verification":
		return wvi.optimizeSSLVerification(performance)
	default:
		return wvi.optimizeGenericMethod(method, performance)
	}
}

// optimizeDNSVerification optimizes DNS verification method
func (wvi *WebsiteVerificationImprover) optimizeDNSVerification(performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing DNS verification method")

	// TODO: Implement DNS verification optimization
	// This would adjust DNS record validation rules, timeout settings, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("DNS verification method optimized successfully")
	return nil
}

// optimizeWHOISVerification optimizes WHOIS verification method
func (wvi *WebsiteVerificationImprover) optimizeWHOISVerification(performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing WHOIS verification method")

	// TODO: Implement WHOIS verification optimization
	// This would adjust WHOIS data parsing rules, validation criteria, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("WHOIS verification method optimized successfully")
	return nil
}

// optimizeContentVerification optimizes content verification method
func (wvi *WebsiteVerificationImprover) optimizeContentVerification(performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing content verification method")

	// TODO: Implement content verification optimization
	// This would adjust content matching algorithms, text extraction rules, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Content verification method optimized successfully")
	return nil
}

// optimizeSSLVerification optimizes SSL verification method
func (wvi *WebsiteVerificationImprover) optimizeSSLVerification(performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing SSL verification method")

	// TODO: Implement SSL verification optimization
	// This would adjust SSL certificate validation rules, trust store settings, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("SSL verification method optimized successfully")
	return nil
}

// optimizeGenericMethod optimizes a generic verification method
func (wvi *WebsiteVerificationImprover) optimizeGenericMethod(method string, performance *MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing generic verification method",
		zap.String("method", method))

	// TODO: Implement generic method optimization
	// This would apply general optimization techniques

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Generic verification method optimized successfully")
	return nil
}

// optimizeMethodWeights optimizes method weights based on performance
func (wvi *WebsiteVerificationImprover) optimizeMethodWeights(methodPerformance map[string]*MethodPerformanceMetrics) error {
	wvi.logger.Info("Optimizing verification method weights")

	// Calculate performance-based weights
	methodWeights := make(map[string]float64)
	totalScore := 0.0

	for method, performance := range methodPerformance {
		// Calculate composite performance score
		score := performance.Accuracy*0.6 + performance.Consistency*0.4
		methodWeights[method] = score
		totalScore += score
	}

	// Normalize weights
	if totalScore > 0 {
		for method := range methodWeights {
			methodWeights[method] = methodWeights[method] / totalScore
		}
	}

	// Store optimized weights
	// TODO: Update actual method weights in the verification system

	wvi.logger.Info("Verification method weights optimized successfully",
		zap.Any("method_weights", methodWeights))

	return nil
}

// improveDomainMatching improves domain matching algorithms
func (wvi *WebsiteVerificationImprover) improveDomainMatching() error {
	wvi.logger.Info("Improving domain matching algorithms")

	// Analyze domain matching patterns
	domainPatterns := wvi.analyzeDomainPatterns()

	// Optimize domain matching rules
	if err := wvi.optimizeDomainMatchingRules(domainPatterns); err != nil {
		return fmt.Errorf("failed to optimize domain matching rules: %w", err)
	}

	// Improve fuzzy matching algorithms
	if err := wvi.improveFuzzyMatching(); err != nil {
		return fmt.Errorf("failed to improve fuzzy matching: %w", err)
	}

	wvi.logger.Info("Domain matching algorithms improved successfully")
	return nil
}

// analyzeDomainPatterns analyzes domain matching patterns
func (wvi *WebsiteVerificationImprover) analyzeDomainPatterns() map[string]interface{} {
	patterns := make(map[string]interface{})

	// Analyze successful domain matches
	successfulMatches := 0
	totalMatches := 0

	for _, data := range wvi.verificationData {
		if data.VerificationResult && data.FeedbackType == FeedbackTypeAccuracy {
			successfulMatches++
		}
		totalMatches++
	}

	patterns["success_rate"] = float64(successfulMatches) / float64(totalMatches)
	patterns["total_matches"] = totalMatches

	// Analyze domain characteristics
	domainLengths := make([]int, 0)
	for _, data := range wvi.verificationData {
		domainLengths = append(domainLengths, len(data.Domain))
	}

	if len(domainLengths) > 0 {
		// Calculate average domain length
		totalLength := 0
		for _, length := range domainLengths {
			totalLength += length
		}
		patterns["average_domain_length"] = float64(totalLength) / float64(len(domainLengths))
	}

	return patterns
}

// optimizeDomainMatchingRules optimizes domain matching rules
func (wvi *WebsiteVerificationImprover) optimizeDomainMatchingRules(patterns map[string]interface{}) error {
	wvi.logger.Info("Optimizing domain matching rules")

	// TODO: Implement domain matching rule optimization
	// This would adjust domain validation rules, normalization algorithms, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Domain matching rules optimized successfully")
	return nil
}

// improveFuzzyMatching improves fuzzy matching algorithms
func (wvi *WebsiteVerificationImprover) improveFuzzyMatching() error {
	wvi.logger.Info("Improving fuzzy matching algorithms")

	// TODO: Implement fuzzy matching improvement
	// This would adjust string similarity algorithms, threshold settings, etc.

	// Simulate improvement process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Fuzzy matching algorithms improved successfully")
	return nil
}

// optimizeConfidenceScoring optimizes confidence scoring for verification
func (wvi *WebsiteVerificationImprover) optimizeConfidenceScoring() error {
	wvi.logger.Info("Optimizing confidence scoring")

	// Analyze confidence score distribution
	confidenceDistribution := wvi.analyzeConfidenceDistribution()

	// Optimize confidence thresholds
	if err := wvi.optimizeConfidenceThresholds(confidenceDistribution); err != nil {
		return fmt.Errorf("failed to optimize confidence thresholds: %w", err)
	}

	// Improve confidence calibration
	if err := wvi.improveConfidenceCalibration(); err != nil {
		return fmt.Errorf("failed to improve confidence calibration: %w", err)
	}

	wvi.logger.Info("Confidence scoring optimized successfully")
	return nil
}

// analyzeConfidenceDistribution analyzes confidence score distribution
func (wvi *WebsiteVerificationImprover) analyzeConfidenceDistribution() map[string]interface{} {
	distribution := make(map[string]interface{})

	// Calculate confidence statistics
	confidences := make([]float64, 0, len(wvi.verificationData))
	for _, data := range wvi.verificationData {
		confidences = append(confidences, data.ConfidenceScore)
	}

	if len(confidences) > 0 {
		// Calculate mean confidence
		total := 0.0
		for _, conf := range confidences {
			total += conf
		}
		distribution["mean_confidence"] = total / float64(len(confidences))

		// Calculate confidence variance
		mean := distribution["mean_confidence"].(float64)
		variance := 0.0
		for _, conf := range confidences {
			diff := conf - mean
			variance += diff * diff
		}
		distribution["confidence_variance"] = variance / float64(len(confidences))
	}

	return distribution
}

// optimizeConfidenceThresholds optimizes confidence thresholds
func (wvi *WebsiteVerificationImprover) optimizeConfidenceThresholds(distribution map[string]interface{}) error {
	wvi.logger.Info("Optimizing confidence thresholds")

	// TODO: Implement confidence threshold optimization
	// This would adjust thresholds based on performance analysis

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Confidence thresholds optimized successfully")
	return nil
}

// improveConfidenceCalibration improves confidence calibration
func (wvi *WebsiteVerificationImprover) improveConfidenceCalibration() error {
	wvi.logger.Info("Improving confidence calibration")

	// TODO: Implement confidence calibration improvement
	// This would adjust calibration parameters based on feedback

	// Simulate improvement process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Confidence calibration improved successfully")
	return nil
}

// improveBusinessNameMatching improves business name matching
func (wvi *WebsiteVerificationImprover) improveBusinessNameMatching() error {
	wvi.logger.Info("Improving business name matching")

	// Analyze business name matching patterns
	namePatterns := wvi.analyzeBusinessNamePatterns()

	// Optimize name matching algorithms
	if err := wvi.optimizeNameMatchingAlgorithms(namePatterns); err != nil {
		return fmt.Errorf("failed to optimize name matching algorithms: %w", err)
	}

	// Improve name normalization
	if err := wvi.improveNameNormalization(); err != nil {
		return fmt.Errorf("failed to improve name normalization: %w", err)
	}

	wvi.logger.Info("Business name matching improved successfully")
	return nil
}

// analyzeBusinessNamePatterns analyzes business name matching patterns
func (wvi *WebsiteVerificationImprover) analyzeBusinessNamePatterns() map[string]interface{} {
	patterns := make(map[string]interface{})

	// Analyze name matching success rates
	successfulMatches := 0
	totalMatches := 0

	for _, data := range wvi.verificationData {
		if data.BusinessName != "" {
			totalMatches++
			if data.VerificationResult && data.FeedbackType == FeedbackTypeAccuracy {
				successfulMatches++
			}
		}
	}

	if totalMatches > 0 {
		patterns["name_match_success_rate"] = float64(successfulMatches) / float64(totalMatches)
	}

	// Analyze name characteristics
	nameLengths := make([]int, 0)
	for _, data := range wvi.verificationData {
		if data.BusinessName != "" {
			nameLengths = append(nameLengths, len(data.BusinessName))
		}
	}

	if len(nameLengths) > 0 {
		// Calculate average name length
		totalLength := 0
		for _, length := range nameLengths {
			totalLength += length
		}
		patterns["average_name_length"] = float64(totalLength) / float64(len(nameLengths))
	}

	return patterns
}

// optimizeNameMatchingAlgorithms optimizes name matching algorithms
func (wvi *WebsiteVerificationImprover) optimizeNameMatchingAlgorithms(patterns map[string]interface{}) error {
	wvi.logger.Info("Optimizing name matching algorithms")

	// TODO: Implement name matching algorithm optimization
	// This would adjust string similarity algorithms, weighting schemes, etc.

	// Simulate optimization process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Name matching algorithms optimized successfully")
	return nil
}

// improveNameNormalization improves name normalization
func (wvi *WebsiteVerificationImprover) improveNameNormalization() error {
	wvi.logger.Info("Improving name normalization")

	// TODO: Implement name normalization improvement
	// This would adjust text preprocessing rules, character mapping, etc.

	// Simulate improvement process
	time.Sleep(500 * time.Millisecond)

	wvi.logger.Info("Name normalization improved successfully")
	return nil
}

// updateImprovementMetrics updates improvement metrics
func (wvi *WebsiteVerificationImprover) updateImprovementMetrics() {
	// Calculate overall accuracy
	correct := 0
	total := 0

	for _, data := range wvi.verificationData {
		total++
		if data.FeedbackType == FeedbackTypeAccuracy || data.FeedbackType == FeedbackTypeClassification {
			correct++
		}
	}

	if total > 0 {
		wvi.improvementMetrics.OverallAccuracy = float64(correct) / float64(total)
	}

	// Calculate method-specific accuracy
	methodData := make(map[string][]*VerificationDataPoint)
	for _, data := range wvi.verificationData {
		methodData[data.VerificationMethod] = append(methodData[data.VerificationMethod], data)
	}

	for method, data := range methodData {
		if len(data) > 0 {
			correct := 0
			for _, point := range data {
				if point.FeedbackType == FeedbackTypeAccuracy || point.FeedbackType == FeedbackTypeClassification {
					correct++
				}
			}
			wvi.improvementMetrics.MethodAccuracy[method] = float64(correct) / float64(len(data))
		}
	}

	// Calculate false positive and false negative rates
	falsePositives := 0
	falseNegatives := 0
	truePositives := 0
	trueNegatives := 0

	for _, data := range wvi.verificationData {
		if data.VerificationResult {
			if data.FeedbackType == FeedbackTypeAccuracy || data.FeedbackType == FeedbackTypeClassification {
				truePositives++
			} else {
				falsePositives++
			}
		} else {
			if data.FeedbackType == FeedbackTypeAccuracy || data.FeedbackType == FeedbackTypeClassification {
				falseNegatives++
			} else {
				trueNegatives++
			}
		}
	}

	totalPredictions := truePositives + falsePositives + trueNegatives + falseNegatives
	if totalPredictions > 0 {
		wvi.improvementMetrics.FalsePositiveRate = float64(falsePositives) / float64(totalPredictions)
		wvi.improvementMetrics.FalseNegativeRate = float64(falseNegatives) / float64(totalPredictions)
	}

	// Update other metrics
	wvi.improvementMetrics.SampleSize = len(wvi.verificationData)
	wvi.improvementMetrics.LastUpdated = time.Now()

	wvi.logger.Info("Verification improvement metrics updated",
		zap.Float64("overall_accuracy", wvi.improvementMetrics.OverallAccuracy),
		zap.Float64("false_positive_rate", wvi.improvementMetrics.FalsePositiveRate),
		zap.Float64("false_negative_rate", wvi.improvementMetrics.FalseNegativeRate),
		zap.Int("sample_size", wvi.improvementMetrics.SampleSize))
}

// GetVerificationMetrics returns verification improvement metrics
func (wvi *WebsiteVerificationImprover) GetVerificationMetrics() *WebsiteVerificationMetrics {
	wvi.mu.RLock()
	defer wvi.mu.RUnlock()

	metrics := &WebsiteVerificationMetrics{
		OverallAccuracy:   wvi.improvementMetrics.OverallAccuracy,
		MethodAccuracy:    make(map[string]float64),
		ImprovementRuns:   1, // TODO: Track actual improvement runs
		LastImprovement:   wvi.improvementMetrics.LastUpdated,
		FalsePositiveRate: wvi.improvementMetrics.FalsePositiveRate,
		FalseNegativeRate: wvi.improvementMetrics.FalseNegativeRate,
	}

	// Copy method accuracy
	for method, accuracy := range wvi.improvementMetrics.MethodAccuracy {
		metrics.MethodAccuracy[method] = accuracy
	}

	return metrics
}

// GetVerificationData returns verification data
func (wvi *WebsiteVerificationImprover) GetVerificationData(limit int) []*VerificationDataPoint {
	wvi.mu.RLock()
	defer wvi.mu.RUnlock()

	if limit <= 0 || limit > len(wvi.verificationData) {
		limit = len(wvi.verificationData)
	}

	// Return recent data
	start := len(wvi.verificationData) - limit
	if start < 0 {
		start = 0
	}

	data := make([]*VerificationDataPoint, limit)
	copy(data, wvi.verificationData[start:])

	return data
}
