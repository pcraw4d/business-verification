package classification_monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MethodDimensionCollector collects metrics by classification method
type MethodDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (mdc *MethodDimensionCollector) GetDimensionName() string {
	return mdc.name
}

// CollectMetrics collects metrics for each classification method
func (mdc *MethodDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			classification_method,
			id, business_name, actual_classification, expected_classification,
			confidence_score, processing_time_ms, is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY classification_method, created_at DESC
	`

	rows, err := mdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications by method: %w", err)
	}
	defer rows.Close()

	methodData := make(map[string][]*ClassificationData)

	for rows.Next() {
		var method string
		var c ClassificationData
		var processingTimeMs sql.NullInt64
		var expectedClassification sql.NullString
		var isCorrect sql.NullBool
		var metadataJSON sql.NullString

		err := rows.Scan(
			&method,
			&c.ID, &c.BusinessName, &c.ActualIndustry, &expectedClassification,
			&c.ConfidenceScore, &processingTimeMs, &isCorrect, &c.Timestamp, &metadataJSON,
		)
		if err != nil {
			mdc.logger.Error("Failed to scan method classification row", zap.Error(err))
			continue
		}

		c.Method = method

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

		methodData[method] = append(methodData[method], &c)
	}

	var result []*DimensionMetrics
	for method, classifications := range methodData {
		metrics := &DimensionMetrics{
			DimensionValue:  method,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// GetSupportedAggregations returns supported aggregation types
func (mdc *MethodDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "confidence", "response_time", "volume"}
}

// ConfidenceRangeDimensionCollector collects metrics by confidence range
type ConfidenceRangeDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (crdc *ConfidenceRangeDimensionCollector) GetDimensionName() string {
	return crdc.name
}

// CollectMetrics collects metrics for each confidence range
func (crdc *ConfidenceRangeDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			id, business_name, actual_classification, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := crdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications for confidence ranges: %w", err)
	}
	defer rows.Close()

	confidenceRangeData := make(map[string][]*ClassificationData)

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
			crdc.logger.Error("Failed to scan confidence range classification row", zap.Error(err))
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

		// Determine confidence range
		confidenceRange := crdc.getConfidenceRange(c.ConfidenceScore)
		confidenceRangeData[confidenceRange] = append(confidenceRangeData[confidenceRange], &c)
	}

	var result []*DimensionMetrics
	for confidenceRange, classifications := range confidenceRangeData {
		metrics := &DimensionMetrics{
			DimensionValue:  confidenceRange,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// getConfidenceRange determines the confidence range for a confidence score
func (crdc *ConfidenceRangeDimensionCollector) getConfidenceRange(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "very_high"
	case confidence >= 0.7:
		return "high"
	case confidence >= 0.5:
		return "medium"
	case confidence >= 0.3:
		return "low"
	default:
		return "very_low"
	}
}

// GetSupportedAggregations returns supported aggregation types
func (crdc *ConfidenceRangeDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "volume", "distribution"}
}

// IndustryDimensionCollector collects metrics by industry
type IndustryDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (idc *IndustryDimensionCollector) GetDimensionName() string {
	return idc.name
}

// CollectMetrics collects metrics for each industry
func (idc *IndustryDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			actual_classification,
			id, business_name, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY actual_classification, created_at DESC
	`

	rows, err := idc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications by industry: %w", err)
	}
	defer rows.Close()

	industryData := make(map[string][]*ClassificationData)

	for rows.Next() {
		var industry string
		var c ClassificationData
		var processingTimeMs sql.NullInt64
		var expectedClassification sql.NullString
		var isCorrect sql.NullBool
		var metadataJSON sql.NullString

		err := rows.Scan(
			&industry,
			&c.ID, &c.BusinessName, &expectedClassification,
			&c.ConfidenceScore, &c.Method, &processingTimeMs,
			&isCorrect, &c.Timestamp, &metadataJSON,
		)
		if err != nil {
			idc.logger.Error("Failed to scan industry classification row", zap.Error(err))
			continue
		}

		c.ActualIndustry = industry

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

		industryData[industry] = append(industryData[industry], &c)
	}

	var result []*DimensionMetrics
	for industry, classifications := range industryData {
		metrics := &DimensionMetrics{
			DimensionValue:  industry,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// GetSupportedAggregations returns supported aggregation types
func (idc *IndustryDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "confidence", "volume", "misclassification_patterns"}
}

// TimeDimensionCollector collects metrics by time of day
type TimeDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (tdc *TimeDimensionCollector) GetDimensionName() string {
	return tdc.name
}

// CollectMetrics collects metrics for each time period
func (tdc *TimeDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			id, business_name, actual_classification, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := tdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications for time analysis: %w", err)
	}
	defer rows.Close()

	timeData := make(map[string][]*ClassificationData)

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
			tdc.logger.Error("Failed to scan time classification row", zap.Error(err))
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

		// Determine time category
		timeCategory := tdc.getTimeCategory(c.Timestamp)
		timeData[timeCategory] = append(timeData[timeCategory], &c)
	}

	var result []*DimensionMetrics
	for timeCategory, classifications := range timeData {
		metrics := &DimensionMetrics{
			DimensionValue:  timeCategory,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// getTimeCategory determines the time category for a timestamp
func (tdc *TimeDimensionCollector) getTimeCategory(timestamp time.Time) string {
	hour := timestamp.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

// GetSupportedAggregations returns supported aggregation types
func (tdc *TimeDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "volume", "temporal_patterns"}
}

// CustomDimensionCollector allows for custom dimension collection
type CustomDimensionCollector struct {
	name                  string
	db                    *sql.DB
	logger                *zap.Logger
	dimensionFunction     func(*ClassificationData) string
	query                 string
	supportedAggregations []string
}

// NewCustomDimensionCollector creates a new custom dimension collector
func NewCustomDimensionCollector(
	name string,
	db *sql.DB,
	logger *zap.Logger,
	dimensionFunction func(*ClassificationData) string,
	query string,
	supportedAggregations []string,
) *CustomDimensionCollector {
	return &CustomDimensionCollector{
		name:                  name,
		db:                    db,
		logger:                logger,
		dimensionFunction:     dimensionFunction,
		query:                 query,
		supportedAggregations: supportedAggregations,
	}
}

// GetDimensionName returns the dimension name
func (cdc *CustomDimensionCollector) GetDimensionName() string {
	return cdc.name
}

// CollectMetrics collects metrics using the custom dimension function
func (cdc *CustomDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := cdc.query
	if query == "" {
		query = `
			SELECT 
				id, business_name, actual_classification, expected_classification,
				confidence_score, classification_method, processing_time_ms,
				is_correct, created_at, metadata
			FROM classifications 
			WHERE created_at BETWEEN $1 AND $2
			ORDER BY created_at DESC
		`
	}

	rows, err := cdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications for custom dimension: %w", err)
	}
	defer rows.Close()

	dimensionData := make(map[string][]*ClassificationData)

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
			cdc.logger.Error("Failed to scan custom dimension classification row", zap.Error(err))
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

		// Apply custom dimension function
		dimensionValue := cdc.dimensionFunction(&c)
		dimensionData[dimensionValue] = append(dimensionData[dimensionValue], &c)
	}

	var result []*DimensionMetrics
	for dimensionValue, classifications := range dimensionData {
		metrics := &DimensionMetrics{
			DimensionValue:  dimensionValue,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// GetSupportedAggregations returns supported aggregation types
func (cdc *CustomDimensionCollector) GetSupportedAggregations() []string {
	return cdc.supportedAggregations
}

// GeographicDimensionCollector collects metrics by geographic region
type GeographicDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (gdc *GeographicDimensionCollector) GetDimensionName() string {
	return gdc.name
}

// CollectMetrics collects metrics for each geographic region
func (gdc *GeographicDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			id, business_name, actual_classification, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := gdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications for geographic analysis: %w", err)
	}
	defer rows.Close()

	geographicData := make(map[string][]*ClassificationData)

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
			gdc.logger.Error("Failed to scan geographic classification row", zap.Error(err))
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

		// Extract geographic region from business name or metadata
		region := gdc.extractGeographicRegion(&c)
		geographicData[region] = append(geographicData[region], &c)
	}

	var result []*DimensionMetrics
	for region, classifications := range geographicData {
		metrics := &DimensionMetrics{
			DimensionValue:  region,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// extractGeographicRegion extracts geographic region from classification data
func (gdc *GeographicDimensionCollector) extractGeographicRegion(c *ClassificationData) string {
	businessName := strings.ToLower(c.BusinessName)

	// Simple region detection based on common patterns
	if strings.Contains(businessName, "new york") || strings.Contains(businessName, "ny") {
		return "northeast"
	} else if strings.Contains(businessName, "california") || strings.Contains(businessName, "ca") || strings.Contains(businessName, "silicon valley") {
		return "west_coast"
	} else if strings.Contains(businessName, "texas") || strings.Contains(businessName, "tx") || strings.Contains(businessName, "houston") || strings.Contains(businessName, "dallas") {
		return "south"
	} else if strings.Contains(businessName, "chicago") || strings.Contains(businessName, "detroit") || strings.Contains(businessName, "milwaukee") {
		return "midwest"
	} else if strings.Contains(businessName, "florida") || strings.Contains(businessName, "fl") || strings.Contains(businessName, "miami") {
		return "southeast"
	}

	return "unknown"
}

// GetSupportedAggregations returns supported aggregation types
func (gdc *GeographicDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "volume", "geographic_patterns"}
}

// BusinessSizeDimensionCollector collects metrics by business size
type BusinessSizeDimensionCollector struct {
	name   string
	db     *sql.DB
	logger *zap.Logger
}

// GetDimensionName returns the dimension name
func (bsdc *BusinessSizeDimensionCollector) GetDimensionName() string {
	return bsdc.name
}

// CollectMetrics collects metrics for each business size category
func (bsdc *BusinessSizeDimensionCollector) CollectMetrics(ctx context.Context, startTime, endTime time.Time) ([]*DimensionMetrics, error) {
	query := `
		SELECT 
			id, business_name, actual_classification, expected_classification,
			confidence_score, classification_method, processing_time_ms,
			is_correct, created_at, metadata
		FROM classifications 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := bsdc.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifications for business size analysis: %w", err)
	}
	defer rows.Close()

	businessSizeData := make(map[string][]*ClassificationData)

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
			bsdc.logger.Error("Failed to scan business size classification row", zap.Error(err))
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

		// Determine business size category
		sizeCategory := bsdc.determineBusinessSize(&c)
		businessSizeData[sizeCategory] = append(businessSizeData[sizeCategory], &c)
	}

	var result []*DimensionMetrics
	for sizeCategory, classifications := range businessSizeData {
		metrics := &DimensionMetrics{
			DimensionValue:  sizeCategory,
			Classifications: classifications,
			Timestamp:       time.Now(),
			AggregatedData:  make(map[string]interface{}),
		}
		result = append(result, metrics)
	}

	return result, nil
}

// determineBusinessSize determines business size from classification data
func (bsdc *BusinessSizeDimensionCollector) determineBusinessSize(c *ClassificationData) string {
	businessName := strings.ToLower(c.BusinessName)

	// Simple business size detection based on common patterns
	if strings.Contains(businessName, "corporation") ||
		strings.Contains(businessName, "corp") ||
		strings.Contains(businessName, "international") ||
		strings.Contains(businessName, "global") ||
		strings.Contains(businessName, "worldwide") {
		return "large"
	} else if strings.Contains(businessName, "llc") ||
		strings.Contains(businessName, "inc") ||
		strings.Contains(businessName, "company") ||
		strings.Contains(businessName, "enterprises") {
		return "medium"
	} else if strings.Contains(businessName, "consulting") ||
		strings.Contains(businessName, "services") ||
		strings.Contains(businessName, "solutions") {
		return "small"
	}

	return "unknown"
}

// GetSupportedAggregations returns supported aggregation types
func (bsdc *BusinessSizeDimensionCollector) GetSupportedAggregations() []string {
	return []string{"accuracy", "volume", "size_patterns"}
}
