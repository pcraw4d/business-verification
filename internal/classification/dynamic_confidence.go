package classification

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ContentQuality represents the quality assessment of content used for classification
type ContentQuality struct {
	Completeness      float64 `json:"completeness"`       // How complete the content is (0-1)
	Relevance         float64 `json:"relevance"`          // How relevant the content is to business classification (0-1)
	Freshness         float64 `json:"freshness"`          // How recent the content is (0-1)
	Accuracy          float64 `json:"accuracy"`           // How accurate the content appears to be (0-1)
	Consistency       float64 `json:"consistency"`        // How consistent the content is across sources (0-1)
	SourceReliability float64 `json:"source_reliability"` // Reliability of the content source (0-1)
}

// GeographicRegion represents geographic information for confidence adjustment
type GeographicRegion struct {
	Country     string  `json:"country"`
	State       string  `json:"state"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Confidence  float64 `json:"confidence"`   // Confidence in geographic detection
	DataQuality float64 `json:"data_quality"` // Quality of geographic data
}

// IndustrySpecificFactors represents industry-specific confidence adjustment factors
type IndustrySpecificFactors struct {
	IndustryCode     string  `json:"industry_code"`
	IndustryCategory string  `json:"industry_category"`
	CodeDensity      float64 `json:"code_density"`    // Number of similar codes in the industry
	ValidationRate   float64 `json:"validation_rate"` // Historical validation rate for this industry
	Popularity       float64 `json:"popularity"`      // Industry popularity/occurrence rate
	Complexity       float64 `json:"complexity"`      // Industry complexity factor
}

// ConfidenceAdjustmentFactors represents all factors that can adjust confidence scores
type ConfidenceAdjustmentFactors struct {
	ContentQuality    *ContentQuality          `json:"content_quality"`
	GeographicRegion  *GeographicRegion        `json:"geographic_region"`
	IndustryFactors   *IndustrySpecificFactors `json:"industry_factors"`
	BusinessSize      string                   `json:"business_size"`
	BusinessAge       int                      `json:"business_age"`
	DataSourceQuality float64                  `json:"data_source_quality"`
	CrossValidation   float64                  `json:"cross_validation"`
	LastUpdated       time.Time                `json:"last_updated"`
}

// DynamicConfidenceAdjuster provides dynamic confidence adjustment based on various factors
type DynamicConfidenceAdjuster struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration weights for different adjustment factors
	contentQualityWeight    float64
	geographicRegionWeight  float64
	industrySpecificWeight  float64
	businessSizeWeight      float64
	businessAgeWeight       float64
	dataSourceQualityWeight float64
	crossValidationWeight   float64
	recencyWeight           float64

	// Geographic region confidence modifiers
	geographicConfidenceModifiers map[string]float64

	// Industry-specific confidence adjustments
	industryConfidenceAdjustments map[string]float64

	// Business size confidence modifiers
	businessSizeModifiers map[string]float64

	// Data source quality thresholds
	dataSourceQualityThresholds map[string]float64
}

// NewDynamicConfidenceAdjuster creates a new dynamic confidence adjuster
func NewDynamicConfidenceAdjuster(logger *observability.Logger, metrics *observability.Metrics) *DynamicConfidenceAdjuster {
	adjuster := &DynamicConfidenceAdjuster{
		logger:  logger,
		metrics: metrics,

		// Configuration weights
		contentQualityWeight:    0.25,
		geographicRegionWeight:  0.20,
		industrySpecificWeight:  0.20,
		businessSizeWeight:      0.10,
		businessAgeWeight:       0.05,
		dataSourceQualityWeight: 0.10,
		crossValidationWeight:   0.05,
		recencyWeight:           0.05,

		// Initialize maps
		geographicConfidenceModifiers: make(map[string]float64),
		industryConfidenceAdjustments: make(map[string]float64),
		businessSizeModifiers:         make(map[string]float64),
		dataSourceQualityThresholds:   make(map[string]float64),
	}

	// Initialize default values
	adjuster.initializeDefaultValues()

	return adjuster
}

// initializeDefaultValues initializes default confidence adjustment values
func (d *DynamicConfidenceAdjuster) initializeDefaultValues() {
	// Geographic region confidence modifiers
	d.geographicConfidenceModifiers = map[string]float64{
		"US": 1.0,  // United States - baseline
		"CA": 0.95, // Canada - slightly lower
		"UK": 0.95, // United Kingdom - slightly lower
		"AU": 0.90, // Australia - lower
		"DE": 0.90, // Germany - lower
		"FR": 0.85, // France - lower
		"JP": 0.85, // Japan - lower
		"CN": 0.80, // China - lower
		"IN": 0.75, // India - lower
		"BR": 0.75, // Brazil - lower
	}

	// Industry-specific confidence adjustments
	d.industryConfidenceAdjustments = map[string]float64{
		"11": 0.95, // Agriculture - high confidence
		"21": 0.90, // Mining - high confidence
		"22": 0.95, // Utilities - very high confidence
		"23": 0.85, // Construction - medium-high confidence
		"31": 0.80, // Manufacturing - medium confidence
		"32": 0.80, // Manufacturing - medium confidence
		"33": 0.80, // Manufacturing - medium confidence
		"42": 0.85, // Wholesale Trade - medium-high confidence
		"44": 0.75, // Retail Trade - medium confidence
		"45": 0.75, // Retail Trade - medium confidence
		"48": 0.90, // Transportation - high confidence
		"49": 0.90, // Transportation - high confidence
		"51": 0.85, // Information - medium-high confidence
		"52": 0.90, // Finance and Insurance - high confidence
		"53": 0.85, // Real Estate - medium-high confidence
		"54": 0.80, // Professional Services - medium confidence
		"55": 0.85, // Management - medium-high confidence
		"56": 0.80, // Administrative Services - medium confidence
		"61": 0.90, // Educational Services - high confidence
		"62": 0.90, // Health Care - high confidence
		"71": 0.80, // Arts and Entertainment - medium confidence
		"72": 0.75, // Accommodation and Food Services - medium confidence
		"81": 0.75, // Other Services - medium confidence
		"92": 0.90, // Public Administration - high confidence
	}

	// Business size confidence modifiers
	d.businessSizeModifiers = map[string]float64{
		"large":   1.0,  // Large businesses - baseline
		"medium":  0.95, // Medium businesses - slightly lower
		"small":   0.90, // Small businesses - lower
		"micro":   0.85, // Micro businesses - lower
		"startup": 0.80, // Startups - lower
		"unknown": 0.75, // Unknown size - lowest
	}

	// Data source quality thresholds
	d.dataSourceQualityThresholds = map[string]float64{
		"official_registry":   1.0,  // Official business registry
		"government_database": 0.95, // Government database
		"credit_bureau":       0.90, // Credit bureau
		"business_directory":  0.85, // Business directory
		"social_media":        0.70, // Social media
		"user_generated":      0.60, // User-generated content
		"scraped_data":        0.75, // Scraped data
		"unknown_source":      0.50, // Unknown source
	}
}

// AdjustConfidence dynamically adjusts confidence scores based on various factors
func (d *DynamicConfidenceAdjuster) AdjustConfidence(ctx context.Context, baseConfidence float64, factors *ConfidenceAdjustmentFactors) float64 {
	start := time.Now()

	// Log adjustment start
	if d.logger != nil {
		d.logger.WithComponent("dynamic_confidence").LogBusinessEvent(ctx, "confidence_adjustment_started", "", map[string]interface{}{
			"base_confidence": baseConfidence,
		})
	}

	// Calculate adjustment factors
	contentQualityAdjustment := d.calculateContentQualityAdjustment(factors.ContentQuality)
	geographicAdjustment := d.calculateGeographicAdjustment(factors.GeographicRegion)
	industryAdjustment := d.calculateIndustryAdjustment(factors.IndustryFactors)
	businessSizeAdjustment := d.calculateBusinessSizeAdjustment(factors.BusinessSize)
	businessAgeAdjustment := d.calculateBusinessAgeAdjustment(factors.BusinessAge)
	dataSourceAdjustment := d.calculateDataSourceAdjustment(factors.DataSourceQuality)
	crossValidationAdjustment := d.calculateCrossValidationAdjustment(factors.CrossValidation)
	recencyAdjustment := d.calculateRecencyAdjustment(factors.LastUpdated)

	// Calculate weighted adjustment
	totalAdjustment := (contentQualityAdjustment * d.contentQualityWeight) +
		(geographicAdjustment * d.geographicRegionWeight) +
		(industryAdjustment * d.industrySpecificWeight) +
		(businessSizeAdjustment * d.businessSizeWeight) +
		(businessAgeAdjustment * d.businessAgeWeight) +
		(dataSourceAdjustment * d.dataSourceQualityWeight) +
		(crossValidationAdjustment * d.crossValidationWeight) +
		(recencyAdjustment * d.recencyWeight)

	// Apply adjustment to base confidence
	adjustedConfidence := baseConfidence * totalAdjustment

	// Ensure confidence is within valid range
	adjustedConfidence = math.Max(0.0, math.Min(1.0, adjustedConfidence))

	// Log adjustment completion
	if d.logger != nil {
		d.logger.WithComponent("dynamic_confidence").LogBusinessEvent(ctx, "confidence_adjustment_completed", "", map[string]interface{}{
			"base_confidence":     baseConfidence,
			"adjusted_confidence": adjustedConfidence,
			"total_adjustment":    totalAdjustment,
			"processing_time_ms":  time.Since(start).Milliseconds(),
		})
	}

	// Record adjustment metrics
	d.RecordConfidenceAdjustmentMetrics(ctx, baseConfidence, adjustedConfidence, factors)

	return adjustedConfidence
}

// calculateContentQualityAdjustment calculates confidence adjustment based on content quality
func (d *DynamicConfidenceAdjuster) calculateContentQualityAdjustment(quality *ContentQuality) float64 {
	if quality == nil {
		return 0.75 // Default adjustment for unknown content quality
	}

	// Calculate weighted content quality score
	qualityScore := (quality.Completeness * 0.25) +
		(quality.Relevance * 0.25) +
		(quality.Freshness * 0.15) +
		(quality.Accuracy * 0.20) +
		(quality.Consistency * 0.10) +
		(quality.SourceReliability * 0.05)

	// Convert quality score to adjustment factor (0.5 to 1.5 range)
	adjustment := 0.5 + (qualityScore * 1.0)

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateGeographicAdjustment calculates confidence adjustment based on geographic region
func (d *DynamicConfidenceAdjuster) calculateGeographicAdjustment(region *GeographicRegion) float64 {
	if region == nil {
		return 0.85 // Default adjustment for unknown region
	}

	// Get base modifier for country
	countryModifier, exists := d.geographicConfidenceModifiers[region.Country]
	if !exists {
		countryModifier = 0.80 // Default for unknown countries
	}

	// Adjust based on data quality
	dataQualityAdjustment := 0.8 + (region.DataQuality * 0.4) // 0.8 to 1.2 range

	// Calculate final adjustment
	adjustment := countryModifier * dataQualityAdjustment

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateIndustryAdjustment calculates confidence adjustment based on industry-specific factors
func (d *DynamicConfidenceAdjuster) calculateIndustryAdjustment(factors *IndustrySpecificFactors) float64 {
	if factors == nil {
		return 0.90 // Default adjustment for unknown industry
	}

	// Get base adjustment for industry
	industryAdjustment, exists := d.industryConfidenceAdjustments[factors.IndustryCode]
	if !exists {
		industryAdjustment = 0.85 // Default for unknown industries
	}

	// Adjust based on code density (higher density = lower confidence)
	densityAdjustment := 1.0 - (factors.CodeDensity * 0.2) // 0.8 to 1.0 range

	// Adjust based on validation rate
	validationAdjustment := 0.8 + (factors.ValidationRate * 0.4) // 0.8 to 1.2 range

	// Adjust based on complexity (higher complexity = lower confidence)
	complexityAdjustment := 1.0 - (factors.Complexity * 0.3) // 0.7 to 1.0 range

	// Calculate final adjustment
	adjustment := industryAdjustment * densityAdjustment * validationAdjustment * complexityAdjustment

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateBusinessSizeAdjustment calculates confidence adjustment based on business size
func (d *DynamicConfidenceAdjuster) calculateBusinessSizeAdjustment(businessSize string) float64 {
	if businessSize == "" {
		businessSize = "unknown"
	}

	modifier, exists := d.businessSizeModifiers[strings.ToLower(businessSize)]
	if !exists {
		modifier = 0.75 // Default for unknown sizes
	}

	// Convert modifier to adjustment factor (0.7 to 1.3 range)
	adjustment := 0.7 + (modifier * 0.6)

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateBusinessAgeAdjustment calculates confidence adjustment based on business age
func (d *DynamicConfidenceAdjuster) calculateBusinessAgeAdjustment(businessAge int) float64 {
	if businessAge <= 0 {
		return 0.80 // Default for unknown age
	}

	// Newer businesses (0-2 years) get slightly lower confidence
	if businessAge <= 2 {
		return 0.85
	}

	// Established businesses (3-10 years) get standard confidence
	if businessAge <= 10 {
		return 1.0
	}

	// Very established businesses (10+ years) get slightly higher confidence
	return 1.05
}

// calculateDataSourceAdjustment calculates confidence adjustment based on data source quality
func (d *DynamicConfidenceAdjuster) calculateDataSourceAdjustment(dataSourceQuality float64) float64 {
	// Convert data source quality to adjustment factor (0.6 to 1.4 range)
	adjustment := 0.6 + (dataSourceQuality * 0.8)

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateCrossValidationAdjustment calculates confidence adjustment based on cross-validation
func (d *DynamicConfidenceAdjuster) calculateCrossValidationAdjustment(crossValidation float64) float64 {
	// Convert cross-validation score to adjustment factor (0.8 to 1.2 range)
	adjustment := 0.8 + (crossValidation * 0.4)

	return math.Max(0.5, math.Min(1.5, adjustment))
}

// calculateRecencyAdjustment calculates confidence adjustment based on data recency
func (d *DynamicConfidenceAdjuster) calculateRecencyAdjustment(lastUpdated time.Time) float64 {
	if lastUpdated.IsZero() {
		return 0.85 // Default for unknown update time
	}

	// Calculate days since last update
	daysSinceUpdate := time.Since(lastUpdated).Hours() / 24

	// Very recent data (0-30 days) gets highest confidence
	if daysSinceUpdate <= 30 {
		return 1.05
	}

	// Recent data (31-90 days) gets high confidence
	if daysSinceUpdate <= 90 {
		return 1.0
	}

	// Older data (91-365 days) gets standard confidence
	if daysSinceUpdate <= 365 {
		return 0.95
	}

	// Very old data (365+ days) gets lower confidence
	return 0.85
}

// RecordConfidenceAdjustmentMetrics records metrics for confidence adjustment
func (d *DynamicConfidenceAdjuster) RecordConfidenceAdjustmentMetrics(ctx context.Context, baseConfidence, adjustedConfidence float64, factors *ConfidenceAdjustmentFactors) {
	if d.metrics == nil {
		return
	}

	// Record confidence adjustment metrics
	d.metrics.RecordHistogram(ctx, "confidence_adjustment_base", baseConfidence, map[string]string{
		"component": "dynamic_confidence",
	})

	d.metrics.RecordHistogram(ctx, "confidence_adjustment_adjusted", adjustedConfidence, map[string]string{
		"component": "dynamic_confidence",
	})

	d.metrics.RecordHistogram(ctx, "confidence_adjustment_delta", adjustedConfidence-baseConfidence, map[string]string{
		"component": "dynamic_confidence",
	})

	// Record adjustment factors if available
	if factors != nil && factors.GeographicRegion != nil {
		d.metrics.RecordHistogram(ctx, "confidence_adjustment_geographic_quality", factors.GeographicRegion.DataQuality, map[string]string{
			"component": "dynamic_confidence",
			"country":   factors.GeographicRegion.Country,
		})
	}

	if factors != nil && factors.ContentQuality != nil {
		d.metrics.RecordHistogram(ctx, "confidence_adjustment_content_quality",
			(factors.ContentQuality.Completeness+factors.ContentQuality.Relevance+factors.ContentQuality.Accuracy)/3,
			map[string]string{
				"component": "dynamic_confidence",
			})
	}
}

// UpdateGeographicConfidenceModifier updates the confidence modifier for a specific country
func (d *DynamicConfidenceAdjuster) UpdateGeographicConfidenceModifier(country string, modifier float64) {
	d.geographicConfidenceModifiers[country] = math.Max(0.5, math.Min(1.5, modifier))
}

// UpdateIndustryConfidenceAdjustment updates the confidence adjustment for a specific industry
func (d *DynamicConfidenceAdjuster) UpdateIndustryConfidenceAdjustment(industryCode string, adjustment float64) {
	d.industryConfidenceAdjustments[industryCode] = math.Max(0.5, math.Min(1.5, adjustment))
}

// UpdateBusinessSizeModifier updates the confidence modifier for a specific business size
func (d *DynamicConfidenceAdjuster) UpdateBusinessSizeModifier(size string, modifier float64) {
	d.businessSizeModifiers[strings.ToLower(size)] = math.Max(0.5, math.Min(1.5, modifier))
}

// GetAdjustmentFactors returns the current adjustment factors for debugging and monitoring
func (d *DynamicConfidenceAdjuster) GetAdjustmentFactors() map[string]interface{} {
	return map[string]interface{}{
		"geographic_modifiers":    d.geographicConfidenceModifiers,
		"industry_adjustments":    d.industryConfidenceAdjustments,
		"business_size_modifiers": d.businessSizeModifiers,
		"data_source_thresholds":  d.dataSourceQualityThresholds,
		"weights": map[string]float64{
			"content_quality":     d.contentQualityWeight,
			"geographic_region":   d.geographicRegionWeight,
			"industry_specific":   d.industrySpecificWeight,
			"business_size":       d.businessSizeWeight,
			"business_age":        d.businessAgeWeight,
			"data_source_quality": d.dataSourceQualityWeight,
			"cross_validation":    d.crossValidationWeight,
			"recency":             d.recencyWeight,
		},
	}
}
