package multi_site_aggregation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// =============================================================================
// Cross-Site Correlation Models
// =============================================================================

// valueWithTime represents a numeric value with timestamp and site ID
type valueWithTime struct {
	value     float64
	timestamp time.Time
	siteID    string
}

// CorrelationAnalysis represents the results of cross-site data correlation analysis
type CorrelationAnalysis struct {
	ID                string                        `json:"id"`
	BusinessID        string                        `json:"business_id"`
	AnalysisDate      time.Time                     `json:"analysis_date"`
	CorrelationMatrix map[string]map[string]float64 `json:"correlation_matrix"`
	DataPatterns      []DataPattern                 `json:"data_patterns"`
	Anomalies         []DataAnomaly                 `json:"anomalies"`
	Trends            []DataTrend                   `json:"trends"`
	Insights          []DataInsight                 `json:"insights"`
	ConfidenceScore   float64                       `json:"confidence_score"`
	AnalysisMethod    string                        `json:"analysis_method"`
	ProcessingTime    time.Duration                 `json:"processing_time"`
	Metadata          map[string]interface{}        `json:"metadata,omitempty"`
}

// DataPattern represents a pattern found across multiple sites
type DataPattern struct {
	ID             string                 `json:"id"`
	PatternType    string                 `json:"pattern_type"` // "consistency", "variation", "trend", "seasonal"
	FieldName      string                 `json:"field_name"`
	PatternValue   interface{}            `json:"pattern_value"`
	Confidence     float64                `json:"confidence"`
	AffectedSites  []string               `json:"affected_sites"`
	Description    string                 `json:"description"`
	Significance   string                 `json:"significance"` // "high", "medium", "low"
	Recommendation string                 `json:"recommendation,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DataAnomaly represents an anomaly detected across sites
type DataAnomaly struct {
	ID             string                 `json:"id"`
	AnomalyType    string                 `json:"anomaly_type"` // "outlier", "missing", "inconsistent", "duplicate"
	FieldName      string                 `json:"field_name"`
	ExpectedValue  interface{}            `json:"expected_value"`
	ActualValue    interface{}            `json:"actual_value"`
	AffectedSite   string                 `json:"affected_site"`
	Severity       string                 `json:"severity"` // "low", "medium", "high", "critical"
	Description    string                 `json:"description"`
	RootCause      string                 `json:"root_cause,omitempty"`
	Recommendation string                 `json:"recommendation,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DataTrend represents a trend identified across sites
type DataTrend struct {
	ID            string                 `json:"id"`
	TrendType     string                 `json:"trend_type"` // "increasing", "decreasing", "stable", "fluctuating"
	FieldName     string                 `json:"field_name"`
	Direction     string                 `json:"direction"`
	Magnitude     float64                `json:"magnitude"`
	Confidence    float64                `json:"confidence"`
	Timeframe     time.Duration          `json:"timeframe"`
	AffectedSites []string               `json:"affected_sites"`
	Description   string                 `json:"description"`
	Prediction    interface{}            `json:"prediction,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DataInsight represents an insight derived from correlation analysis
type DataInsight struct {
	ID             string                 `json:"id"`
	InsightType    string                 `json:"insight_type"` // "business", "operational", "quality", "trend"
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Confidence     float64                `json:"confidence"`
	Impact         string                 `json:"impact"` // "high", "medium", "low"
	Recommendation string                 `json:"recommendation"`
	RelatedFields  []string               `json:"related_fields"`
	AffectedSites  []string               `json:"affected_sites"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CorrelationConfig holds configuration for correlation analysis
type CorrelationConfig struct {
	MinCorrelationThreshold float64       `json:"min_correlation_threshold"`
	MaxAnalysisTime         time.Duration `json:"max_analysis_time"`
	EnablePatternDetection  bool          `json:"enable_pattern_detection"`
	EnableAnomalyDetection  bool          `json:"enable_anomaly_detection"`
	EnableTrendAnalysis     bool          `json:"enable_trend_analysis"`
	EnableInsightGeneration bool          `json:"enable_insight_generation"`
	ConfidenceThreshold     float64       `json:"confidence_threshold"`
	MaxPatternsPerField     int           `json:"max_patterns_per_field"`
	MaxAnomaliesPerField    int           `json:"max_anomalies_per_field"`
	MaxTrendsPerField       int           `json:"max_trends_per_field"`
	MaxInsightsPerAnalysis  int           `json:"max_insights_per_analysis"`
}

// DefaultCorrelationConfig returns default configuration for correlation analysis
func DefaultCorrelationConfig() *CorrelationConfig {
	return &CorrelationConfig{
		MinCorrelationThreshold: 0.3,
		MaxAnalysisTime:         30 * time.Second,
		EnablePatternDetection:  true,
		EnableAnomalyDetection:  true,
		EnableTrendAnalysis:     true,
		EnableInsightGeneration: true,
		ConfidenceThreshold:     0.7,
		MaxPatternsPerField:     5,
		MaxAnomaliesPerField:    10,
		MaxTrendsPerField:       3,
		MaxInsightsPerAnalysis:  10,
	}
}

// =============================================================================
// Cross-Site Correlation Service
// =============================================================================

// CrossSiteCorrelationService provides cross-site data correlation and analysis
type CrossSiteCorrelationService struct {
	config *CorrelationConfig
	logger *zap.Logger
}

// NewCrossSiteCorrelationService creates a new cross-site correlation service
func NewCrossSiteCorrelationService(config *CorrelationConfig, logger *zap.Logger) *CrossSiteCorrelationService {
	if config == nil {
		config = DefaultCorrelationConfig()
	}

	return &CrossSiteCorrelationService{
		config: config,
		logger: logger,
	}
}

// AnalyzeCorrelations performs comprehensive cross-site correlation analysis
func (s *CrossSiteCorrelationService) AnalyzeCorrelations(
	ctx context.Context,
	businessID string,
	sitesData []SiteData,
) (*CorrelationAnalysis, error) {
	startTime := time.Now()

	s.logger.Info("Starting cross-site correlation analysis",
		zap.String("business_id", businessID),
		zap.Int("sites_count", len(sitesData)))

	// Set timeout context
	ctx, cancel := context.WithTimeout(ctx, s.config.MaxAnalysisTime)
	defer cancel()

	if len(sitesData) < 2 {
		return nil, fmt.Errorf("insufficient data for correlation analysis: need at least 2 sites, got %d", len(sitesData))
	}

	// Extract all unique fields across all sites
	fields := s.extractAllFields(sitesData)

	// Calculate correlation matrix
	correlationMatrix := s.calculateCorrelationMatrix(sitesData, fields)

	// Detect patterns
	var patterns []DataPattern
	if s.config.EnablePatternDetection {
		patterns = s.detectPatterns(sitesData, fields)
	}

	// Detect anomalies
	var anomalies []DataAnomaly
	if s.config.EnableAnomalyDetection {
		anomalies = s.detectAnomalies(sitesData, fields)
	}

	// Analyze trends
	var trends []DataTrend
	if s.config.EnableTrendAnalysis {
		trends = s.analyzeTrends(sitesData, fields)
	}

	// Generate insights
	var insights []DataInsight
	if s.config.EnableInsightGeneration {
		insights = s.generateInsights(sitesData, correlationMatrix, patterns, anomalies, trends)
	}

	// Calculate overall confidence score
	confidenceScore := s.calculateConfidenceScore(correlationMatrix, patterns, anomalies, trends, insights)

	analysis := &CorrelationAnalysis{
		ID:                generateID(),
		BusinessID:        businessID,
		AnalysisDate:      time.Now(),
		CorrelationMatrix: correlationMatrix,
		DataPatterns:      patterns,
		Anomalies:         anomalies,
		Trends:            trends,
		Insights:          insights,
		ConfidenceScore:   confidenceScore,
		AnalysisMethod:    "comprehensive_correlation",
		ProcessingTime:    time.Since(startTime),
		Metadata: map[string]interface{}{
			"fields_analyzed":    len(fields),
			"patterns_detected":  len(patterns),
			"anomalies_detected": len(anomalies),
			"trends_identified":  len(trends),
			"insights_generated": len(insights),
			"correlation_pairs":  len(correlationMatrix),
		},
	}

	s.logger.Info("Cross-site correlation analysis completed",
		zap.String("business_id", businessID),
		zap.Float64("confidence_score", confidenceScore),
		zap.Duration("processing_time", analysis.ProcessingTime),
		zap.Int("patterns", len(patterns)),
		zap.Int("anomalies", len(anomalies)),
		zap.Int("trends", len(trends)),
		zap.Int("insights", len(insights)))

	return analysis, nil
}

// =============================================================================
// Correlation Analysis Methods
// =============================================================================

// extractAllFields extracts all unique fields from site data
func (s *CrossSiteCorrelationService) extractAllFields(sitesData []SiteData) []string {
	fieldSet := make(map[string]bool)

	for _, siteData := range sitesData {
		for field := range siteData.ExtractedData {
			fieldSet[field] = true
		}
	}

	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}

	sort.Strings(fields)
	return fields
}

// calculateCorrelationMatrix calculates correlation coefficients between fields
func (s *CrossSiteCorrelationService) calculateCorrelationMatrix(sitesData []SiteData, fields []string) map[string]map[string]float64 {
	matrix := make(map[string]map[string]float64)

	for i, field1 := range fields {
		matrix[field1] = make(map[string]float64)
		for j, field2 := range fields {
			if i == j {
				matrix[field1][field2] = 1.0 // Perfect correlation with self
			} else {
				correlation := s.calculateFieldCorrelation(sitesData, field1, field2)
				// Always include the correlation value, even if it's below threshold
				matrix[field1][field2] = correlation
			}
		}
	}

	return matrix
}

// calculateFieldCorrelation calculates correlation between two fields
func (s *CrossSiteCorrelationService) calculateFieldCorrelation(sitesData []SiteData, field1, field2 string) float64 {
	var values1, values2 []float64

	// Extract numeric values for both fields
	for _, siteData := range sitesData {
		val1 := s.extractNumericValue(siteData.ExtractedData[field1])
		val2 := s.extractNumericValue(siteData.ExtractedData[field2])

		if val1 != nil && val2 != nil {
			values1 = append(values1, *val1)
			values2 = append(values2, *val2)
		}
	}

	if len(values1) < 2 {
		return 0.0 // Insufficient data for correlation
	}

	// Calculate Pearson correlation coefficient
	return s.calculatePearsonCorrelation(values1, values2)
}

// extractNumericValue extracts numeric value from interface{}
func (s *CrossSiteCorrelationService) extractNumericValue(value interface{}) *float64 {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case float64:
		return &v
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case string:
		// Try to parse as number
		if f, err := parseNumericString(v); err == nil {
			return &f
		}
		// Try to extract numeric patterns
		if f, err := extractNumericFromString(v); err == nil {
			return &f
		}
		return nil
	default:
		return nil
	}
}

// calculatePearsonCorrelation calculates Pearson correlation coefficient
func (s *CrossSiteCorrelationService) calculatePearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	n := float64(len(x))

	// Calculate means
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate correlation
	var numerator, sumXSquared, sumYSquared float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXSquared += dx * dx
		sumYSquared += dy * dy
	}

	denominator := math.Sqrt(sumXSquared * sumYSquared)
	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

// detectPatterns detects patterns in the data across sites
func (s *CrossSiteCorrelationService) detectPatterns(sitesData []SiteData, fields []string) []DataPattern {
	var patterns []DataPattern

	for _, field := range fields {
		fieldPatterns := s.detectFieldPatterns(sitesData, field)
		patterns = append(patterns, fieldPatterns...)

		// Limit patterns per field
		if len(fieldPatterns) > s.config.MaxPatternsPerField {
			patterns = patterns[:s.config.MaxPatternsPerField]
		}
	}

	return patterns
}

// detectFieldPatterns detects patterns for a specific field
func (s *CrossSiteCorrelationService) detectFieldPatterns(sitesData []SiteData, field string) []DataPattern {
	var patterns []DataPattern

	// Extract values for this field
	values := make(map[string]interface{})
	for _, siteData := range sitesData {
		if value, exists := siteData.ExtractedData[field]; exists {
			values[siteData.LocationID] = value
		}
	}

	if len(values) < 2 {
		return patterns
	}

	// Detect consistency pattern
	if consistencyPattern := s.detectConsistencyPattern(field, values); consistencyPattern != nil {
		patterns = append(patterns, *consistencyPattern)
	}

	// Detect variation pattern
	if variationPattern := s.detectVariationPattern(field, values); variationPattern != nil {
		patterns = append(patterns, *variationPattern)
	}

	// Detect seasonal pattern (if temporal data available)
	if seasonalPattern := s.detectSeasonalPattern(field, values); seasonalPattern != nil {
		patterns = append(patterns, *seasonalPattern)
	}

	return patterns
}

// detectConsistencyPattern detects if values are consistent across sites
func (s *CrossSiteCorrelationService) detectConsistencyPattern(field string, values map[string]interface{}) *DataPattern {
	uniqueValues := make(map[string]int)
	for _, value := range values {
		valueStr := fmt.Sprintf("%v", value)
		uniqueValues[valueStr]++
	}

	if len(uniqueValues) == 1 {
		// All values are the same
		var valueStr string
		for v := range uniqueValues {
			valueStr = v
			break
		}

		affectedSites := make([]string, 0, len(values))
		for siteID := range values {
			affectedSites = append(affectedSites, siteID)
		}

		return &DataPattern{
			ID:             generateID(),
			PatternType:    "consistency",
			FieldName:      field,
			PatternValue:   valueStr,
			Confidence:     0.95,
			AffectedSites:  affectedSites,
			Description:    fmt.Sprintf("All sites have consistent value '%s' for field '%s'", valueStr, field),
			Significance:   "high",
			Recommendation: "This consistency indicates reliable data across all sites",
		}
	}

	return nil
}

// detectVariationPattern detects variation patterns in values
func (s *CrossSiteCorrelationService) detectVariationPattern(field string, values map[string]interface{}) *DataPattern {
	uniqueValues := make(map[string]int)
	for _, value := range values {
		valueStr := fmt.Sprintf("%v", value)
		uniqueValues[valueStr]++
	}

	if len(uniqueValues) > 1 {
		// Values vary across sites
		affectedSites := make([]string, 0, len(values))
		for siteID := range values {
			affectedSites = append(affectedSites, siteID)
		}

		variationLevel := "low"
		if len(uniqueValues) > len(values)/2 {
			variationLevel = "high"
		} else if len(uniqueValues) > len(values)/4 {
			variationLevel = "medium"
		}

		return &DataPattern{
			ID:             generateID(),
			PatternType:    "variation",
			FieldName:      field,
			PatternValue:   fmt.Sprintf("%d unique values", len(uniqueValues)),
			Confidence:     0.85,
			AffectedSites:  affectedSites,
			Description:    fmt.Sprintf("Field '%s' shows %s variation across sites (%d unique values)", field, variationLevel, len(uniqueValues)),
			Significance:   variationLevel,
			Recommendation: "Investigate reasons for variation and consider standardization",
		}
	}

	return nil
}

// detectSeasonalPattern detects seasonal patterns (placeholder implementation)
func (s *CrossSiteCorrelationService) detectSeasonalPattern(field string, values map[string]interface{}) *DataPattern {
	// This would require temporal data analysis
	// For now, return nil as this is a placeholder
	return nil
}

// detectAnomalies detects anomalies in the data
func (s *CrossSiteCorrelationService) detectAnomalies(sitesData []SiteData, fields []string) []DataAnomaly {
	var anomalies []DataAnomaly

	for _, field := range fields {
		fieldAnomalies := s.detectFieldAnomalies(sitesData, field)
		anomalies = append(anomalies, fieldAnomalies...)

		// Limit anomalies per field
		if len(fieldAnomalies) > s.config.MaxAnomaliesPerField {
			anomalies = anomalies[:s.config.MaxAnomaliesPerField]
		}
	}

	return anomalies
}

// detectFieldAnomalies detects anomalies for a specific field
func (s *CrossSiteCorrelationService) detectFieldAnomalies(sitesData []SiteData, field string) []DataAnomaly {
	var anomalies []DataAnomaly

	// Extract numeric values for statistical analysis
	var numericValues []float64
	valueMap := make(map[string]float64)

	for _, siteData := range sitesData {
		if value, exists := siteData.ExtractedData[field]; exists {
			if numericValue := s.extractNumericValue(value); numericValue != nil {
				numericValues = append(numericValues, *numericValue)
				valueMap[siteData.LocationID] = *numericValue
			}
		}
	}

	// Detect missing values first (this works regardless of numeric values)
	for _, siteData := range sitesData {
		if _, exists := siteData.ExtractedData[field]; !exists {
			anomaly := DataAnomaly{
				ID:             generateID(),
				AnomalyType:    "missing",
				FieldName:      field,
				ExpectedValue:  "Present in other sites",
				ActualValue:    "Missing",
				AffectedSite:   siteData.LocationID,
				Severity:       "medium",
				Description:    fmt.Sprintf("Field '%s' is missing from site %s", field, siteData.LocationID),
				RootCause:      "Data extraction failure or field not present",
				Recommendation: "Investigate data extraction process and verify field availability",
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	if len(numericValues) < 3 {
		return anomalies // Need at least 3 values for statistical analysis
	}

	// Calculate statistics
	mean := s.calculateMean(numericValues)
	stdDev := s.calculateStandardDeviation(numericValues, mean)

	// Detect outliers (values more than 1.0 standard deviations from mean)
	for siteID, value := range valueMap {
		zScore := math.Abs((value - mean) / stdDev)
		if zScore > 1.0 {
			anomaly := DataAnomaly{
				ID:             generateID(),
				AnomalyType:    "outlier",
				FieldName:      field,
				ExpectedValue:  fmt.Sprintf("%.2f ± %.2f", mean, 2*stdDev),
				ActualValue:    value,
				AffectedSite:   siteID,
				Severity:       s.determineAnomalySeverity(zScore),
				Description:    fmt.Sprintf("Value %.2f is %.2f standard deviations from mean (%.2f)", value, zScore, mean),
				RootCause:      "Statistical outlier detected",
				Recommendation: "Verify data accuracy and investigate cause of deviation",
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// analyzeTrends analyzes trends in the data
func (s *CrossSiteCorrelationService) analyzeTrends(sitesData []SiteData, fields []string) []DataTrend {
	var trends []DataTrend

	for _, field := range fields {
		fieldTrends := s.analyzeFieldTrends(sitesData, field)
		trends = append(trends, fieldTrends...)

		// Limit trends per field
		if len(fieldTrends) > s.config.MaxTrendsPerField {
			trends = trends[:s.config.MaxTrendsPerField]
		}
	}

	return trends
}

// analyzeFieldTrends analyzes trends for a specific field
func (s *CrossSiteCorrelationService) analyzeFieldTrends(sitesData []SiteData, field string) []DataTrend {
	var trends []DataTrend

	// Extract numeric values with timestamps
	var values []valueWithTime
	for _, siteData := range sitesData {
		if value, exists := siteData.ExtractedData[field]; exists {
			if numericValue := s.extractNumericValue(value); numericValue != nil {
				values = append(values, valueWithTime{
					value:     *numericValue,
					timestamp: siteData.LastExtracted,
					siteID:    siteData.LocationID,
				})
			}
		}
	}

	if len(values) < 3 {
		return trends // Need at least 3 values for trend analysis
	}

	// Sort by timestamp
	sort.Slice(values, func(i, j int) bool {
		return values[i].timestamp.Before(values[j].timestamp)
	})

	// Calculate trend
	trend := s.calculateLinearTrend(values)
	if trend != nil {
		trends = append(trends, *trend)
	}

	return trends
}

// calculateLinearTrend calculates linear trend from time series data
func (s *CrossSiteCorrelationService) calculateLinearTrend(values []valueWithTime) *DataTrend {
	if len(values) < 3 {
		return nil
	}

	// Convert timestamps to numeric values for regression
	var x, y []float64
	for i, v := range values {
		x = append(x, float64(i))
		y = append(y, v.value)
	}

	// Calculate linear regression
	slope, intercept := s.calculateLinearRegression(x, y)

	// Determine trend direction and magnitude
	direction := "stable"
	magnitude := math.Abs(slope)

	if slope > 0.01 {
		direction = "increasing"
	} else if slope < -0.01 {
		direction = "decreasing"
	}

	// Calculate confidence based on R-squared
	rSquared := s.calculateRSquared(x, y, slope, intercept)
	confidence := math.Min(rSquared, 0.95) // Cap at 0.95

	if confidence < 0.3 {
		return nil // Low confidence trend
	}

	// Get affected sites
	affectedSites := make([]string, len(values))
	for i, v := range values {
		affectedSites[i] = v.siteID
	}

	// Calculate timeframe
	timeframe := values[len(values)-1].timestamp.Sub(values[0].timestamp)

	return &DataTrend{
		ID:            generateID(),
		TrendType:     "linear",
		FieldName:     "temporal_analysis", // This would be the actual field name
		Direction:     direction,
		Magnitude:     magnitude,
		Confidence:    confidence,
		Timeframe:     timeframe,
		AffectedSites: affectedSites,
		Description:   fmt.Sprintf("Linear %s trend with magnitude %.3f over %v", direction, magnitude, timeframe),
		Prediction:    fmt.Sprintf("Next value: %.2f", slope*float64(len(values))+intercept),
	}
}

// generateInsights generates insights from correlation analysis
func (s *CrossSiteCorrelationService) generateInsights(
	sitesData []SiteData,
	correlationMatrix map[string]map[string]float64,
	patterns []DataPattern,
	anomalies []DataAnomaly,
	trends []DataTrend,
) []DataInsight {
	var insights []DataInsight

	// Generate insights from correlation matrix
	correlationInsights := s.generateCorrelationInsights(correlationMatrix)
	insights = append(insights, correlationInsights...)

	// Generate insights from patterns
	patternInsights := s.generatePatternInsights(patterns)
	insights = append(insights, patternInsights...)

	// Generate insights from anomalies
	anomalyInsights := s.generateAnomalyInsights(anomalies)
	insights = append(insights, anomalyInsights...)

	// Generate insights from trends
	trendInsights := s.generateTrendInsights(trends)
	insights = append(insights, trendInsights...)

	// Limit total insights
	if len(insights) > s.config.MaxInsightsPerAnalysis {
		insights = insights[:s.config.MaxInsightsPerAnalysis]
	}

	return insights
}

// generateCorrelationInsights generates insights from correlation matrix
func (s *CrossSiteCorrelationService) generateCorrelationInsights(correlationMatrix map[string]map[string]float64) []DataInsight {
	var insights []DataInsight

	// Find strong correlations
	for field1, correlations := range correlationMatrix {
		for field2, correlation := range correlations {
			if field1 != field2 && math.Abs(correlation) > 0.7 {
				insight := DataInsight{
					ID:          generateID(),
					InsightType: "correlation",
					Title:       fmt.Sprintf("Strong correlation between %s and %s", field1, field2),
					Description: fmt.Sprintf("Fields '%s' and '%s' show strong %s correlation (%.2f)",
						field1, field2,
						s.getCorrelationDirection(correlation),
						math.Abs(correlation)),
					Confidence:     math.Abs(correlation),
					Impact:         "medium",
					Recommendation: "Consider using one field to predict or validate the other",
					RelatedFields:  []string{field1, field2},
				}
				insights = append(insights, insight)
			}
		}
	}

	return insights
}

// generatePatternInsights generates insights from patterns
func (s *CrossSiteCorrelationService) generatePatternInsights(patterns []DataPattern) []DataInsight {
	var insights []DataInsight

	for _, pattern := range patterns {
		insight := DataInsight{
			ID:             generateID(),
			InsightType:    "pattern",
			Title:          fmt.Sprintf("Data pattern detected in %s", pattern.FieldName),
			Description:    pattern.Description,
			Confidence:     pattern.Confidence,
			Impact:         pattern.Significance,
			Recommendation: pattern.Recommendation,
			RelatedFields:  []string{pattern.FieldName},
			AffectedSites:  pattern.AffectedSites,
		}
		insights = append(insights, insight)
	}

	return insights
}

// generateAnomalyInsights generates insights from anomalies
func (s *CrossSiteCorrelationService) generateAnomalyInsights(anomalies []DataAnomaly) []DataInsight {
	var insights []DataInsight

	for _, anomaly := range anomalies {
		insight := DataInsight{
			ID:             generateID(),
			InsightType:    "anomaly",
			Title:          fmt.Sprintf("Anomaly detected in %s", anomaly.FieldName),
			Description:    anomaly.Description,
			Confidence:     0.8, // High confidence for detected anomalies
			Impact:         anomaly.Severity,
			Recommendation: anomaly.Recommendation,
			RelatedFields:  []string{anomaly.FieldName},
			AffectedSites:  []string{anomaly.AffectedSite},
		}
		insights = append(insights, insight)
	}

	return insights
}

// generateTrendInsights generates insights from trends
func (s *CrossSiteCorrelationService) generateTrendInsights(trends []DataTrend) []DataInsight {
	var insights []DataInsight

	for _, trend := range trends {
		insight := DataInsight{
			ID:             generateID(),
			InsightType:    "trend",
			Title:          fmt.Sprintf("Trend detected in %s", trend.FieldName),
			Description:    trend.Description,
			Confidence:     trend.Confidence,
			Impact:         "medium",
			Recommendation: "Monitor trend continuation and investigate underlying causes",
			RelatedFields:  []string{trend.FieldName},
			AffectedSites:  trend.AffectedSites,
		}
		insights = append(insights, insight)
	}

	return insights
}

// =============================================================================
// Utility Methods
// =============================================================================

// calculateConfidenceScore calculates overall confidence score for the analysis
func (s *CrossSiteCorrelationService) calculateConfidenceScore(
	correlationMatrix map[string]map[string]float64,
	patterns []DataPattern,
	anomalies []DataAnomaly,
	trends []DataTrend,
	insights []DataInsight,
) float64 {
	// Calculate confidence based on various factors
	var totalConfidence float64
	var weightSum float64

	// Correlation matrix confidence (weight: 0.3)
	if len(correlationMatrix) > 0 {
		correlationConfidence := s.calculateCorrelationConfidence(correlationMatrix)
		totalConfidence += correlationConfidence * 0.3
		weightSum += 0.3
	}

	// Pattern detection confidence (weight: 0.25)
	if len(patterns) > 0 {
		patternConfidence := s.calculatePatternConfidence(patterns)
		totalConfidence += patternConfidence * 0.25
		weightSum += 0.25
	}

	// Anomaly detection confidence (weight: 0.25)
	if len(anomalies) > 0 {
		anomalyConfidence := s.calculateAnomalyConfidence(anomalies)
		totalConfidence += anomalyConfidence * 0.25
		weightSum += 0.25
	}

	// Trend analysis confidence (weight: 0.2)
	if len(trends) > 0 {
		trendConfidence := s.calculateTrendConfidence(trends)
		totalConfidence += trendConfidence * 0.2
		weightSum += 0.2
	}

	if weightSum == 0 {
		return 0.0
	}

	return totalConfidence / weightSum
}

// calculateCorrelationConfidence calculates confidence for correlation analysis
func (s *CrossSiteCorrelationService) calculateCorrelationConfidence(correlationMatrix map[string]map[string]float64) float64 {
	if len(correlationMatrix) == 0 {
		return 0.0
	}

	var totalCorrelation float64
	var correlationCount int

	for _, correlations := range correlationMatrix {
		for _, correlation := range correlations {
			if correlation != 1.0 { // Exclude self-correlations
				totalCorrelation += math.Abs(correlation)
				correlationCount++
			}
		}
	}

	if correlationCount == 0 {
		return 0.0
	}

	return totalCorrelation / float64(correlationCount)
}

// calculatePatternConfidence calculates confidence for pattern detection
func (s *CrossSiteCorrelationService) calculatePatternConfidence(patterns []DataPattern) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	var totalConfidence float64
	for _, pattern := range patterns {
		totalConfidence += pattern.Confidence
	}

	return totalConfidence / float64(len(patterns))
}

// calculateAnomalyConfidence calculates confidence for anomaly detection
func (s *CrossSiteCorrelationService) calculateAnomalyConfidence(anomalies []DataAnomaly) float64 {
	if len(anomalies) == 0 {
		return 0.0
	}

	// Anomaly detection confidence is based on severity distribution
	var highSeverityCount int
	for _, anomaly := range anomalies {
		if anomaly.Severity == "high" || anomaly.Severity == "critical" {
			highSeverityCount++
		}
	}

	// Higher confidence if we detect meaningful anomalies
	return math.Min(0.9, 0.5+float64(highSeverityCount)/float64(len(anomalies))*0.4)
}

// calculateTrendConfidence calculates confidence for trend analysis
func (s *CrossSiteCorrelationService) calculateTrendConfidence(trends []DataTrend) float64 {
	if len(trends) == 0 {
		return 0.0
	}

	var totalConfidence float64
	for _, trend := range trends {
		totalConfidence += trend.Confidence
	}

	return totalConfidence / float64(len(trends))
}

// calculateMean calculates mean of values
func (s *CrossSiteCorrelationService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// calculateStandardDeviation calculates standard deviation
func (s *CrossSiteCorrelationService) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	var sumSquaredDiff float64
	for _, value := range values {
		diff := value - mean
		sumSquaredDiff += diff * diff
	}

	return math.Sqrt(sumSquaredDiff / float64(len(values)-1))
}

// calculateLinearRegression calculates linear regression parameters
func (s *CrossSiteCorrelationService) calculateLinearRegression(x, y []float64) (slope, intercept float64) {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0, 0.0
	}

	n := float64(len(x))

	var sumX, sumY, sumXY, sumX2 float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0.0, 0.0
	}

	slope = (n*sumXY - sumX*sumY) / denominator
	intercept = (sumY - slope*sumX) / n

	return slope, intercept
}

// calculateRSquared calculates R-squared value
func (s *CrossSiteCorrelationService) calculateRSquared(x, y []float64, slope, intercept float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	meanY := s.calculateMean(y)

	var ssRes, ssTot float64
	for i := range x {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	if ssTot == 0 {
		return 0.0
	}

	return 1 - (ssRes / ssTot)
}

// determineAnomalySeverity determines severity based on z-score
func (s *CrossSiteCorrelationService) determineAnomalySeverity(zScore float64) string {
	if zScore > 3.0 {
		return "critical"
	} else if zScore > 2.5 {
		return "high"
	} else if zScore > 2.0 {
		return "medium"
	}
	return "low"
}

// getCorrelationDirection returns correlation direction description
func (s *CrossSiteCorrelationService) getCorrelationDirection(correlation float64) string {
	if correlation > 0.7 {
		return "positive"
	} else if correlation < -0.7 {
		return "negative"
	}
	return "weak"
}

// parseNumericString attempts to parse string as numeric value
func parseNumericString(s string) (float64, error) {
	// Remove common non-numeric characters
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "%", "")

	// Try to parse as float
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// extractNumericFromString extracts numeric patterns from string
func extractNumericFromString(s string) (float64, error) {
	// This is a simplified implementation
	// In a real system, you might use regex patterns to extract numbers
	return 0.0, fmt.Errorf("numeric extraction not implemented")
}
