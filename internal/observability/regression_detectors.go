package observability

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// StatisticalDetector implements statistical regression detection
type StatisticalDetector struct {
	config RegressionDetectionConfig
	logger *zap.Logger
}

// NewStatisticalDetector creates a new statistical detector
func NewStatisticalDetector(config RegressionDetectionConfig) *StatisticalDetector {
	return &StatisticalDetector{
		config: config,
		logger: zap.NewNop(),
	}
}

// Name returns the detector name
func (sd *StatisticalDetector) Name() string {
	return "statistical_detector"
}

// Type returns the detector type
func (sd *StatisticalDetector) Type() string {
	return "statistical"
}

// Detect performs statistical regression detection
func (sd *StatisticalDetector) Detect(baseline *PerformanceBaseline, currentData []*PerformanceDataPoint) (*RegressionResult, error) {
	if len(currentData) < 10 {
		return nil, fmt.Errorf("insufficient data for statistical detection: need at least 10 points, got %d", len(currentData))
	}

	// Extract current values
	currentValues := sd.extractMetricValues(baseline.Metric, currentData)
	if len(currentValues) == 0 {
		return nil, fmt.Errorf("no valid values found for metric %s", baseline.Metric)
	}

	// Calculate current statistics
	currentMean := sd.calculateMean(currentValues)
	currentStdDev := sd.calculateStdDev(currentValues, currentMean)

	// Calculate change percentage
	changePercent := ((currentMean - baseline.Mean) / baseline.Mean) * 100

	// Determine change direction
	changeDirection := "stable"
	if changePercent > 0 {
		changeDirection = "increase"
	} else if changePercent < 0 {
		changeDirection = "decrease"
	}

	// Perform t-test for statistical significance
	pValue := sd.performTTest(baseline, currentValues)
	isSignificant := pValue < sd.config.PValueThreshold

	// Determine regression type
	regressionType := "none"
	severity := "low"

	if isSignificant {
		threshold := sd.getDegradationThreshold(baseline.Metric)
		improvementThreshold := sd.getImprovementThreshold(baseline.Metric)

		if math.Abs(changePercent) >= threshold {
			if changePercent > 0 {
				regressionType = "degradation"
				severity = sd.determineSeverity(changePercent, threshold)
			} else if changePercent < 0 && math.Abs(changePercent) >= improvementThreshold {
				regressionType = "improvement"
				severity = "low" // Improvements are typically low severity
			}
		}
	}

	// Create result
	result := &RegressionResult{
		ID:              fmt.Sprintf("reg_%s_%d", baseline.Metric, time.Now().UnixNano()),
		Metric:          baseline.Metric,
		DetectedAt:      time.Now().UTC(),
		Type:            regressionType,
		BaselineMean:    baseline.Mean,
		CurrentMean:     currentMean,
		ChangePercent:   changePercent,
		ChangeDirection: changeDirection,
		PValue:          pValue,
		Confidence:      sd.config.ConfidenceLevel,
		IsSignificant:   isSignificant,
		DetectorUsed:    sd.Name(),
		DetectionTime:   time.Now().UTC(),
		Severity:        severity,
		BaselineID:      baseline.ID,
		BaselineStart:   baseline.SampleStart,
		BaselineEnd:     baseline.SampleEnd,
		CurrentStart:    currentData[0].Timestamp,
		CurrentEnd:      currentData[len(currentData)-1].Timestamp,
		Tags:            make(map[string]string),
	}

	// Add trend analysis
	result.TrendAnalysis = sd.analyzeTrend(currentData, baseline.Metric)

	return result, nil
}

// GetConfidence returns the detector confidence
func (sd *StatisticalDetector) GetConfidence() float64 {
	return sd.config.ConfidenceLevel
}

// IsApplicable checks if the detector is applicable to a metric
func (sd *StatisticalDetector) IsApplicable(metric string) bool {
	// Statistical detector is applicable to all metrics
	return true
}

// performTTest performs a t-test between baseline and current data
func (sd *StatisticalDetector) performTTest(baseline *PerformanceBaseline, currentValues []float64) float64 {
	// Simplified t-test implementation
	// In a real implementation, you'd use a proper statistical library

	// Calculate pooled standard deviation
	n1 := float64(baseline.SampleSize)
	n2 := float64(len(currentValues))

	pooledStdDev := math.Sqrt(((n1-1)*baseline.StdDev*baseline.StdDev + (n2-1)*sd.calculateStdDev(currentValues, sd.calculateMean(currentValues))*sd.calculateStdDev(currentValues, sd.calculateMean(currentValues))) / (n1 + n2 - 2))

	// Calculate t-statistic
	tStat := (sd.calculateMean(currentValues) - baseline.Mean) / (pooledStdDev * math.Sqrt(1/n1+1/n2))

	// Convert to p-value (simplified)
	// In a real implementation, you'd use t-distribution tables or functions
	pValue := 2 * (1 - sd.normalCDF(math.Abs(tStat)))

	return pValue
}

// normalCDF calculates the cumulative distribution function of the normal distribution
func (sd *StatisticalDetector) normalCDF(x float64) float64 {
	// Simplified normal CDF approximation
	return 0.5 * (1 + sd.erf(x/math.Sqrt(2)))
}

// erf calculates the error function
func (sd *StatisticalDetector) erf(x float64) float64 {
	// Simplified error function approximation
	if x < 0 {
		return -sd.erf(-x)
	}

	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	t := 1.0 / (1.0 + p*x)
	return 1 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)
}

// getDegradationThreshold gets the degradation threshold for a metric
func (sd *StatisticalDetector) getDegradationThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return sd.config.RegressionThresholds.ResponseTime.Degradation
	case "success_rate":
		return sd.config.RegressionThresholds.SuccessRate.Degradation
	case "throughput":
		return sd.config.RegressionThresholds.Throughput.Degradation
	case "error_rate":
		return sd.config.RegressionThresholds.ErrorRate.Degradation
	case "cpu_usage":
		return sd.config.RegressionThresholds.ResourceUtilization.CPU.Degradation
	case "memory_usage":
		return sd.config.RegressionThresholds.ResourceUtilization.Memory.Degradation
	default:
		return 10.0 // Default 10% threshold
	}
}

// getImprovementThreshold gets the improvement threshold for a metric
func (sd *StatisticalDetector) getImprovementThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return sd.config.RegressionThresholds.ResponseTime.Improvement
	case "success_rate":
		return sd.config.RegressionThresholds.SuccessRate.Improvement
	case "throughput":
		return sd.config.RegressionThresholds.Throughput.Improvement
	case "error_rate":
		return sd.config.RegressionThresholds.ErrorRate.Improvement
	case "cpu_usage":
		return sd.config.RegressionThresholds.ResourceUtilization.CPU.Improvement
	case "memory_usage":
		return sd.config.RegressionThresholds.ResourceUtilization.Memory.Improvement
	default:
		return 5.0 // Default 5% threshold
	}
}

// determineSeverity determines the severity based on change percentage
func (sd *StatisticalDetector) determineSeverity(changePercent, threshold float64) string {
	ratio := math.Abs(changePercent) / threshold

	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	} else {
		return "low"
	}
}

// analyzeTrend analyzes the trend in current data
func (sd *StatisticalDetector) analyzeTrend(data []*PerformanceDataPoint, metric string) *TrendAnalysis {
	if len(data) < 5 {
		return &TrendAnalysis{
			TrendDirection: "stable",
			TrendStrength:  0.0,
			Slope:          0.0,
			R2:             0.0,
			PValue:         1.0,
		}
	}

	// Extract values and time points
	values := sd.extractMetricValues(metric, data)
	times := make([]float64, len(data))
	for i, point := range data {
		times[i] = float64(point.Timestamp.Unix())
	}

	// Calculate linear regression
	slope, intercept, r2 := sd.calculateLinearRegression(times, values)

	// Determine trend direction
	trendDirection := "stable"
	if math.Abs(slope) > 0.001 {
		if slope > 0 {
			trendDirection = "increasing"
		} else {
			trendDirection = "decreasing"
		}
	}

	// Calculate trend strength (0-1)
	trendStrength := math.Min(math.Abs(slope)/1000.0, 1.0) // Normalize slope

	return &TrendAnalysis{
		TrendDirection: trendDirection,
		TrendStrength:  trendStrength,
		Slope:          slope,
		R2:             r2,
		PValue:         0.05, // Placeholder
	}
}

// calculateLinearRegression calculates linear regression parameters
func (sd *StatisticalDetector) calculateLinearRegression(x, y []float64) (slope, intercept, r2 float64) {
	n := float64(len(x))
	if n < 2 {
		return 0, 0, 0
	}

	// Calculate means
	xMean := sd.calculateMean(x)
	yMean := sd.calculateMean(y)

	// Calculate slope and intercept
	var numerator, denominator float64
	for i := 0; i < len(x); i++ {
		numerator += (x[i] - xMean) * (y[i] - yMean)
		denominator += (x[i] - xMean) * (x[i] - xMean)
	}

	if denominator == 0 {
		return 0, yMean, 0
	}

	slope = numerator / denominator
	intercept = yMean - slope*xMean

	// Calculate R-squared
	var ssRes, ssTot float64
	for i := 0; i < len(x); i++ {
		yPred := slope*x[i] + intercept
		ssRes += (y[i] - yPred) * (y[i] - yPred)
		ssTot += (y[i] - yMean) * (y[i] - yMean)
	}

	if ssTot == 0 {
		r2 = 0
	} else {
		r2 = 1 - (ssRes / ssTot)
	}

	return slope, intercept, r2
}

// extractMetricValues extracts metric values from data points
func (sd *StatisticalDetector) extractMetricValues(metric string, data []*PerformanceDataPoint) []float64 {
	values := make([]float64, 0, len(data))

	for _, point := range data {
		var value float64
		switch metric {
		case "response_time":
			value = float64(point.ResponseTime.Milliseconds())
		case "success_rate":
			value = point.SuccessRate
		case "throughput":
			value = point.Throughput
		case "error_rate":
			value = point.ErrorRate
		case "cpu_usage":
			value = point.CPUUsage
		case "memory_usage":
			value = point.MemoryUsage
		default:
			continue
		}
		values = append(values, value)
	}

	return values
}

// calculateMean calculates the mean of values
func (sd *StatisticalDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation of values
func (sd *StatisticalDetector) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSq := 0.0
	for _, value := range values {
		sumSq += (value - mean) * (value - mean)
	}

	variance := sumSq / float64(len(values)-1)
	return math.Sqrt(variance)
}

// TrendDetector implements trend-based regression detection
type TrendDetector struct {
	config RegressionDetectionConfig
	logger *zap.Logger
}

// NewTrendDetector creates a new trend detector
func NewTrendDetector(config RegressionDetectionConfig) *TrendDetector {
	return &TrendDetector{
		config: config,
		logger: zap.NewNop(),
	}
}

// Name returns the detector name
func (td *TrendDetector) Name() string {
	return "trend_detector"
}

// Type returns the detector type
func (td *TrendDetector) Type() string {
	return "trend"
}

// Detect performs trend-based regression detection
func (td *TrendDetector) Detect(baseline *PerformanceBaseline, currentData []*PerformanceDataPoint) (*RegressionResult, error) {
	if len(currentData) < 20 {
		return nil, fmt.Errorf("insufficient data for trend detection: need at least 20 points, got %d", len(currentData))
	}

	// Extract current values
	currentValues := td.extractMetricValues(baseline.Metric, currentData)
	if len(currentValues) == 0 {
		return nil, fmt.Errorf("no valid values found for metric %s", baseline.Metric)
	}

	// Calculate trend
	times := make([]float64, len(currentData))
	for i, point := range currentData {
		times[i] = float64(point.Timestamp.Unix())
	}

	slope, _, r2 := td.calculateLinearRegression(times, currentValues)

	// Determine if trend is significant
	isSignificant := r2 > 0.7 && math.Abs(slope) > 0.001

	// Calculate change over the detection window
	timeSpan := times[len(times)-1] - times[0]
	changeAmount := slope * timeSpan
	changePercent := (changeAmount / baseline.Mean) * 100

	// Determine regression type
	regressionType := "none"
	severity := "low"

	if isSignificant {
		threshold := td.getDegradationThreshold(baseline.Metric)
		improvementThreshold := td.getImprovementThreshold(baseline.Metric)

		if math.Abs(changePercent) >= threshold {
			if changePercent > 0 {
				regressionType = "degradation"
				severity = td.determineSeverity(changePercent, threshold)
			} else if changePercent < 0 && math.Abs(changePercent) >= improvementThreshold {
				regressionType = "improvement"
				severity = "low"
			}
		}
	}

	// Create result
	result := &RegressionResult{
		ID:              fmt.Sprintf("trend_%s_%d", baseline.Metric, time.Now().UnixNano()),
		Metric:          baseline.Metric,
		DetectedAt:      time.Now().UTC(),
		Type:            regressionType,
		BaselineMean:    baseline.Mean,
		CurrentMean:     td.calculateMean(currentValues),
		ChangePercent:   changePercent,
		ChangeDirection: td.getChangeDirection(slope),
		PValue:          0.05, // Placeholder
		Confidence:      r2,
		IsSignificant:   isSignificant,
		DetectorUsed:    td.Name(),
		DetectionTime:   time.Now().UTC(),
		Severity:        severity,
		BaselineID:      baseline.ID,
		BaselineStart:   baseline.SampleStart,
		BaselineEnd:     baseline.SampleEnd,
		CurrentStart:    currentData[0].Timestamp,
		CurrentEnd:      currentData[len(currentData)-1].Timestamp,
		Tags:            make(map[string]string),
	}

	// Add trend analysis
	result.TrendAnalysis = &TrendAnalysis{
		TrendDirection: td.getTrendDirection(slope),
		TrendStrength:  math.Min(math.Abs(slope)/1000.0, 1.0),
		Slope:          slope,
		R2:             r2,
		PValue:         0.05,
	}

	return result, nil
}

// GetConfidence returns the detector confidence
func (td *TrendDetector) GetConfidence() float64 {
	return 0.8 // Trend detector has moderate confidence
}

// IsApplicable checks if the detector is applicable to a metric
func (td *TrendDetector) IsApplicable(metric string) bool {
	// Trend detector is applicable to all metrics
	return true
}

// getChangeDirection gets the change direction based on slope
func (td *TrendDetector) getChangeDirection(slope float64) string {
	if math.Abs(slope) < 0.001 {
		return "stable"
	} else if slope > 0 {
		return "increase"
	} else {
		return "decrease"
	}
}

// getTrendDirection gets the trend direction based on slope
func (td *TrendDetector) getTrendDirection(slope float64) string {
	if math.Abs(slope) < 0.001 {
		return "stable"
	} else if slope > 0 {
		return "increasing"
	} else {
		return "decreasing"
	}
}

// getDegradationThreshold gets the degradation threshold for a metric
func (td *TrendDetector) getDegradationThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return td.config.RegressionThresholds.ResponseTime.Degradation
	case "success_rate":
		return td.config.RegressionThresholds.SuccessRate.Degradation
	case "throughput":
		return td.config.RegressionThresholds.Throughput.Degradation
	case "error_rate":
		return td.config.RegressionThresholds.ErrorRate.Degradation
	case "cpu_usage":
		return td.config.RegressionThresholds.ResourceUtilization.CPU.Degradation
	case "memory_usage":
		return td.config.RegressionThresholds.ResourceUtilization.Memory.Degradation
	default:
		return 10.0
	}
}

// getImprovementThreshold gets the improvement threshold for a metric
func (td *TrendDetector) getImprovementThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return td.config.RegressionThresholds.ResponseTime.Improvement
	case "success_rate":
		return td.config.RegressionThresholds.SuccessRate.Improvement
	case "throughput":
		return td.config.RegressionThresholds.Throughput.Improvement
	case "error_rate":
		return td.config.RegressionThresholds.ErrorRate.Improvement
	case "cpu_usage":
		return td.config.RegressionThresholds.ResourceUtilization.CPU.Improvement
	case "memory_usage":
		return td.config.RegressionThresholds.ResourceUtilization.Memory.Improvement
	default:
		return 5.0
	}
}

// determineSeverity determines the severity based on change percentage
func (td *TrendDetector) determineSeverity(changePercent, threshold float64) string {
	ratio := math.Abs(changePercent) / threshold

	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	} else {
		return "low"
	}
}

// extractMetricValues extracts metric values from data points
func (td *TrendDetector) extractMetricValues(metric string, data []*PerformanceDataPoint) []float64 {
	values := make([]float64, 0, len(data))

	for _, point := range data {
		var value float64
		switch metric {
		case "response_time":
			value = float64(point.ResponseTime.Milliseconds())
		case "success_rate":
			value = point.SuccessRate
		case "throughput":
			value = point.Throughput
		case "error_rate":
			value = point.ErrorRate
		case "cpu_usage":
			value = point.CPUUsage
		case "memory_usage":
			value = point.MemoryUsage
		default:
			continue
		}
		values = append(values, value)
	}

	return values
}

// calculateLinearRegression calculates linear regression parameters
func (td *TrendDetector) calculateLinearRegression(x, y []float64) (slope, intercept, r2 float64) {
	n := float64(len(x))
	if n < 2 {
		return 0, 0, 0
	}

	// Calculate means
	xMean := td.calculateMean(x)
	yMean := td.calculateMean(y)

	// Calculate slope and intercept
	var numerator, denominator float64
	for i := 0; i < len(x); i++ {
		numerator += (x[i] - xMean) * (y[i] - yMean)
		denominator += (x[i] - xMean) * (x[i] - xMean)
	}

	if denominator == 0 {
		return 0, yMean, 0
	}

	slope = numerator / denominator
	intercept = yMean - slope*xMean

	// Calculate R-squared
	var ssRes, ssTot float64
	for i := 0; i < len(x); i++ {
		yPred := slope*x[i] + intercept
		ssRes += (y[i] - yPred) * (y[i] - yPred)
		ssTot += (y[i] - yMean) * (y[i] - yMean)
	}

	if ssTot == 0 {
		r2 = 0
	} else {
		r2 = 1 - (ssRes / ssTot)
	}

	return slope, intercept, r2
}

// calculateMean calculates the mean of values
func (td *TrendDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// ThresholdDetector implements threshold-based regression detection
type ThresholdDetector struct {
	config RegressionDetectionConfig
	logger *zap.Logger
}

// NewThresholdDetector creates a new threshold detector
func NewThresholdDetector(config RegressionDetectionConfig) *ThresholdDetector {
	return &ThresholdDetector{
		config: config,
		logger: zap.NewNop(),
	}
}

// Name returns the detector name
func (thd *ThresholdDetector) Name() string {
	return "threshold_detector"
}

// Type returns the detector type
func (thd *ThresholdDetector) Type() string {
	return "threshold"
}

// Detect performs threshold-based regression detection
func (thd *ThresholdDetector) Detect(baseline *PerformanceBaseline, currentData []*PerformanceDataPoint) (*RegressionResult, error) {
	if len(currentData) < 5 {
		return nil, fmt.Errorf("insufficient data for threshold detection: need at least 5 points, got %d", len(currentData))
	}

	// Extract current values
	currentValues := thd.extractMetricValues(baseline.Metric, currentData)
	if len(currentValues) == 0 {
		return nil, fmt.Errorf("no valid values found for metric %s", baseline.Metric)
	}

	// Calculate current statistics
	currentMean := thd.calculateMean(currentValues)
	currentMax := thd.calculateMax(currentValues)
	currentMin := thd.calculateMin(currentValues)

	// Check against thresholds
	regressionType := "none"
	severity := "low"
	changePercent := 0.0

	// Check if current mean exceeds baseline thresholds
	upperThreshold := baseline.Mean + 2*baseline.StdDev // 95% confidence interval
	lowerThreshold := baseline.Mean - 2*baseline.StdDev

	if currentMean > upperThreshold {
		regressionType = "degradation"
		changePercent = ((currentMean - baseline.Mean) / baseline.Mean) * 100
		severity = thd.determineSeverity(changePercent, thd.getDegradationThreshold(baseline.Metric))
	} else if currentMean < lowerThreshold {
		regressionType = "improvement"
		changePercent = ((currentMean - baseline.Mean) / baseline.Mean) * 100
		severity = "low"
	}

	// Check for extreme values
	if currentMax > baseline.Percentile99 {
		regressionType = "degradation"
		changePercent = ((currentMax - baseline.Mean) / baseline.Mean) * 100
		severity = "high"
	}

	// Create result
	result := &RegressionResult{
		ID:              fmt.Sprintf("threshold_%s_%d", baseline.Metric, time.Now().UnixNano()),
		Metric:          baseline.Metric,
		DetectedAt:      time.Now().UTC(),
		Type:            regressionType,
		BaselineMean:    baseline.Mean,
		CurrentMean:     currentMean,
		ChangePercent:   changePercent,
		ChangeDirection: thd.getChangeDirection(currentMean, baseline.Mean),
		PValue:          0.05, // Placeholder
		Confidence:      0.9,  // High confidence for threshold-based detection
		IsSignificant:   regressionType != "none",
		DetectorUsed:    thd.Name(),
		DetectionTime:   time.Now().UTC(),
		Severity:        severity,
		BaselineID:      baseline.ID,
		BaselineStart:   baseline.SampleStart,
		BaselineEnd:     baseline.SampleEnd,
		CurrentStart:    currentData[0].Timestamp,
		CurrentEnd:      currentData[len(currentData)-1].Timestamp,
		Tags:            make(map[string]string),
	}

	return result, nil
}

// GetConfidence returns the detector confidence
func (thd *ThresholdDetector) GetConfidence() float64 {
	return 0.9 // Threshold detector has high confidence
}

// IsApplicable checks if the detector is applicable to a metric
func (thd *ThresholdDetector) IsApplicable(metric string) bool {
	// Threshold detector is applicable to all metrics
	return true
}

// getChangeDirection gets the change direction
func (thd *ThresholdDetector) getChangeDirection(current, baseline float64) string {
	if math.Abs(current-baseline) < 0.001 {
		return "stable"
	} else if current > baseline {
		return "increase"
	} else {
		return "decrease"
	}
}

// getDegradationThreshold gets the degradation threshold for a metric
func (thd *ThresholdDetector) getDegradationThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return thd.config.RegressionThresholds.ResponseTime.Degradation
	case "success_rate":
		return thd.config.RegressionThresholds.SuccessRate.Degradation
	case "throughput":
		return thd.config.RegressionThresholds.Throughput.Degradation
	case "error_rate":
		return thd.config.RegressionThresholds.ErrorRate.Degradation
	case "cpu_usage":
		return thd.config.RegressionThresholds.ResourceUtilization.CPU.Degradation
	case "memory_usage":
		return thd.config.RegressionThresholds.ResourceUtilization.Memory.Degradation
	default:
		return 10.0
	}
}

// determineSeverity determines the severity based on change percentage
func (thd *ThresholdDetector) determineSeverity(changePercent, threshold float64) string {
	ratio := math.Abs(changePercent) / threshold

	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	} else {
		return "low"
	}
}

// extractMetricValues extracts metric values from data points
func (thd *ThresholdDetector) extractMetricValues(metric string, data []*PerformanceDataPoint) []float64 {
	values := make([]float64, 0, len(data))

	for _, point := range data {
		var value float64
		switch metric {
		case "response_time":
			value = float64(point.ResponseTime.Milliseconds())
		case "success_rate":
			value = point.SuccessRate
		case "throughput":
			value = point.Throughput
		case "error_rate":
			value = point.ErrorRate
		case "cpu_usage":
			value = point.CPUUsage
		case "memory_usage":
			value = point.MemoryUsage
		default:
			continue
		}
		values = append(values, value)
	}

	return values
}

// calculateMean calculates the mean of values
func (thd *ThresholdDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateMax calculates the maximum value
func (thd *ThresholdDetector) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	max := values[0]
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

// calculateMin calculates the minimum value
func (thd *ThresholdDetector) calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	min := values[0]
	for _, value := range values {
		if value < min {
			min = value
		}
	}
	return min
}

// AnomalyDetector implements anomaly-based regression detection
type AnomalyDetector struct {
	config RegressionDetectionConfig
	logger *zap.Logger
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config RegressionDetectionConfig) *AnomalyDetector {
	return &AnomalyDetector{
		config: config,
		logger: zap.NewNop(),
	}
}

// Name returns the detector name
func (ad *AnomalyDetector) Name() string {
	return "anomaly_detector"
}

// Type returns the detector type
func (ad *AnomalyDetector) Type() string {
	return "anomaly"
}

// Detect performs anomaly-based regression detection
func (ad *AnomalyDetector) Detect(baseline *PerformanceBaseline, currentData []*PerformanceDataPoint) (*RegressionResult, error) {
	if len(currentData) < 10 {
		return nil, fmt.Errorf("insufficient data for anomaly detection: need at least 10 points, got %d", len(currentData))
	}

	// Extract current values
	currentValues := ad.extractMetricValues(baseline.Metric, currentData)
	if len(currentValues) == 0 {
		return nil, fmt.Errorf("no valid values found for metric %s", baseline.Metric)
	}

	// Calculate anomaly scores
	anomalyScores := ad.calculateAnomalyScores(currentValues, baseline)

	// Count anomalies
	anomalyCount := 0
	for _, score := range anomalyScores {
		if score > 2.0 { // Z-score threshold
			anomalyCount++
		}
	}

	anomalyPercent := float64(anomalyCount) / float64(len(anomalyScores)) * 100

	// Determine if anomalies indicate regression
	regressionType := "none"
	severity := "low"
	changePercent := 0.0

	if anomalyPercent > 20 { // More than 20% anomalies
		currentMean := ad.calculateMean(currentValues)
		changePercent = ((currentMean - baseline.Mean) / baseline.Mean) * 100

		if changePercent > 0 {
			regressionType = "degradation"
			severity = ad.determineSeverity(changePercent, ad.getDegradationThreshold(baseline.Metric))
		} else if changePercent < 0 {
			regressionType = "improvement"
			severity = "low"
		}
	}

	// Create result
	result := &RegressionResult{
		ID:              fmt.Sprintf("anomaly_%s_%d", baseline.Metric, time.Now().UnixNano()),
		Metric:          baseline.Metric,
		DetectedAt:      time.Now().UTC(),
		Type:            regressionType,
		BaselineMean:    baseline.Mean,
		CurrentMean:     ad.calculateMean(currentValues),
		ChangePercent:   changePercent,
		ChangeDirection: ad.getChangeDirection(ad.calculateMean(currentValues), baseline.Mean),
		PValue:          0.05, // Placeholder
		Confidence:      0.7,  // Moderate confidence for anomaly detection
		IsSignificant:   regressionType != "none",
		DetectorUsed:    ad.Name(),
		DetectionTime:   time.Now().UTC(),
		Severity:        severity,
		BaselineID:      baseline.ID,
		BaselineStart:   baseline.SampleStart,
		BaselineEnd:     baseline.SampleEnd,
		CurrentStart:    currentData[0].Timestamp,
		CurrentEnd:      currentData[len(currentData)-1].Timestamp,
		Tags:            make(map[string]string),
	}

	// Add outlier analysis
	result.OutlierAnalysis = &OutlierAnalysis{
		OutlierCount:     anomalyCount,
		OutlierPercent:   anomalyPercent,
		OutlierThreshold: 2.0,
		OutlierMethod:    "z-score",
	}

	return result, nil
}

// GetConfidence returns the detector confidence
func (ad *AnomalyDetector) GetConfidence() float64 {
	return 0.7 // Anomaly detector has moderate confidence
}

// IsApplicable checks if the detector is applicable to a metric
func (ad *AnomalyDetector) IsApplicable(metric string) bool {
	// Anomaly detector is applicable to all metrics
	return true
}

// calculateAnomalyScores calculates anomaly scores using z-score method
func (ad *AnomalyDetector) calculateAnomalyScores(values []float64, baseline *PerformanceBaseline) []float64 {
	scores := make([]float64, len(values))

	for i, value := range values {
		if baseline.StdDev > 0 {
			scores[i] = math.Abs((value - baseline.Mean) / baseline.StdDev)
		} else {
			scores[i] = 0.0
		}
	}

	return scores
}

// getChangeDirection gets the change direction
func (ad *AnomalyDetector) getChangeDirection(current, baseline float64) string {
	if math.Abs(current-baseline) < 0.001 {
		return "stable"
	} else if current > baseline {
		return "increase"
	} else {
		return "decrease"
	}
}

// getDegradationThreshold gets the degradation threshold for a metric
func (ad *AnomalyDetector) getDegradationThreshold(metric string) float64 {
	switch metric {
	case "response_time":
		return ad.config.RegressionThresholds.ResponseTime.Degradation
	case "success_rate":
		return ad.config.RegressionThresholds.SuccessRate.Degradation
	case "throughput":
		return ad.config.RegressionThresholds.Throughput.Degradation
	case "error_rate":
		return ad.config.RegressionThresholds.ErrorRate.Degradation
	case "cpu_usage":
		return ad.config.RegressionThresholds.ResourceUtilization.CPU.Degradation
	case "memory_usage":
		return ad.config.RegressionThresholds.ResourceUtilization.Memory.Degradation
	default:
		return 10.0
	}
}

// determineSeverity determines the severity based on change percentage
func (ad *AnomalyDetector) determineSeverity(changePercent, threshold float64) string {
	ratio := math.Abs(changePercent) / threshold

	if ratio >= 3.0 {
		return "critical"
	} else if ratio >= 2.0 {
		return "high"
	} else if ratio >= 1.5 {
		return "medium"
	} else {
		return "low"
	}
}

// extractMetricValues extracts metric values from data points
func (ad *AnomalyDetector) extractMetricValues(metric string, data []*PerformanceDataPoint) []float64 {
	values := make([]float64, 0, len(data))

	for _, point := range data {
		var value float64
		switch metric {
		case "response_time":
			value = float64(point.ResponseTime.Milliseconds())
		case "success_rate":
			value = point.SuccessRate
		case "throughput":
			value = point.Throughput
		case "error_rate":
			value = point.ErrorRate
		case "cpu_usage":
			value = point.CPUUsage
		case "memory_usage":
			value = point.MemoryUsage
		default:
			continue
		}
		values = append(values, value)
	}

	return values
}

// calculateMean calculates the mean of values
func (ad *AnomalyDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}
