package classification_monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AccuracyMetricsCollector collects and aggregates classification accuracy metrics
type AccuracyMetricsCollector struct {
	config         *CollectorConfig
	logger         *zap.Logger
	db             *sql.DB
	mu             sync.RWMutex
	metrics        map[string]*AggregatedMetrics
	collectors     map[string]DimensionCollector
	startTime      time.Time
	lastCollection time.Time
}

// CollectorConfig holds configuration for metrics collection
type CollectorConfig struct {
	EnableDatabaseStorage      bool            `json:"enable_database_storage"`
	EnableMemoryAggregation    bool            `json:"enable_memory_aggregation"`
	CollectionInterval         time.Duration   `json:"collection_interval"`
	RetentionPeriod            time.Duration   `json:"retention_period"`
	BatchSize                  int             `json:"batch_size"`
	EnableConcurrentCollection bool            `json:"enable_concurrent_collection"`
	MetricsResolution          time.Duration   `json:"metrics_resolution"`
	DimensionalAnalysis        []string        `json:"dimensional_analysis"`
	CalculatePercentiles       bool            `json:"calculate_percentiles"`
	EnableTrendCalculation     bool            `json:"enable_trend_calculation"`
	AggregationWindows         []time.Duration `json:"aggregation_windows"`
}

// AggregatedMetrics represents aggregated accuracy metrics
type AggregatedMetrics struct {
	DimensionName  string        `json:"dimension_name"`
	DimensionValue string        `json:"dimension_value"`
	TimeWindow     time.Duration `json:"time_window"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`

	// Basic accuracy metrics
	TotalClassifications   int     `json:"total_classifications"`
	CorrectClassifications int     `json:"correct_classifications"`
	AccuracyRate           float64 `json:"accuracy_rate"`
	ErrorRate              float64 `json:"error_rate"`

	// Confidence metrics
	AverageConfidence     float64            `json:"average_confidence"`
	ConfidenceStdDev      float64            `json:"confidence_std_dev"`
	ConfidencePercentiles map[string]float64 `json:"confidence_percentiles"`

	// Performance metrics
	AverageResponseTime     time.Duration            `json:"average_response_time"`
	ResponseTimePercentiles map[string]time.Duration `json:"response_time_percentiles"`
	ThroughputPerSecond     float64                  `json:"throughput_per_second"`

	// Distribution metrics
	ClassificationDistribution map[string]int `json:"classification_distribution"`
	MethodDistribution         map[string]int `json:"method_distribution"`
	ErrorTypeDistribution      map[string]int `json:"error_type_distribution"`

	// Trend metrics
	TrendDirection  string  `json:"trend_direction"`
	TrendStrength   float64 `json:"trend_strength"`
	TrendConfidence float64 `json:"trend_confidence"`

	// Quality metrics
	DataQualityScore float64 `json:"data_quality_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	ReliabilityScore float64 `json:"reliability_score"`

	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DimensionCollector interface for collecting metrics by dimension
type DimensionCollector interface {
	GetDimensionName() string
	CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error)
	GetSupportedAggregations() []string
}

// DimensionMetrics represents metrics for a specific dimension value
type DimensionMetrics struct {
	DimensionValue  string                 `json:"dimension_value"`
	Classifications []*ClassificationData  `json:"classifications"`
	AggregatedData  map[string]interface{} `json:"aggregated_data"`
	Timestamp       time.Time              `json:"timestamp"`
}

// ClassificationData represents individual classification data for metrics
type ClassificationData struct {
	ID               string                 `json:"id"`
	BusinessName     string                 `json:"business_name"`
	ActualIndustry   string                 `json:"actual_industry"`
	ExpectedIndustry *string                `json:"expected_industry"`
	ConfidenceScore  float64                `json:"confidence_score"`
	Method           string                 `json:"method"`
	ResponseTime     time.Duration          `json:"response_time"`
	IsCorrect        *bool                  `json:"is_correct"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// MetricsAggregationResult represents the result of metrics aggregation
type MetricsAggregationResult struct {
	AggregationPeriod  time.Duration                 `json:"aggregation_period"`
	StartTime          time.Time                     `json:"start_time"`
	EndTime            time.Time                     `json:"end_time"`
	OverallMetrics     *AggregatedMetrics            `json:"overall_metrics"`
	DimensionalMetrics map[string]*AggregatedMetrics `json:"dimensional_metrics"`
	ComparisonMetrics  *ComparisonMetrics            `json:"comparison_metrics"`
	TrendAnalysis      *TrendAnalysis                `json:"trend_analysis"`
	QualityAssessment  *QualityAssessment            `json:"quality_assessment"`
	LastUpdated        time.Time                     `json:"last_updated"`
}

// ComparisonMetrics represents comparison with previous periods
type ComparisonMetrics struct {
	PreviousPeriodAccuracy float64   `json:"previous_period_accuracy"`
	AccuracyChange         float64   `json:"accuracy_change"`
	ChangeDirection        string    `json:"change_direction"`
	ChangeSignificance     string    `json:"change_significance"`
	ChangeConfidence       float64   `json:"change_confidence"`
	Timestamp              time.Time `json:"timestamp"`
}

// TrendAnalysis represents trend analysis of metrics
type TrendAnalysis struct {
	ShortTermTrend  string    `json:"short_term_trend"`
	LongTermTrend   string    `json:"long_term_trend"`
	TrendStrength   float64   `json:"trend_strength"`
	SeasonalPattern bool      `json:"seasonal_pattern"`
	Anomalies       []string  `json:"anomalies"`
	Forecast        *Forecast `json:"forecast"`
}

// Forecast represents accuracy forecast
type Forecast struct {
	NextPeriodAccuracy float64       `json:"next_period_accuracy"`
	ConfidenceInterval [2]float64    `json:"confidence_interval"`
	ForecastHorizon    time.Duration `json:"forecast_horizon"`
	ForecastConfidence float64       `json:"forecast_confidence"`
}

// QualityAssessment represents data quality assessment
type QualityAssessment struct {
	CompletenessScore   float64  `json:"completeness_score"`
	ConsistencyScore    float64  `json:"consistency_score"`
	AccuracyScore       float64  `json:"accuracy_score"`
	TimelinessScore     float64  `json:"timeliness_score"`
	OverallQualityScore float64  `json:"overall_quality_score"`
	QualityIssues       []string `json:"quality_issues"`
	Recommendations     []string `json:"recommendations"`
}

// NewAccuracyMetricsCollector creates a new metrics collector
func NewAccuracyMetricsCollector(config *CollectorConfig, logger *zap.Logger, db *sql.DB) *AccuracyMetricsCollector {
	if config == nil {
		config = DefaultCollectorConfig()
	}

	collector := &AccuracyMetricsCollector{
		config:         config,
		logger:         logger,
		db:             db,
		metrics:        make(map[string]*AggregatedMetrics),
		collectors:     make(map[string]DimensionCollector),
		startTime:      time.Now(),
		lastCollection: time.Now(),
	}

	// Initialize built-in dimension collectors
	collector.initializeBuiltInCollectors()

	return collector
}

// DefaultCollectorConfig returns default configuration
func DefaultCollectorConfig() *CollectorConfig {
	return &CollectorConfig{
		EnableDatabaseStorage:      true,
		EnableMemoryAggregation:    true,
		CollectionInterval:         5 * time.Minute,
		RetentionPeriod:            30 * 24 * time.Hour, // 30 days
		BatchSize:                  1000,
		EnableConcurrentCollection: true,
		MetricsResolution:          1 * time.Minute,
		DimensionalAnalysis:        []string{"method", "confidence_range", "industry", "time_of_day"},
		CalculatePercentiles:       true,
		EnableTrendCalculation:     true,
		AggregationWindows:         []time.Duration{1 * time.Hour, 6 * time.Hour, 24 * time.Hour, 7 * 24 * time.Hour},
	}
}

// initializeBuiltInCollectors initializes built-in dimension collectors
func (amc *AccuracyMetricsCollector) initializeBuiltInCollectors() {
	// Method-based collector
	amc.collectors["method"] = &MethodDimensionCollector{
		name:   "method",
		db:     amc.db,
		logger: amc.logger,
	}

	// Confidence range collector
	amc.collectors["confidence_range"] = &ConfidenceRangeDimensionCollector{
		name:   "confidence_range",
		db:     amc.db,
		logger: amc.logger,
	}

	// Industry collector
	amc.collectors["industry"] = &IndustryDimensionCollector{
		name:   "industry",
		db:     amc.db,
		logger: amc.logger,
	}

	// Time-based collector
	amc.collectors["time_of_day"] = &TimeDimensionCollector{
		name:   "time_of_day",
		db:     amc.db,
		logger: amc.logger,
	}
}

// CollectAndAggregateMetrics collects and aggregates metrics for all dimensions
func (amc *AccuracyMetricsCollector) CollectAndAggregateMetrics(ctx context.Context) (*MetricsAggregationResult, error) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	amc.logger.Info("Starting metrics collection and aggregation")

	endTime := time.Now()
	startTime := amc.lastCollection

	result := &MetricsAggregationResult{
		AggregationPeriod:  endTime.Sub(startTime),
		StartTime:          startTime,
		EndTime:            endTime,
		DimensionalMetrics: make(map[string]*AggregatedMetrics),
		LastUpdated:        endTime,
	}

	// Collect overall metrics
	overallMetrics, err := amc.collectOverallMetrics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to collect overall metrics: %w", err)
	}
	result.OverallMetrics = overallMetrics

	// Collect dimensional metrics
	if amc.config.EnableConcurrentCollection {
		err = amc.collectDimensionalMetricsConcurrently(ctx, startTime, endTime, result)
	} else {
		err = amc.collectDimensionalMetricsSequentially(ctx, startTime, endTime, result)
	}
	if err != nil {
		amc.logger.Error("Failed to collect dimensional metrics", zap.Error(err))
		// Continue with partial results
	}

	// Perform comparison analysis
	comparisonMetrics, err := amc.performComparisonAnalysis(ctx, result.OverallMetrics)
	if err != nil {
		amc.logger.Error("Failed to perform comparison analysis", zap.Error(err))
	} else {
		result.ComparisonMetrics = comparisonMetrics
	}

	// Perform trend analysis
	if amc.config.EnableTrendCalculation {
		trendAnalysis, err := amc.performTrendAnalysis(ctx, result.OverallMetrics)
		if err != nil {
			amc.logger.Error("Failed to perform trend analysis", zap.Error(err))
		} else {
			result.TrendAnalysis = trendAnalysis
		}
	}

	// Perform quality assessment
	qualityAssessment, err := amc.performQualityAssessment(ctx, result)
	if err != nil {
		amc.logger.Error("Failed to perform quality assessment", zap.Error(err))
	} else {
		result.QualityAssessment = qualityAssessment
	}

	// Store results if database storage is enabled
	if amc.config.EnableDatabaseStorage {
		if err := amc.storeMetricsInDatabase(ctx, result); err != nil {
			amc.logger.Error("Failed to store metrics in database", zap.Error(err))
		}
	}

	// Update in-memory cache if enabled
	if amc.config.EnableMemoryAggregation {
		amc.updateMemoryCache(result)
	}

	amc.lastCollection = endTime

	amc.logger.Info("Metrics collection and aggregation completed",
		zap.Duration("duration", endTime.Sub(startTime)),
		zap.Int("dimensional_metrics", len(result.DimensionalMetrics)),
		zap.Float64("overall_accuracy", result.OverallMetrics.AccuracyRate))

	return result, nil
}

// collectOverallMetrics collects overall accuracy metrics
func (amc *AccuracyMetricsCollector) collectOverallMetrics(ctx context.Context, startTime, endTime time.Time) (*AggregatedMetrics, error) {
	// Query classification data from database
	query := `
		SELECT 
			id, business_name, actual_classification, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := amc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications: %w", err)
	}
	defer rows.Close()

	var classifications []*ClassificationData
	for rows.Next() {
		var c ClassificationData
		var processingTimeMs sql.NullInt64
		var expectedClassification sql.NullString
		var isCorrect sql.NullBool
		var metadataJSON sql.NullString

		err := rows.Scan(
			&c.ID, &c.BusinessName, &c.ActualIndustry, &expectedClassification,
			&c.ConfidenceScore, &c.Method, &processingTimeMs,
			&isCorrect, &c.Timestamp, &metadataJSON,
		)
		if err != nil {
			amc.logger.Error("Failed to scan classification row", zap.Error(err))
			continue
		}

		if expectedClassification.Valid {
			c.ExpectedIndustry = &expectedClassification.String
		}

		if isCorrect.Valid {
			correct := isCorrect.Bool
			c.IsCorrect = &correct
		}

		if processingTimeMs.Valid {
			c.ResponseTime = time.Duration(processingTimeMs.Int64) * time.Millisecond
		}

		classifications = append(classifications, &c)
	}

	// Aggregate the metrics
	metrics := amc.aggregateClassificationData(classifications, "overall", "all")
	metrics.TimeWindow = endTime.Sub(startTime)
	metrics.StartTime = startTime
	metrics.EndTime = endTime

	return metrics, nil
}

// aggregateClassificationData aggregates classification data into metrics
func (amc *AccuracyMetricsCollector) aggregateClassificationData(classifications []*ClassificationData, dimensionName, dimensionValue string) *AggregatedMetrics {
	metrics := &AggregatedMetrics{
		DimensionName:              dimensionName,
		DimensionValue:             dimensionValue,
		ClassificationDistribution: make(map[string]int),
		MethodDistribution:         make(map[string]int),
		ErrorTypeDistribution:      make(map[string]int),
		ConfidencePercentiles:      make(map[string]float64),
		ResponseTimePercentiles:    make(map[string]time.Duration),
		Metadata:                   make(map[string]interface{}),
		LastUpdated:                time.Now(),
	}

	if len(classifications) == 0 {
		return metrics
	}

	// Basic counts and accumulations
	var totalResponseTime time.Duration
	var confidenceSum float64
	var confidenceValues []float64
	var responseTimeValues []time.Duration
	var correctCount int

	for _, c := range classifications {
		metrics.TotalClassifications++

		// Count correct classifications
		if c.IsCorrect != nil && *c.IsCorrect {
			correctCount++
		}

		// Accumulate confidence and response time
		confidenceSum += c.ConfidenceScore
		confidenceValues = append(confidenceValues, c.ConfidenceScore)

		if c.ResponseTime > 0 {
			totalResponseTime += c.ResponseTime
			responseTimeValues = append(responseTimeValues, c.ResponseTime)
		}

		// Distribution tracking
		metrics.ClassificationDistribution[c.ActualIndustry]++
		metrics.MethodDistribution[c.Method]++

		// Error type distribution (for incorrect classifications)
		if c.IsCorrect != nil && !*c.IsCorrect {
			errorType := amc.classifyErrorType(c.ConfidenceScore)
			metrics.ErrorTypeDistribution[errorType]++
		}
	}

	metrics.CorrectClassifications = correctCount

	// Calculate rates
	if metrics.TotalClassifications > 0 {
		metrics.AccuracyRate = float64(correctCount) / float64(metrics.TotalClassifications)
		metrics.ErrorRate = 1.0 - metrics.AccuracyRate
		metrics.AverageConfidence = confidenceSum / float64(metrics.TotalClassifications)

		// Calculate confidence standard deviation
		if len(confidenceValues) > 1 {
			varianceSum := 0.0
			for _, conf := range confidenceValues {
				diff := conf - metrics.AverageConfidence
				varianceSum += diff * diff
			}
			variance := varianceSum / float64(len(confidenceValues)-1)
			metrics.ConfidenceStdDev = math.Sqrt(variance)
		}

		// Calculate average response time
		if len(responseTimeValues) > 0 {
			metrics.AverageResponseTime = totalResponseTime / time.Duration(len(responseTimeValues))
		}

		// Calculate throughput
		if len(classifications) > 0 {
			timeSpan := classifications[0].Timestamp.Sub(classifications[len(classifications)-1].Timestamp)
			if timeSpan > 0 {
				metrics.ThroughputPerSecond = float64(len(classifications)) / timeSpan.Seconds()
			}
		}
	}

	// Calculate percentiles if enabled
	if amc.config.CalculatePercentiles {
		metrics.ConfidencePercentiles = amc.calculatePercentiles(confidenceValues)
		metrics.ResponseTimePercentiles = amc.calculateDurationPercentiles(responseTimeValues)
	}

	// Calculate quality scores
	metrics.DataQualityScore = amc.calculateDataQualityScore(classifications)
	metrics.ConsistencyScore = amc.calculateConsistencyScore(classifications)
	metrics.ReliabilityScore = amc.calculateReliabilityScore(metrics)

	return metrics
}

// calculatePercentiles calculates percentiles for float64 values
func (amc *AccuracyMetricsCollector) calculatePercentiles(values []float64) map[string]float64 {
	if len(values) == 0 {
		return make(map[string]float64)
	}

	sort.Float64s(values)

	percentiles := map[string]float64{
		"p25": amc.calculatePercentile(values, 0.25),
		"p50": amc.calculatePercentile(values, 0.50),
		"p75": amc.calculatePercentile(values, 0.75),
		"p90": amc.calculatePercentile(values, 0.90),
		"p95": amc.calculatePercentile(values, 0.95),
		"p99": amc.calculatePercentile(values, 0.99),
	}

	return percentiles
}

// calculateDurationPercentiles calculates percentiles for duration values
func (amc *AccuracyMetricsCollector) calculateDurationPercentiles(values []time.Duration) map[string]time.Duration {
	if len(values) == 0 {
		return make(map[string]time.Duration)
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	percentiles := map[string]time.Duration{
		"p25": amc.calculateDurationPercentile(values, 0.25),
		"p50": amc.calculateDurationPercentile(values, 0.50),
		"p75": amc.calculateDurationPercentile(values, 0.75),
		"p90": amc.calculateDurationPercentile(values, 0.90),
		"p95": amc.calculateDurationPercentile(values, 0.95),
		"p99": amc.calculateDurationPercentile(values, 0.99),
	}

	return percentiles
}

// calculatePercentile calculates a specific percentile
func (amc *AccuracyMetricsCollector) calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := percentile * float64(len(sortedValues)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sortedValues) {
		return sortedValues[len(sortedValues)-1]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// calculateDurationPercentile calculates a specific percentile for durations
func (amc *AccuracyMetricsCollector) calculateDurationPercentile(sortedValues []time.Duration, percentile float64) time.Duration {
	if len(sortedValues) == 0 {
		return 0
	}

	index := percentile * float64(len(sortedValues)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sortedValues) {
		return sortedValues[len(sortedValues)-1]
	}

	weight := index - float64(lower)
	lowerDuration := float64(sortedValues[lower])
	upperDuration := float64(sortedValues[upper])

	result := lowerDuration*(1-weight) + upperDuration*weight
	return time.Duration(result)
}

// classifyErrorType classifies the type of error based on confidence
func (amc *AccuracyMetricsCollector) classifyErrorType(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "very_high_confidence_error"
	case confidence >= 0.7:
		return "high_confidence_error"
	case confidence >= 0.5:
		return "medium_confidence_error"
	case confidence >= 0.3:
		return "low_confidence_error"
	default:
		return "very_low_confidence_error"
	}
}

// collectDimensionalMetricsConcurrently collects dimensional metrics concurrently
func (amc *AccuracyMetricsCollector) collectDimensionalMetricsConcurrently(ctx context.Context, startTime, endTime time.Time, result *MetricsAggregationResult) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	for dimensionName, collector := range amc.collectors {
		wg.Add(1)
		go func(name string, col DimensionCollector) {
			defer wg.Done()

			dimensionMetrics, err := col.CollectMetrics(ctx, startTime, endTime)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to collect metrics for dimension %s: %w", name, err))
				mu.Unlock()
				return
			}

			// Aggregate metrics for each dimension value
			for _, dimMetrics := range dimensionMetrics {
				aggregated := amc.aggregateClassificationData(dimMetrics.Classifications, name, dimMetrics.DimensionValue)
				aggregated.TimeWindow = endTime.Sub(startTime)
				aggregated.StartTime = startTime
				aggregated.EndTime = endTime

				mu.Lock()
				result.DimensionalMetrics[fmt.Sprintf("%s:%s", name, dimMetrics.DimensionValue)] = aggregated
				mu.Unlock()
			}
		}(dimensionName, collector)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred during concurrent collection: %v", errors)
	}

	return nil
}

// collectDimensionalMetricsSequentially collects dimensional metrics sequentially
func (amc *AccuracyMetricsCollector) collectDimensionalMetricsSequentially(ctx context.Context, startTime, endTime time.Time, result *MetricsAggregationResult) error {
	for dimensionName, collector := range amc.collectors {
		dimensionMetrics, err := collector.CollectMetrics(ctx, startTime, endTime)
		if err != nil {
			amc.logger.Error("Failed to collect metrics for dimension",
				zap.String("dimension", dimensionName),
				zap.Error(err))
			continue
		}

		// Aggregate metrics for each dimension value
		for _, dimMetrics := range dimensionMetrics {
			aggregated := amc.aggregateClassificationData(dimMetrics.Classifications, dimensionName, dimMetrics.DimensionValue)
			aggregated.TimeWindow = endTime.Sub(startTime)
			aggregated.StartTime = startTime
			aggregated.EndTime = endTime

			result.DimensionalMetrics[fmt.Sprintf("%s:%s", dimensionName, dimMetrics.DimensionValue)] = aggregated
		}
	}

	return nil
}

// Quality calculation methods

func (amc *AccuracyMetricsCollector) calculateDataQualityScore(classifications []*ClassificationData) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	score := 0.0
	factors := 0

	// Completeness: presence of expected fields
	completeRecords := 0
	for _, c := range classifications {
		if c.BusinessName != "" && c.ActualIndustry != "" && c.ConfidenceScore > 0 {
			completeRecords++
		}
	}
	completeness := float64(completeRecords) / float64(len(classifications))
	score += completeness
	factors++

	// Validity: reasonable confidence scores
	validConfidenceRecords := 0
	for _, c := range classifications {
		if c.ConfidenceScore >= 0.0 && c.ConfidenceScore <= 1.0 {
			validConfidenceRecords++
		}
	}
	validity := float64(validConfidenceRecords) / float64(len(classifications))
	score += validity
	factors++

	// Timeliness: recent data
	now := time.Now()
	recentRecords := 0
	for _, c := range classifications {
		if now.Sub(c.Timestamp) < 24*time.Hour {
			recentRecords++
		}
	}
	timeliness := float64(recentRecords) / float64(len(classifications))
	score += timeliness
	factors++

	return score / float64(factors)
}

func (amc *AccuracyMetricsCollector) calculateConsistencyScore(classifications []*ClassificationData) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	// Check for consistency in similar business names
	businessGroups := make(map[string][]*ClassificationData)
	for _, c := range classifications {
		// Group by similar business names (simplified)
		normalizedName := amc.normalizeName(c.BusinessName)
		businessGroups[normalizedName] = append(businessGroups[normalizedName], c)
	}

	consistentGroups := 0
	totalGroups := 0

	for _, group := range businessGroups {
		if len(group) < 2 {
			continue
		}

		totalGroups++

		// Check if all classifications in the group are consistent
		firstIndustry := group[0].ActualIndustry
		consistent := true
		for _, c := range group[1:] {
			if c.ActualIndustry != firstIndustry {
				consistent = false
				break
			}
		}

		if consistent {
			consistentGroups++
		}
	}

	if totalGroups == 0 {
		return 1.0 // No groups to compare, assume consistent
	}

	return float64(consistentGroups) / float64(totalGroups)
}

func (amc *AccuracyMetricsCollector) calculateReliabilityScore(metrics *AggregatedMetrics) float64 {
	score := 0.0
	factors := 0

	// Factor 1: High accuracy rate
	score += metrics.AccuracyRate
	factors++

	// Factor 2: Consistent confidence (low std dev relative to mean)
	if metrics.AverageConfidence > 0 {
		confidenceVariability := metrics.ConfidenceStdDev / metrics.AverageConfidence
		confidenceReliability := math.Max(0, 1.0-confidenceVariability)
		score += confidenceReliability
		factors++
	}

	// Factor 3: Sufficient sample size
	sampleSizeScore := math.Min(1.0, float64(metrics.TotalClassifications)/100.0)
	score += sampleSizeScore
	factors++

	if factors == 0 {
		return 0.0
	}

	return score / float64(factors)
}

func (amc *AccuracyMetricsCollector) normalizeName(name string) string {
	// Simple normalization - remove common suffixes and convert to lowercase
	normalized := strings.ToLower(name)
	suffixes := []string{" inc", " llc", " corp", " ltd", " company", " co"}

	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			break
		}
	}

	return strings.TrimSpace(normalized)
}

// performComparisonAnalysis performs comparison with previous periods
func (amc *AccuracyMetricsCollector) performComparisonAnalysis(ctx context.Context, currentMetrics *AggregatedMetrics) (*ComparisonMetrics, error) {
	// Get previous period metrics from database
	endTime := currentMetrics.StartTime
	startTime := endTime.Add(-currentMetrics.TimeWindow)

	previousMetrics, err := amc.collectOverallMetrics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to collect previous period metrics: %w", err)
	}

	comparison := &ComparisonMetrics{
		PreviousPeriodAccuracy: previousMetrics.AccuracyRate,
		AccuracyChange:         currentMetrics.AccuracyRate - previousMetrics.AccuracyRate,
		Timestamp:              time.Now(),
	}

	// Determine change direction
	if comparison.AccuracyChange > 0.01 {
		comparison.ChangeDirection = "improving"
	} else if comparison.AccuracyChange < -0.01 {
		comparison.ChangeDirection = "declining"
	} else {
		comparison.ChangeDirection = "stable"
	}

	// Calculate change significance
	if math.Abs(comparison.AccuracyChange) > 0.05 {
		comparison.ChangeSignificance = "significant"
	} else if math.Abs(comparison.AccuracyChange) > 0.02 {
		comparison.ChangeSignificance = "moderate"
	} else {
		comparison.ChangeSignificance = "minimal"
	}

	// Simple confidence calculation (could be enhanced with statistical tests)
	sampleSizeWeight := math.Min(1.0, float64(currentMetrics.TotalClassifications)/100.0)
	comparison.ChangeConfidence = sampleSizeWeight * 0.8 // Max 80% confidence

	return comparison, nil
}

// performTrendAnalysis performs trend analysis
func (amc *AccuracyMetricsCollector) performTrendAnalysis(ctx context.Context, currentMetrics *AggregatedMetrics) (*TrendAnalysis, error) {
	// This is a simplified implementation
	// In a real system, you would collect historical data points and perform more sophisticated trend analysis

	analysis := &TrendAnalysis{
		ShortTermTrend:  "stable",
		LongTermTrend:   "stable",
		TrendStrength:   0.5,
		SeasonalPattern: false,
		Anomalies:       make([]string, 0),
	}

	// Simple forecast based on current metrics
	analysis.Forecast = &Forecast{
		NextPeriodAccuracy: currentMetrics.AccuracyRate,
		ConfidenceInterval: [2]float64{currentMetrics.AccuracyRate - 0.05, currentMetrics.AccuracyRate + 0.05},
		ForecastHorizon:    currentMetrics.TimeWindow,
		ForecastConfidence: 0.7,
	}

	return analysis, nil
}

// performQualityAssessment performs comprehensive quality assessment
func (amc *AccuracyMetricsCollector) performQualityAssessment(ctx context.Context, result *MetricsAggregationResult) (*QualityAssessment, error) {
	assessment := &QualityAssessment{
		QualityIssues:   make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Use the quality scores from overall metrics
	if result.OverallMetrics != nil {
		assessment.CompletenessScore = result.OverallMetrics.DataQualityScore
		assessment.ConsistencyScore = result.OverallMetrics.ConsistencyScore
		assessment.AccuracyScore = result.OverallMetrics.AccuracyRate
		assessment.TimelinessScore = 1.0 // Simplified

		// Calculate overall quality score
		assessment.OverallQualityScore = (assessment.CompletenessScore +
			assessment.ConsistencyScore +
			assessment.AccuracyScore +
			assessment.TimelinessScore) / 4.0
	}

	// Identify quality issues
	if assessment.AccuracyScore < 0.85 {
		assessment.QualityIssues = append(assessment.QualityIssues, "Low accuracy rate")
		assessment.Recommendations = append(assessment.Recommendations, "Improve classification algorithms")
	}

	if assessment.ConsistencyScore < 0.8 {
		assessment.QualityIssues = append(assessment.QualityIssues, "Inconsistent classifications")
		assessment.Recommendations = append(assessment.Recommendations, "Review classification rules and training data")
	}

	if assessment.CompletenessScore < 0.9 {
		assessment.QualityIssues = append(assessment.QualityIssues, "Incomplete data records")
		assessment.Recommendations = append(assessment.Recommendations, "Improve data validation and collection processes")
	}

	return assessment, nil
}

// storeMetricsInDatabase stores aggregated metrics in the database
func (amc *AccuracyMetricsCollector) storeMetricsInDatabase(ctx context.Context, result *MetricsAggregationResult) error {
	// This would implement database storage of metrics
	// For now, just log that we would store them
	amc.logger.Info("Storing metrics in database",
		zap.Time("start_time", result.StartTime),
		zap.Time("end_time", result.EndTime),
		zap.Float64("overall_accuracy", result.OverallMetrics.AccuracyRate),
		zap.Int("dimensional_metrics_count", len(result.DimensionalMetrics)))

	return nil
}

// updateMemoryCache updates the in-memory metrics cache
func (amc *AccuracyMetricsCollector) updateMemoryCache(result *MetricsAggregationResult) {
	// Store overall metrics
	amc.metrics["overall"] = result.OverallMetrics

	// Store dimensional metrics
	for key, metrics := range result.DimensionalMetrics {
		amc.metrics[key] = metrics
	}

	// Clean up old entries beyond retention period
	cutoff := time.Now().Add(-amc.config.RetentionPeriod)
	for key, metrics := range amc.metrics {
		if metrics.EndTime.Before(cutoff) {
			delete(amc.metrics, key)
		}
	}
}

// GetMetrics returns cached metrics
func (amc *AccuracyMetricsCollector) GetMetrics(dimensionKey string) *AggregatedMetrics {
	amc.mu.RLock()
	defer amc.mu.RUnlock()

	return amc.metrics[dimensionKey]
}

// GetAllMetrics returns all cached metrics
func (amc *AccuracyMetricsCollector) GetAllMetrics() map[string]*AggregatedMetrics {
	amc.mu.RLock()
	defer amc.mu.RUnlock()

	result := make(map[string]*AggregatedMetrics)
	for k, v := range amc.metrics {
		result[k] = v
	}
	return result
}

// RegisterDimensionCollector registers a custom dimension collector
func (amc *AccuracyMetricsCollector) RegisterDimensionCollector(collector DimensionCollector) {
	amc.mu.Lock()
	defer amc.mu.Unlock()

	amc.collectors[collector.GetDimensionName()] = collector
}

// StartPeriodicCollection starts periodic metrics collection
func (amc *AccuracyMetricsCollector) StartPeriodicCollection(ctx context.Context) {
	ticker := time.NewTicker(amc.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := amc.CollectAndAggregateMetrics(ctx); err != nil {
				amc.logger.Error("Periodic metrics collection failed", zap.Error(err))
			}
		}
	}
}
