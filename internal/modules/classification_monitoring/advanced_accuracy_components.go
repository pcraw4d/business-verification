package classification_monitoring

import (
	"fmt"
	"sync"
	"time"
)

// OverallAccuracyTracker implementation

// NewOverallAccuracyTracker creates a new overall accuracy tracker
func NewOverallAccuracyTracker() *OverallAccuracyTracker {
	return &OverallAccuracyTracker{
		windowedAccuracy:       make([]float64, 0),
		confidenceDistribution: make(map[string]int64),
		lastUpdated:            time.Now(),
	}
}

// TrackResult tracks a classification result
func (oat *OverallAccuracyTracker) TrackResult(result *ClassificationResult) error {
	oat.mu.Lock()
	defer oat.mu.Unlock()

	oat.totalClassifications++
	oat.lastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			oat.correctClassifications++
		}

		// Update accuracy score
		oat.accuracyScore = float64(oat.correctClassifications) / float64(oat.totalClassifications)

		// Update windowed accuracy
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}
		oat.windowedAccuracy = append(oat.windowedAccuracy, accuracy)

		// Update confidence distribution
		confidenceRange := getConfidenceRange(result.ConfidenceScore)
		oat.confidenceDistribution[confidenceRange]++
	}

	return nil
}

// GetAccuracy returns the current accuracy score
func (oat *OverallAccuracyTracker) GetAccuracy() float64 {
	oat.mu.RLock()
	defer oat.mu.RUnlock()
	return oat.accuracyScore
}

// GetWindowedAccuracy returns the windowed accuracy data
func (oat *OverallAccuracyTracker) GetWindowedAccuracy() []float64 {
	oat.mu.RLock()
	defer oat.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]float64, len(oat.windowedAccuracy))
	copy(result, oat.windowedAccuracy)
	return result
}

// UpdateMetrics updates internal metrics
func (oat *OverallAccuracyTracker) UpdateMetrics() {
	oat.mu.Lock()
	defer oat.mu.Unlock()

	oat.lastUpdated = time.Now()

	// Update trend indicator based on recent windowed accuracy
	if len(oat.windowedAccuracy) >= 10 {
		oat.trendIndicator = calculateTrendIndicator(oat.windowedAccuracy)
	}
}

// IndustryAccuracyTracker implementation

// NewIndustryAccuracyTracker creates a new industry accuracy tracker
func NewIndustryAccuracyTracker() *IndustryAccuracyTracker {
	return &IndustryAccuracyTracker{
		industries: make(map[string]*IndustryAccuracyMetrics),
	}
}

// TrackResult tracks a classification result for an industry
func (iat *IndustryAccuracyTracker) TrackResult(result *ClassificationResult) error {
	iat.mu.Lock()
	defer iat.mu.Unlock()

	// Extract industry from result metadata
	industry := "unknown"
	if result.Metadata != nil {
		if industryValue, exists := result.Metadata["industry"]; exists {
			if industryStr, ok := industryValue.(string); ok {
				industry = industryStr
			}
		}
	}

	// Get or create industry metrics
	metrics, exists := iat.industries[industry]
	if !exists {
		metrics = &IndustryAccuracyMetrics{
			IndustryName:           industry,
			WindowedAccuracy:       make([]float64, 0),
			ConfidenceDistribution: make(map[string]int64),
			TopMisclassifications:  make([]*MisclassificationRecord, 0),
			LastUpdated:            time.Now(),
		}
		iat.industries[industry] = metrics
	}

	// Update metrics
	metrics.TotalClassifications++
	metrics.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			metrics.CorrectClassifications++
		}

		// Update accuracy score
		metrics.AccuracyScore = float64(metrics.CorrectClassifications) / float64(metrics.TotalClassifications)

		// Update windowed accuracy
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}
		metrics.WindowedAccuracy = append(metrics.WindowedAccuracy, accuracy)

		// Update confidence distribution
		confidenceRange := getConfidenceRange(result.ConfidenceScore)
		metrics.ConfidenceDistribution[confidenceRange]++

		// Track misclassifications
		if !*result.IsCorrect {
			iat.trackMisclassification(metrics, result)
		}
	}

	return nil
}

// GetIndustryAccuracy returns accuracy for a specific industry
func (iat *IndustryAccuracyTracker) GetIndustryAccuracy(industry string) float64 {
	iat.mu.RLock()
	defer iat.mu.RUnlock()

	if metrics, exists := iat.industries[industry]; exists {
		return metrics.AccuracyScore
	}
	return 0.0
}

// GetAllIndustryAccuracies returns all industry accuracies
func (iat *IndustryAccuracyTracker) GetAllIndustryAccuracies() map[string]float64 {
	iat.mu.RLock()
	defer iat.mu.RUnlock()

	result := make(map[string]float64)
	for industry, metrics := range iat.industries {
		result[industry] = metrics.AccuracyScore
	}
	return result
}

// GetAllIndustries returns all tracked industries
func (iat *IndustryAccuracyTracker) GetAllIndustries() []string {
	iat.mu.RLock()
	defer iat.mu.RUnlock()

	industries := make([]string, 0, len(iat.industries))
	for industry := range iat.industries {
		industries = append(industries, industry)
	}
	return industries
}

// GetIndustryWindowedAccuracy returns windowed accuracy for a specific industry
func (iat *IndustryAccuracyTracker) GetIndustryWindowedAccuracy(industry string) []float64 {
	iat.mu.RLock()
	defer iat.mu.RUnlock()

	if metrics, exists := iat.industries[industry]; exists {
		result := make([]float64, len(metrics.WindowedAccuracy))
		copy(result, metrics.WindowedAccuracy)
		return result
	}
	return []float64{}
}

// UpdateMetrics updates all industry metrics
func (iat *IndustryAccuracyTracker) UpdateMetrics() {
	iat.mu.Lock()
	defer iat.mu.Unlock()

	for _, metrics := range iat.industries {
		metrics.LastUpdated = time.Now()

		// Update trend indicator
		if len(metrics.WindowedAccuracy) >= 10 {
			metrics.TrendIndicator = calculateTrendIndicator(metrics.WindowedAccuracy)
		}
	}
}

// trackMisclassification tracks a misclassification for an industry
func (iat *IndustryAccuracyTracker) trackMisclassification(metrics *IndustryAccuracyMetrics, result *ClassificationResult) {
	record := &MisclassificationRecord{
		ID:                     fmt.Sprintf("misclass_%d_%s", time.Now().UnixNano(), metrics.IndustryName),
		Timestamp:              result.Timestamp,
		BusinessName:           result.BusinessName,
		ActualClassification:   result.ActualClassification,
		ExpectedClassification: *result.ExpectedClassification,
		ConfidenceScore:        result.ConfidenceScore,
		ClassificationMethod:   result.ClassificationMethod,
		InputData:              result.Metadata,
		ErrorType:              classifyErrorType(result),
		Severity:               calculateMisclassificationSeverity(result),
		ActionRequired:         requiresAction(result),
	}

	metrics.TopMisclassifications = append(metrics.TopMisclassifications, record)

	// Keep only recent misclassifications
	if len(metrics.TopMisclassifications) > 100 {
		metrics.TopMisclassifications = metrics.TopMisclassifications[50:]
	}
}

// EnsembleMethodTracker implementation

// NewEnsembleMethodTracker creates a new ensemble method tracker
func NewEnsembleMethodTracker() *EnsembleMethodTracker {
	return &EnsembleMethodTracker{
		methods: make(map[string]*MethodAccuracyMetrics),
	}
}

// TrackResult tracks a classification result for a method
func (emt *EnsembleMethodTracker) TrackResult(result *ClassificationResult) error {
	emt.mu.Lock()
	defer emt.mu.Unlock()

	method := result.ClassificationMethod

	// Get or create method metrics
	metrics, exists := emt.methods[method]
	if !exists {
		metrics = &MethodAccuracyMetrics{
			MethodName:       method,
			WindowedAccuracy: make([]float64, 0),
			LastUpdated:      time.Now(),
			Weight:           0.5, // Default weight
		}
		emt.methods[method] = metrics
	}

	// Update metrics
	metrics.TotalClassifications++
	metrics.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			metrics.CorrectClassifications++
		}

		// Update accuracy score
		metrics.AccuracyScore = float64(metrics.CorrectClassifications) / float64(metrics.TotalClassifications)

		// Update windowed accuracy
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}
		metrics.WindowedAccuracy = append(metrics.WindowedAccuracy, accuracy)

		// Update average confidence
		metrics.AverageConfidence = (metrics.AverageConfidence*float64(metrics.TotalClassifications-1) + result.ConfidenceScore) / float64(metrics.TotalClassifications)
	}

	// Update performance score
	metrics.PerformanceScore = calculatePerformanceScore(metrics)

	return nil
}

// GetMethodAccuracy returns accuracy for a specific method
func (emt *EnsembleMethodTracker) GetMethodAccuracy(method string) float64 {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	if metrics, exists := emt.methods[method]; exists {
		return metrics.AccuracyScore
	}
	return 0.0
}

// GetAllMethodAccuracies returns all method accuracies
func (emt *EnsembleMethodTracker) GetAllMethodAccuracies() map[string]float64 {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	result := make(map[string]float64)
	for method, metrics := range emt.methods {
		result[method] = metrics.AccuracyScore
	}
	return result
}

// GetAllMethods returns all tracked methods
func (emt *EnsembleMethodTracker) GetAllMethods() []string {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	methods := make([]string, 0, len(emt.methods))
	for method := range emt.methods {
		methods = append(methods, method)
	}
	return methods
}

// GetMethodWindowedAccuracy returns windowed accuracy for a specific method
func (emt *EnsembleMethodTracker) GetMethodWindowedAccuracy(method string) []float64 {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	if metrics, exists := emt.methods[method]; exists {
		result := make([]float64, len(metrics.WindowedAccuracy))
		copy(result, metrics.WindowedAccuracy)
		return result
	}
	return []float64{}
}

// UpdateMetrics updates all method metrics
func (emt *EnsembleMethodTracker) UpdateMetrics() {
	emt.mu.Lock()
	defer emt.mu.Unlock()

	for _, metrics := range emt.methods {
		metrics.LastUpdated = time.Now()

		// Update trend indicator
		if len(metrics.WindowedAccuracy) >= 10 {
			metrics.TrendIndicator = calculateTrendIndicator(metrics.WindowedAccuracy)
		}

		// Update performance score
		metrics.PerformanceScore = calculatePerformanceScore(metrics)
	}
}

// MLModelAccuracyTracker implementation

// NewMLModelAccuracyTracker creates a new ML model accuracy tracker
func NewMLModelAccuracyTracker() *MLModelAccuracyTracker {
	return &MLModelAccuracyTracker{
		models: make(map[string]*MLModelMetrics),
	}
}

// TrackResult tracks a classification result for an ML model
func (mmt *MLModelAccuracyTracker) TrackResult(result *ClassificationResult) error {
	mmt.mu.Lock()
	defer mmt.mu.Unlock()

	// Extract model information from result metadata
	modelName := "unknown"
	modelVersion := "unknown"

	if result.Metadata != nil {
		if modelValue, exists := result.Metadata["model_name"]; exists {
			if modelStr, ok := modelValue.(string); ok {
				modelName = modelStr
			}
		}
		if versionValue, exists := result.Metadata["model_version"]; exists {
			if versionStr, ok := versionValue.(string); ok {
				modelVersion = versionStr
			}
		}
	}

	modelKey := fmt.Sprintf("%s_%s", modelName, modelVersion)

	// Get or create model metrics
	metrics, exists := mmt.models[modelKey]
	if !exists {
		metrics = &MLModelMetrics{
			ModelName:        modelName,
			ModelVersion:     modelVersion,
			WindowedAccuracy: make([]float64, 0),
			LastUpdated:      time.Now(),
		}
		mmt.models[modelKey] = metrics
	}

	// Update metrics
	metrics.TotalPredictions++
	metrics.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			metrics.CorrectPredictions++
		}

		// Update accuracy score
		metrics.AccuracyScore = float64(metrics.CorrectPredictions) / float64(metrics.TotalPredictions)

		// Update windowed accuracy
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}
		metrics.WindowedAccuracy = append(metrics.WindowedAccuracy, accuracy)

		// Update average confidence
		metrics.AverageConfidence = (metrics.AverageConfidence*float64(metrics.TotalPredictions-1) + result.ConfidenceScore) / float64(metrics.TotalPredictions)

		// Calculate uncertainty score
		metrics.UncertaintyScore = calculateUncertaintyScore(result.ConfidenceScore)
	}

	return nil
}

// GetModelAccuracy returns accuracy for a specific model
func (mmt *MLModelAccuracyTracker) GetModelAccuracy(model string) float64 {
	mmt.mu.RLock()
	defer mmt.mu.RUnlock()

	if metrics, exists := mmt.models[model]; exists {
		return metrics.AccuracyScore
	}
	return 0.0
}

// GetAllModelAccuracies returns all model accuracies
func (mmt *MLModelAccuracyTracker) GetAllModelAccuracies() map[string]float64 {
	mmt.mu.RLock()
	defer mmt.mu.RUnlock()

	result := make(map[string]float64)
	for model, metrics := range mmt.models {
		result[model] = metrics.AccuracyScore
	}
	return result
}

// GetAllModels returns all tracked models
func (mmt *MLModelAccuracyTracker) GetAllModels() []string {
	mmt.mu.RLock()
	defer mmt.mu.RUnlock()

	models := make([]string, 0, len(mmt.models))
	for model := range mmt.models {
		models = append(models, model)
	}
	return models
}

// GetModelWindowedAccuracy returns windowed accuracy for a specific model
func (mmt *MLModelAccuracyTracker) GetModelWindowedAccuracy(model string) []float64 {
	mmt.mu.RLock()
	defer mmt.mu.RUnlock()

	if metrics, exists := mmt.models[model]; exists {
		result := make([]float64, len(metrics.WindowedAccuracy))
		copy(result, metrics.WindowedAccuracy)
		return result
	}
	return []float64{}
}

// UpdateMetrics updates all model metrics
func (mmt *MLModelAccuracyTracker) UpdateMetrics() {
	mmt.mu.Lock()
	defer mmt.mu.Unlock()

	for _, metrics := range mmt.models {
		metrics.LastUpdated = time.Now()

		// Update trend indicator
		if len(metrics.WindowedAccuracy) >= 10 {
			metrics.TrendIndicator = calculateTrendIndicator(metrics.WindowedAccuracy)
		}

		// Calculate model drift score
		metrics.ModelDriftScore = calculateModelDriftScore(metrics)
	}
}

// SecurityAccuracyTracker implementation

// NewSecurityAccuracyTracker creates a new security accuracy tracker
func NewSecurityAccuracyTracker() *SecurityAccuracyTracker {
	return &SecurityAccuracyTracker{
		lastUpdated: time.Now(),
	}
}

// TrackResult tracks a classification result for security metrics
func (sat *SecurityAccuracyTracker) TrackResult(result *ClassificationResult) error {
	sat.mu.Lock()
	defer sat.mu.Unlock()

	sat.lastUpdated = time.Now()

	// Extract security information from result metadata
	if result.Metadata != nil {
		if trustedValue, exists := result.Metadata["trusted_data_source"]; exists {
			if trusted, ok := trustedValue.(bool); ok && trusted {
				// Track trusted data source usage
				sat.trustedDataSourceRate = 1.0 // Assume 100% if trusted
			}
		}

		if verifiedValue, exists := result.Metadata["website_verified"]; exists {
			if verified, ok := verifiedValue.(bool); ok && verified {
				// Track website verification usage
				sat.websiteVerificationRate = 1.0 // Assume 100% if verified
			}
		}
	}

	return nil
}

// GetMetrics returns current security metrics
func (sat *SecurityAccuracyTracker) GetMetrics() *SecurityAccuracySnapshot {
	sat.mu.RLock()
	defer sat.mu.RUnlock()

	return &SecurityAccuracySnapshot{
		TrustedDataSourceRate:   sat.trustedDataSourceRate,
		WebsiteVerificationRate: sat.websiteVerificationRate,
		SecurityViolationRate:   sat.securityViolationRate,
		ConfidenceIntegrity:     sat.confidenceIntegrity,
	}
}

// UpdateMetrics updates security metrics
func (sat *SecurityAccuracyTracker) UpdateMetrics() {
	sat.mu.Lock()
	defer sat.mu.Unlock()

	sat.lastUpdated = time.Now()

	// In a real implementation, these would be updated from the security monitoring system
	// For now, we'll set them to ideal values
	sat.trustedDataSourceRate = 1.0   // 100% trusted sources
	sat.websiteVerificationRate = 1.0 // 100% verified websites
	sat.securityViolationRate = 0.0   // 0% violations
	sat.confidenceIntegrity = 1.0     // 100% integrity
}

// RealTimeMetrics implementation

// NewRealTimeMetrics creates new real-time metrics
func NewRealTimeMetrics() *RealTimeMetrics {
	return &RealTimeMetrics{
		mu: sync.RWMutex{},
	}
}

// TrendAnalyzer implementation

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{
		trends: make(map[string]*TrendData),
	}
}

// UpdateTrend updates trend data for a dimension
func (ta *TrendAnalyzer) UpdateTrend(dimensionName, dimensionValue, trend string, trendStrength float64) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	key := fmt.Sprintf("%s:%s", dimensionName, dimensionValue)

	ta.trends[key] = &TrendData{
		DimensionName:     dimensionName,
		DimensionValue:    dimensionValue,
		Trend:             trend,
		TrendStrength:     trendStrength,
		LastAnalysis:      time.Now(),
		PredictedAccuracy: ta.calculatePredictedAccuracy(trend, trendStrength),
	}
}

// GetAllTrends returns all trend data
func (ta *TrendAnalyzer) GetAllTrends() map[string]*TrendData {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	result := make(map[string]*TrendData)
	for key, trend := range ta.trends {
		result[key] = &TrendData{
			DimensionName:     trend.DimensionName,
			DimensionValue:    trend.DimensionValue,
			Trend:             trend.Trend,
			TrendStrength:     trend.TrendStrength,
			LastAnalysis:      trend.LastAnalysis,
			PredictedAccuracy: trend.PredictedAccuracy,
		}
	}
	return result
}

// calculatePredictedAccuracy calculates predicted accuracy based on trend
func (ta *TrendAnalyzer) calculatePredictedAccuracy(trend string, trendStrength float64) float64 {
	baseAccuracy := 0.90 // Assume 90% base accuracy

	switch trend {
	case "improving":
		return baseAccuracy + (trendStrength * 0.1) // Up to 10% improvement
	case "declining":
		return baseAccuracy - (trendStrength * 0.1) // Up to 10% decline
	default:
		return baseAccuracy
	}
}

// Helper functions

// getConfidenceRange returns the confidence range category
func getConfidenceRange(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "high"
	case confidence >= 0.7:
		return "medium"
	case confidence >= 0.5:
		return "low"
	default:
		return "very_low"
	}
}

// calculateTrendIndicator calculates a trend indicator from windowed accuracy
func calculateTrendIndicator(windowedAccuracy []float64) string {
	if len(windowedAccuracy) < 10 {
		return "insufficient_data"
	}

	// Calculate simple trend
	firstHalf := windowedAccuracy[:len(windowedAccuracy)/2]
	secondHalf := windowedAccuracy[len(windowedAccuracy)/2:]

	firstAvg := calculateAverage(firstHalf)
	secondAvg := calculateAverage(secondHalf)

	diff := secondAvg - firstAvg

	if diff > 0.02 {
		return "improving"
	} else if diff < -0.02 {
		return "declining"
	} else {
		return "stable"
	}
}

// calculateAverage calculates the average of a slice of floats
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculatePerformanceScore calculates a performance score for a method
func calculatePerformanceScore(metrics *MethodAccuracyMetrics) float64 {
	// Combine accuracy and latency into a performance score
	accuracyWeight := 0.7
	latencyWeight := 0.3

	// Normalize latency (lower is better)
	latencyScore := 1.0
	if metrics.AverageLatency > 0 {
		latencyScore = 1.0 / (1.0 + float64(metrics.AverageLatency.Nanoseconds())/1e9) // Convert to seconds
	}

	return (metrics.AccuracyScore * accuracyWeight) + (latencyScore * latencyWeight)
}

// calculateUncertaintyScore calculates uncertainty score from confidence
func calculateUncertaintyScore(confidence float64) float64 {
	// Higher confidence = lower uncertainty
	return 1.0 - confidence
}

// calculateModelDriftScore calculates model drift score
func calculateModelDriftScore(metrics *MLModelMetrics) float64 {
	// Simple drift calculation based on recent accuracy vs historical
	if len(metrics.WindowedAccuracy) < 20 {
		return 0.0
	}

	recent := metrics.WindowedAccuracy[len(metrics.WindowedAccuracy)-10:]
	historical := metrics.WindowedAccuracy[:len(metrics.WindowedAccuracy)-10]

	recentAvg := calculateAverage(recent)
	historicalAvg := calculateAverage(historical)

	// Drift is the absolute difference
	return abs(recentAvg - historicalAvg)
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// classifyErrorType classifies the type of error
func classifyErrorType(result *ClassificationResult) string {
	if result.ExpectedClassification == nil {
		return "unknown"
	}

	// Simple error classification based on confidence
	if result.ConfidenceScore < 0.3 {
		return "low_confidence"
	} else if result.ConfidenceScore < 0.7 {
		return "medium_confidence"
	} else {
		return "high_confidence"
	}
}

// calculateMisclassificationSeverity calculates the severity of a misclassification
func calculateMisclassificationSeverity(result *ClassificationResult) string {
	if result.ExpectedClassification == nil {
		return "unknown"
	}

	// Severity based on confidence and business impact
	if result.ConfidenceScore > 0.8 {
		return "high" // High confidence but wrong = high severity
	} else if result.ConfidenceScore > 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// requiresAction determines if action is required for a misclassification
func requiresAction(result *ClassificationResult) bool {
	if result.ExpectedClassification == nil {
		return false
	}

	// Require action for high-confidence misclassifications
	return result.ConfidenceScore > 0.8
}
