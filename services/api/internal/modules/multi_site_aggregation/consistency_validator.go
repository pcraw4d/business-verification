package multi_site_aggregation

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConsistencyValidator defines the interface for validating data consistency across multiple sites
type ConsistencyValidator interface {
	// ValidateConsistency validates data consistency across multiple sites for a business
	ValidateConsistency(ctx context.Context, businessID string) (*ConsistencyValidationResult, error)

	// ValidateFieldConsistency validates consistency for a specific field across sites
	ValidateFieldConsistency(ctx context.Context, businessID string, fieldName string) (*FieldConsistencyResult, error)

	// GetConsistencyReport generates a comprehensive consistency report
	GetConsistencyReport(ctx context.Context, businessID string) (*ConsistencyReport, error)

	// ValidateDataIntegrity validates the integrity of aggregated data
	ValidateDataIntegrity(ctx context.Context, aggregatedData *AggregatedBusinessData) (*DataIntegrityResult, error)
}

// ConsistencyValidationResult represents the result of a consistency validation
type ConsistencyValidationResult struct {
	BusinessID        string                   `json:"business_id"`
	OverallScore      float64                  `json:"overall_score"`
	ConsistencyLevel  ConsistencyLevel         `json:"consistency_level"`
	FieldResults      []FieldConsistencyResult `json:"field_results"`
	Issues            []ConsistencyIssue       `json:"issues"`
	Recommendations   []string                 `json:"recommendations"`
	ValidatedAt       time.Time                `json:"validated_at"`
	TotalSites        int                      `json:"total_sites"`
	ConsistentSites   int                      `json:"consistent_sites"`
	InconsistentSites int                      `json:"inconsistent_sites"`
}

// FieldConsistencyResult represents consistency validation for a specific field
type FieldConsistencyResult struct {
	FieldName        string                  `json:"field_name"`
	DataType         string                  `json:"data_type"`
	ConsistencyScore float64                 `json:"consistency_score"`
	ConsistencyLevel ConsistencyLevel        `json:"consistency_level"`
	Values           []FieldValue            `json:"values"`
	IsConsistent     bool                    `json:"is_consistent"`
	Issues           []FieldConsistencyIssue `json:"issues"`
	Recommendation   string                  `json:"recommendation"`
}

// FieldValue represents a value found for a field across sites
type FieldValue struct {
	Value       interface{} `json:"value"`
	SiteID      string      `json:"site_id"`
	URL         string      `json:"url"`
	Confidence  float64     `json:"confidence"`
	LastUpdated time.Time   `json:"last_updated"`
}

// FieldConsistencyIssue represents a specific consistency issue for a field
type FieldConsistencyIssue struct {
	IssueType         string        `json:"issue_type"`
	Severity          string        `json:"severity"`
	Description       string        `json:"description"`
	AffectedSites     []string      `json:"affected_sites"`
	ConflictingValues []interface{} `json:"conflicting_values"`
}

// ConsistencyIssue represents a general consistency issue
type ConsistencyIssue struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Severity       string    `json:"severity"`
	Description    string    `json:"description"`
	AffectedFields []string  `json:"affected_fields"`
	AffectedSites  []string  `json:"affected_sites"`
	Recommendation string    `json:"recommendation"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConsistencyLevel represents the level of consistency
type ConsistencyLevel string

const (
	ConsistencyLevelHigh   ConsistencyLevel = "high"
	ConsistencyLevelMedium ConsistencyLevel = "medium"
	ConsistencyLevelLow    ConsistencyLevel = "low"
	ConsistencyLevelPoor   ConsistencyLevel = "poor"
)

// ConsistencyReport represents a comprehensive consistency report
type ConsistencyReport struct {
	BusinessID      string                      `json:"business_id"`
	GeneratedAt     time.Time                   `json:"generated_at"`
	Summary         ConsistencySummary          `json:"summary"`
	FieldAnalysis   []FieldAnalysis             `json:"field_analysis"`
	SiteAnalysis    []SiteAnalysis              `json:"site_analysis"`
	Recommendations []ConsistencyRecommendation `json:"recommendations"`
	Trends          []ConsistencyTrend          `json:"trends"`
}

// ConsistencySummary provides a high-level summary of consistency
type ConsistencySummary struct {
	OverallScore       float64          `json:"overall_score"`
	ConsistencyLevel   ConsistencyLevel `json:"consistency_level"`
	TotalFields        int              `json:"total_fields"`
	ConsistentFields   int              `json:"consistent_fields"`
	InconsistentFields int              `json:"inconsistent_fields"`
	TotalSites         int              `json:"total_sites"`
	HighQualitySites   int              `json:"high_quality_sites"`
	MediumQualitySites int              `json:"medium_quality_sites"`
	LowQualitySites    int              `json:"low_quality_sites"`
}

// FieldAnalysis provides detailed analysis for a specific field
type FieldAnalysis struct {
	FieldName        string   `json:"field_name"`
	DataType         string   `json:"data_type"`
	ConsistencyScore float64  `json:"consistency_score"`
	Coverage         float64  `json:"coverage"`
	QualityScore     float64  `json:"quality_score"`
	Trend            string   `json:"trend"`
	Issues           []string `json:"issues"`
}

// SiteAnalysis provides analysis for a specific site
type SiteAnalysis struct {
	SiteID           string   `json:"site_id"`
	URL              string   `json:"url"`
	ConsistencyScore float64  `json:"consistency_score"`
	DataQuality      float64  `json:"data_quality"`
	Coverage         float64  `json:"coverage"`
	Issues           []string `json:"issues"`
}

// ConsistencyRecommendation provides actionable recommendations
type ConsistencyRecommendation struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
	ActionItems []string `json:"action_items"`
}

// ConsistencyTrend represents consistency trends over time
type ConsistencyTrend struct {
	Period string  `json:"period"`
	Score  float64 `json:"score"`
	Trend  string  `json:"trend"`
	Change float64 `json:"change"`
}

// DataIntegrityResult represents the result of data integrity validation
type DataIntegrityResult struct {
	IsValid        bool      `json:"is_valid"`
	IntegrityScore float64   `json:"integrity_score"`
	Issues         []string  `json:"issues"`
	Warnings       []string  `json:"warnings"`
	ValidatedAt    time.Time `json:"validated_at"`
}

// SiteConsistencyValidator implements the ConsistencyValidator interface
type SiteConsistencyValidator struct {
	dataStore     DataStore
	locationStore LocationStore
	logger        *zap.Logger
	config        *ConsistencyValidationConfig
}

// ConsistencyValidationConfig holds configuration for consistency validation
type ConsistencyValidationConfig struct {
	MinConsistencyScore        float64       `json:"min_consistency_score"`
	HighConsistencyThreshold   float64       `json:"high_consistency_threshold"`
	MediumConsistencyThreshold float64       `json:"medium_consistency_threshold"`
	LowConsistencyThreshold    float64       `json:"low_consistency_threshold"`
	MaxFieldVariations         int           `json:"max_field_variations"`
	ValidationTimeout          time.Duration `json:"validation_timeout"`
	EnableTrendAnalysis        bool          `json:"enable_trend_analysis"`
	EnableRecommendations      bool          `json:"enable_recommendations"`
}

// DefaultConsistencyValidationConfig returns default configuration
func DefaultConsistencyValidationConfig() *ConsistencyValidationConfig {
	return &ConsistencyValidationConfig{
		MinConsistencyScore:        0.7,
		HighConsistencyThreshold:   0.9,
		MediumConsistencyThreshold: 0.7,
		LowConsistencyThreshold:    0.5,
		MaxFieldVariations:         5,
		ValidationTimeout:          30 * time.Second,
		EnableTrendAnalysis:        true,
		EnableRecommendations:      true,
	}
}

// NewSiteConsistencyValidator creates a new consistency validator
func NewSiteConsistencyValidator(
	dataStore DataStore,
	locationStore LocationStore,
	logger *zap.Logger,
	config *ConsistencyValidationConfig,
) *SiteConsistencyValidator {
	if config == nil {
		config = DefaultConsistencyValidationConfig()
	}

	return &SiteConsistencyValidator{
		dataStore:     dataStore,
		locationStore: locationStore,
		logger:        logger,
		config:        config,
	}
}

// ValidateConsistency validates data consistency across multiple sites for a business
func (v *SiteConsistencyValidator) ValidateConsistency(
	ctx context.Context,
	businessID string,
) (*ConsistencyValidationResult, error) {
	// Set timeout context
	ctx, cancel := context.WithTimeout(ctx, v.config.ValidationTimeout)
	defer cancel()

	v.logger.Info("starting consistency validation",
		zap.String("business_id", businessID))

	// Get all locations for the business
	locations, err := v.locationStore.GetLocationsByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no locations found for business %s", businessID)
	}

	// Get site data for all locations
	var allSiteData []*SiteData
	for _, location := range locations {
		siteDataList, err := v.dataStore.GetSiteDataByLocationID(ctx, location.ID)
		if err != nil {
			v.logger.Warn("failed to get site data for location",
				zap.String("location_id", location.ID),
				zap.Error(err))
			continue
		}
		for _, siteData := range siteDataList {
			allSiteData = append(allSiteData, &siteData)
		}
	}

	if len(allSiteData) == 0 {
		return nil, fmt.Errorf("no site data found for business %s", businessID)
	}

	// Validate consistency for each field
	fieldResults := v.validateFieldConsistency(allSiteData)

	// Calculate overall consistency score
	overallScore := v.calculateOverallConsistencyScore(fieldResults)

	// Determine consistency level
	consistencyLevel := v.determineConsistencyLevel(overallScore)

	// Generate issues and recommendations
	issues := v.generateConsistencyIssues(fieldResults)
	recommendations := v.generateRecommendations(fieldResults, overallScore)

	// Count sites
	consistentSites, inconsistentSites := v.countSiteConsistency(allSiteData, fieldResults)

	result := &ConsistencyValidationResult{
		BusinessID:        businessID,
		OverallScore:      overallScore,
		ConsistencyLevel:  consistencyLevel,
		FieldResults:      fieldResults,
		Issues:            issues,
		Recommendations:   recommendations,
		ValidatedAt:       time.Now(),
		TotalSites:        len(allSiteData),
		ConsistentSites:   consistentSites,
		InconsistentSites: inconsistentSites,
	}

	v.logger.Info("consistency validation completed",
		zap.String("business_id", businessID),
		zap.Float64("overall_score", overallScore),
		zap.String("consistency_level", string(consistencyLevel)))

	return result, nil
}

// ValidateFieldConsistency validates consistency for a specific field across sites
func (v *SiteConsistencyValidator) ValidateFieldConsistency(
	ctx context.Context,
	businessID string,
	fieldName string,
) (*FieldConsistencyResult, error) {
	// Get all site data for the business
	allSiteData, err := v.dataStore.GetSiteDataByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site data: %w", err)
	}

	if len(allSiteData) == 0 {
		return nil, fmt.Errorf("no site data found for business %s", businessID)
	}

	// Extract values for the specific field
	var fieldValues []FieldValue
	for _, siteData := range allSiteData {
		if value, exists := siteData.ExtractedData[fieldName]; exists {
			fieldValues = append(fieldValues, FieldValue{
				Value:       value,
				SiteID:      siteData.LocationID,
				URL:         v.getLocationURL(siteData.LocationID),
				Confidence:  siteData.ConfidenceScore,
				LastUpdated: siteData.LastExtracted,
			})
		}
	}

	if len(fieldValues) == 0 {
		return nil, fmt.Errorf("field %s not found in any site data", fieldName)
	}

	// Analyze field consistency
	consistencyScore := v.calculateFieldConsistencyScore(fieldValues)
	consistencyLevel := v.determineConsistencyLevel(consistencyScore)
	isConsistent := consistencyScore >= v.config.MinConsistencyScore

	// Generate field-specific issues
	issues := v.generateFieldConsistencyIssues(fieldName, fieldValues)

	// Generate recommendation
	recommendation := v.generateFieldRecommendation(fieldName, fieldValues, consistencyScore)

	result := &FieldConsistencyResult{
		FieldName:        fieldName,
		DataType:         v.determineFieldDataType(fieldValues),
		ConsistencyScore: consistencyScore,
		ConsistencyLevel: consistencyLevel,
		Values:           fieldValues,
		IsConsistent:     isConsistent,
		Issues:           issues,
		Recommendation:   recommendation,
	}

	return result, nil
}

// GetConsistencyReport generates a comprehensive consistency report
func (v *SiteConsistencyValidator) GetConsistencyReport(
	ctx context.Context,
	businessID string,
) (*ConsistencyReport, error) {
	// Validate consistency first
	validationResult, err := v.ValidateConsistency(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate consistency: %w", err)
	}

	// Get locations for site analysis
	locations, err := v.locationStore.GetLocationsByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}

	// Convert to pointer slice for helper methods
	var locationPtrs []*BusinessLocation
	for i := range locations {
		locationPtrs = append(locationPtrs, &locations[i])
	}

	// Generate summary
	summary := v.generateConsistencySummary(validationResult, locationPtrs)

	// Generate field analysis
	fieldAnalysis := v.generateFieldAnalysis(validationResult.FieldResults)

	// Generate site analysis
	siteAnalysis := v.generateSiteAnalysis(locationPtrs, validationResult.FieldResults)

	// Generate recommendations
	var recommendations []ConsistencyRecommendation
	if v.config.EnableRecommendations {
		recommendations = v.generateDetailedRecommendations(validationResult)
	}

	// Generate trends (if enabled)
	var trends []ConsistencyTrend
	if v.config.EnableTrendAnalysis {
		trends = v.generateConsistencyTrends(businessID)
	}

	report := &ConsistencyReport{
		BusinessID:      businessID,
		GeneratedAt:     time.Now(),
		Summary:         summary,
		FieldAnalysis:   fieldAnalysis,
		SiteAnalysis:    siteAnalysis,
		Recommendations: recommendations,
		Trends:          trends,
	}

	return report, nil
}

// ValidateDataIntegrity validates the integrity of aggregated data
func (v *SiteConsistencyValidator) ValidateDataIntegrity(
	ctx context.Context,
	aggregatedData *AggregatedBusinessData,
) (*DataIntegrityResult, error) {
	var issues []string
	var warnings []string

	// Check for required fields
	if aggregatedData.BusinessID == "" {
		issues = append(issues, "missing business ID")
	}

	// Check aggregated data structure
	if aggregatedData.AggregatedData == nil {
		issues = append(issues, "missing aggregated data")
	} else {
		// Validate each data type
		for dataType, data := range aggregatedData.AggregatedData {
			if data == nil {
				warnings = append(warnings, fmt.Sprintf("empty data for type: %s", dataType))
				continue
			}

			// Check for required fields based on data type
			switch dataType {
			case "contact_info":
				if dataMap, ok := data.(map[string]interface{}); ok {
					if _, hasPhone := dataMap["phone"]; !hasPhone {
						warnings = append(warnings, "contact_info missing phone number")
					}
					if _, hasEmail := dataMap["email"]; !hasEmail {
						warnings = append(warnings, "contact_info missing email")
					}
				}
			case "business_details":
				if dataMap, ok := data.(map[string]interface{}); ok {
					if _, hasName := dataMap["business_name"]; !hasName {
						issues = append(issues, "business_details missing business name")
					}
				}
			}
		}
	}

	// Check consistency issues
	if len(aggregatedData.ConsistencyIssues) > 0 {
		for _, issue := range aggregatedData.ConsistencyIssues {
			if issue.Severity == "high" {
				issues = append(issues, fmt.Sprintf("high severity consistency issue: %s", issue.Description))
			} else {
				warnings = append(warnings, fmt.Sprintf("consistency issue: %s", issue.Description))
			}
		}
	}

	// Calculate integrity score
	integrityScore := v.calculateIntegrityScore(issues, warnings)

	result := &DataIntegrityResult{
		IsValid:        len(issues) == 0,
		IntegrityScore: integrityScore,
		Issues:         issues,
		Warnings:       warnings,
		ValidatedAt:    time.Now(),
	}

	return result, nil
}

// Helper methods

func (v *SiteConsistencyValidator) validateFieldConsistency(sitesData []*SiteData) []FieldConsistencyResult {
	var results []FieldConsistencyResult

	// Get all unique field names across all sites
	fieldNames := v.getAllFieldNames(sitesData)

	for _, fieldName := range fieldNames {
		fieldResult := v.validateSingleField(sitesData, fieldName)
		results = append(results, fieldResult)
	}

	return results
}

func (v *SiteConsistencyValidator) getAllFieldNames(sitesData []*SiteData) []string {
	fieldMap := make(map[string]bool)

	for _, siteData := range sitesData {
		for fieldName := range siteData.ExtractedData {
			fieldMap[fieldName] = true
		}
	}

	var fieldNames []string
	for fieldName := range fieldMap {
		fieldNames = append(fieldNames, fieldName)
	}

	return fieldNames
}

func (v *SiteConsistencyValidator) validateSingleField(sitesData []*SiteData, fieldName string) FieldConsistencyResult {
	var fieldValues []FieldValue

	for _, siteData := range sitesData {
		if value, exists := siteData.ExtractedData[fieldName]; exists {
			fieldValues = append(fieldValues, FieldValue{
				Value:       value,
				SiteID:      siteData.LocationID,
				URL:         v.getLocationURL(siteData.LocationID),
				Confidence:  siteData.ConfidenceScore,
				LastUpdated: siteData.LastExtracted,
			})
		}
	}

	consistencyScore := v.calculateFieldConsistencyScore(fieldValues)
	consistencyLevel := v.determineConsistencyLevel(consistencyScore)
	isConsistent := consistencyScore >= v.config.MinConsistencyScore

	issues := v.generateFieldConsistencyIssues(fieldName, fieldValues)
	recommendation := v.generateFieldRecommendation(fieldName, fieldValues, consistencyScore)

	return FieldConsistencyResult{
		FieldName:        fieldName,
		DataType:         v.determineFieldDataType(fieldValues),
		ConsistencyScore: consistencyScore,
		ConsistencyLevel: consistencyLevel,
		Values:           fieldValues,
		IsConsistent:     isConsistent,
		Issues:           issues,
		Recommendation:   recommendation,
	}
}

func (v *SiteConsistencyValidator) calculateFieldConsistencyScore(fieldValues []FieldValue) float64 {
	if len(fieldValues) == 0 {
		return 0.0
	}

	if len(fieldValues) == 1 {
		return 1.0 // Perfect consistency for single value
	}

	// Group values by their string representation
	valueGroups := make(map[string][]FieldValue)
	for _, fv := range fieldValues {
		valueStr := v.valueToString(fv.Value)
		valueGroups[valueStr] = append(valueGroups[valueStr], fv)
	}

	// Calculate consistency based on value distribution
	totalValues := len(fieldValues)
	uniqueValues := len(valueGroups)

	// If all values are the same, perfect consistency
	if uniqueValues == 1 {
		return 1.0
	}

	// Calculate weighted consistency based on value frequency and confidence
	var totalWeightedScore float64
	var totalWeight float64

	for _, group := range valueGroups {
		// Calculate average confidence for this value group
		var avgConfidence float64
		for _, fv := range group {
			avgConfidence += fv.Confidence
		}
		avgConfidence /= float64(len(group))

		// Weight by frequency and confidence
		frequency := float64(len(group)) / float64(totalValues)
		weight := frequency * avgConfidence

		// Score based on frequency (higher frequency = higher score)
		score := frequency

		totalWeightedScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalWeightedScore / totalWeight
}

func (v *SiteConsistencyValidator) valueToString(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []string:
		return strings.Join(v, ",")
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (v *SiteConsistencyValidator) determineConsistencyLevel(score float64) ConsistencyLevel {
	switch {
	case score >= v.config.HighConsistencyThreshold:
		return ConsistencyLevelHigh
	case score >= v.config.MediumConsistencyThreshold:
		return ConsistencyLevelMedium
	case score >= v.config.LowConsistencyThreshold:
		return ConsistencyLevelLow
	default:
		return ConsistencyLevelPoor
	}
}

func (v *SiteConsistencyValidator) calculateOverallConsistencyScore(fieldResults []FieldConsistencyResult) float64 {
	if len(fieldResults) == 0 {
		return 0.0
	}

	var totalScore float64
	var totalWeight float64

	for _, result := range fieldResults {
		// Weight by number of values (more values = higher importance)
		weight := float64(len(result.Values))
		totalScore += result.ConsistencyScore * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

func (v *SiteConsistencyValidator) generateConsistencyIssues(fieldResults []FieldConsistencyResult) []ConsistencyIssue {
	var issues []ConsistencyIssue

	for _, fieldResult := range fieldResults {
		if !fieldResult.IsConsistent {
			issue := ConsistencyIssue{
				ID:             generateID(),
				Type:           "field_inconsistency",
				Severity:       v.determineIssueSeverity(fieldResult.ConsistencyScore),
				Description:    fmt.Sprintf("Inconsistent values for field '%s' across sites", fieldResult.FieldName),
				AffectedFields: []string{fieldResult.FieldName},
				AffectedSites:  v.getAffectedSiteIDs(fieldResult.Values),
				Recommendation: fieldResult.Recommendation,
				CreatedAt:      time.Now(),
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

func (v *SiteConsistencyValidator) generateRecommendations(fieldResults []FieldConsistencyResult, overallScore float64) []string {
	var recommendations []string

	// Overall recommendations
	if overallScore < v.config.MinConsistencyScore {
		recommendations = append(recommendations, "Overall consistency is below acceptable threshold. Review and standardize data across all sites.")
	}

	// Field-specific recommendations
	for _, fieldResult := range fieldResults {
		if !fieldResult.IsConsistent {
			recommendations = append(recommendations, fieldResult.Recommendation)
		}
	}

	// Remove duplicates
	return v.removeDuplicateStrings(recommendations)
}

func (v *SiteConsistencyValidator) countSiteConsistency(sitesData []*SiteData, fieldResults []FieldConsistencyResult) (consistent, inconsistent int) {
	siteConsistency := make(map[string]bool)

	// Initialize all sites as consistent
	for _, siteData := range sitesData {
		siteConsistency[siteData.LocationID] = true
	}

	// Mark sites as inconsistent if they have inconsistent fields
	for _, fieldResult := range fieldResults {
		if !fieldResult.IsConsistent {
			for _, value := range fieldResult.Values {
				siteConsistency[value.SiteID] = false
			}
		}
	}

	// Count consistent and inconsistent sites
	for _, isConsistent := range siteConsistency {
		if isConsistent {
			consistent++
		} else {
			inconsistent++
		}
	}

	return consistent, inconsistent
}

func (v *SiteConsistencyValidator) determineIssueSeverity(score float64) string {
	switch {
	case score < 0.3:
		return "high"
	case score < 0.6:
		return "medium"
	default:
		return "low"
	}
}

func (v *SiteConsistencyValidator) getAffectedSiteIDs(fieldValues []FieldValue) []string {
	siteMap := make(map[string]bool)
	for _, value := range fieldValues {
		siteMap[value.SiteID] = true
	}

	var siteIDs []string
	for siteID := range siteMap {
		siteIDs = append(siteIDs, siteID)
	}

	return siteIDs
}

func (v *SiteConsistencyValidator) removeDuplicateStrings(strings []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range strings {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

func (v *SiteConsistencyValidator) getLocationURL(locationID string) string {
	// In a real implementation, this would look up the location URL
	// For now, return a placeholder
	return fmt.Sprintf("https://example.com/location/%s", locationID)
}

func (v *SiteConsistencyValidator) determineFieldDataType(fieldValues []FieldValue) string {
	if len(fieldValues) == 0 {
		return "unknown"
	}

	// Determine data type based on the first non-nil value
	for _, fv := range fieldValues {
		if fv.Value != nil {
			return reflect.TypeOf(fv.Value).String()
		}
	}

	return "unknown"
}

func (v *SiteConsistencyValidator) generateFieldConsistencyIssues(fieldName string, fieldValues []FieldValue) []FieldConsistencyIssue {
	var issues []FieldConsistencyIssue

	if len(fieldValues) == 0 {
		return issues
	}

	// Check for too many variations
	uniqueValues := make(map[string]bool)
	for _, fv := range fieldValues {
		valueStr := v.valueToString(fv.Value)
		uniqueValues[valueStr] = true
	}

	if len(uniqueValues) > v.config.MaxFieldVariations {
		issue := FieldConsistencyIssue{
			IssueType:         "too_many_variations",
			Severity:          "medium",
			Description:       fmt.Sprintf("Field '%s' has too many variations (%d)", fieldName, len(uniqueValues)),
			AffectedSites:     v.getAffectedSiteIDs(fieldValues),
			ConflictingValues: v.getUniqueValues(fieldValues),
		}
		issues = append(issues, issue)
	}

	// Check for empty or null values
	var emptyValues []FieldValue
	for _, fv := range fieldValues {
		if fv.Value == nil || v.valueToString(fv.Value) == "" {
			emptyValues = append(emptyValues, fv)
		}
	}

	if len(emptyValues) > 0 {
		issue := FieldConsistencyIssue{
			IssueType:         "empty_values",
			Severity:          "low",
			Description:       fmt.Sprintf("Field '%s' has empty values in %d sites", fieldName, len(emptyValues)),
			AffectedSites:     v.getAffectedSiteIDs(emptyValues),
			ConflictingValues: []interface{}{nil, ""},
		}
		issues = append(issues, issue)
	}

	return issues
}

func (v *SiteConsistencyValidator) generateFieldRecommendation(fieldName string, fieldValues []FieldValue, consistencyScore float64) string {
	if consistencyScore >= v.config.MinConsistencyScore {
		return fmt.Sprintf("Field '%s' shows good consistency. No action required.", fieldName)
	}

	uniqueValues := make(map[string]int)
	for _, fv := range fieldValues {
		valueStr := v.valueToString(fv.Value)
		uniqueValues[valueStr]++
	}

	if len(uniqueValues) > v.config.MaxFieldVariations {
		return fmt.Sprintf("Field '%s' has too many variations. Consider standardizing values across sites.", fieldName)
	}

	return fmt.Sprintf("Field '%s' shows inconsistency. Review and standardize values across affected sites.", fieldName)
}

func (v *SiteConsistencyValidator) getUniqueValues(fieldValues []FieldValue) []interface{} {
	uniqueMap := make(map[string]interface{})
	for _, fv := range fieldValues {
		valueStr := v.valueToString(fv.Value)
		uniqueMap[valueStr] = fv.Value
	}

	var values []interface{}
	for _, value := range uniqueMap {
		values = append(values, value)
	}

	return values
}

func (v *SiteConsistencyValidator) calculateIntegrityScore(issues []string, warnings []string) float64 {
	totalChecks := len(issues) + len(warnings)
	if totalChecks == 0 {
		return 1.0
	}

	// Weight issues more heavily than warnings
	issueWeight := 2.0
	warningWeight := 1.0

	totalScore := float64(len(issues))*issueWeight + float64(len(warnings))*warningWeight
	maxPossibleScore := float64(totalChecks) * issueWeight

	return 1.0 - (totalScore / maxPossibleScore)
}

func (v *SiteConsistencyValidator) generateConsistencySummary(validationResult *ConsistencyValidationResult, locations []*BusinessLocation) ConsistencySummary {
	consistentFields := 0
	for _, fieldResult := range validationResult.FieldResults {
		if fieldResult.IsConsistent {
			consistentFields++
		}
	}

	// Count sites by quality level
	highQualitySites := 0
	mediumQualitySites := 0
	lowQualitySites := 0

	// This is a simplified quality assessment
	// In a real implementation, you'd calculate quality based on data completeness and accuracy
	highQualitySites = len(locations)

	return ConsistencySummary{
		OverallScore:       validationResult.OverallScore,
		ConsistencyLevel:   validationResult.ConsistencyLevel,
		TotalFields:        len(validationResult.FieldResults),
		ConsistentFields:   consistentFields,
		InconsistentFields: len(validationResult.FieldResults) - consistentFields,
		TotalSites:         len(locations),
		HighQualitySites:   highQualitySites,
		MediumQualitySites: mediumQualitySites,
		LowQualitySites:    lowQualitySites,
	}
}

func (v *SiteConsistencyValidator) generateFieldAnalysis(fieldResults []FieldConsistencyResult) []FieldAnalysis {
	var analysis []FieldAnalysis

	for _, result := range fieldResults {
		// Calculate coverage (percentage of sites that have this field)
		// For now, we'll use a fixed total sites count since it's not available in the result
		totalSites := len(result.Values) + 1 // Add 1 to avoid division by zero
		coverage := float64(len(result.Values)) / float64(totalSites) * 100

		// Calculate quality score (average confidence)
		var totalConfidence float64
		for _, value := range result.Values {
			totalConfidence += value.Confidence
		}
		qualityScore := totalConfidence / float64(len(result.Values))

		// Determine trend (simplified)
		trend := "stable"
		if result.ConsistencyScore < 0.5 {
			trend = "declining"
		} else if result.ConsistencyScore > 0.9 {
			trend = "improving"
		}

		// Generate issues list
		var issues []string
		for _, issue := range result.Issues {
			issues = append(issues, issue.Description)
		}

		analysis = append(analysis, FieldAnalysis{
			FieldName:        result.FieldName,
			DataType:         result.DataType,
			ConsistencyScore: result.ConsistencyScore,
			Coverage:         coverage,
			QualityScore:     qualityScore,
			Trend:            trend,
			Issues:           issues,
		})
	}

	return analysis
}

func (v *SiteConsistencyValidator) generateSiteAnalysis(locations []*BusinessLocation, fieldResults []FieldConsistencyResult) []SiteAnalysis {
	var analysis []SiteAnalysis

	for _, location := range locations {
		// Calculate site-specific metrics
		consistencyScore := v.calculateSiteConsistencyScore(location.ID, fieldResults)
		dataQuality := v.calculateSiteDataQuality(location.ID, fieldResults)
		coverage := v.calculateSiteCoverage(location.ID, fieldResults)

		// Generate issues for this site
		var issues []string
		for _, fieldResult := range fieldResults {
			for _, value := range fieldResult.Values {
				if value.SiteID == location.ID && !fieldResult.IsConsistent {
					issues = append(issues, fmt.Sprintf("Inconsistent field: %s", fieldResult.FieldName))
				}
			}
		}

		analysis = append(analysis, SiteAnalysis{
			SiteID:           location.ID,
			URL:              location.URL,
			ConsistencyScore: consistencyScore,
			DataQuality:      dataQuality,
			Coverage:         coverage,
			Issues:           issues,
		})
	}

	return analysis
}

func (v *SiteConsistencyValidator) generateDetailedRecommendations(validationResult *ConsistencyValidationResult) []ConsistencyRecommendation {
	var recommendations []ConsistencyRecommendation

	// Overall consistency recommendation
	if validationResult.OverallScore < v.config.MinConsistencyScore {
		recommendations = append(recommendations, ConsistencyRecommendation{
			ID:          generateID(),
			Type:        "overall_consistency",
			Priority:    "high",
			Description: "Overall data consistency is below acceptable threshold",
			Impact:      "high",
			Effort:      "medium",
			ActionItems: []string{
				"Review all inconsistent fields",
				"Standardize data formats across sites",
				"Implement data validation rules",
			},
		})
	}

	// Field-specific recommendations
	for _, fieldResult := range validationResult.FieldResults {
		if !fieldResult.IsConsistent {
			recommendations = append(recommendations, ConsistencyRecommendation{
				ID:          generateID(),
				Type:        "field_consistency",
				Priority:    v.determineRecommendationPriority(fieldResult.ConsistencyScore),
				Description: fmt.Sprintf("Inconsistent values for field '%s'", fieldResult.FieldName),
				Impact:      "medium",
				Effort:      "low",
				ActionItems: []string{
					fmt.Sprintf("Review values for field '%s'", fieldResult.FieldName),
					"Standardize field format across sites",
					"Update data extraction rules if needed",
				},
			})
		}
	}

	return recommendations
}

func (v *SiteConsistencyValidator) determineRecommendationPriority(score float64) string {
	switch {
	case score < 0.3:
		return "high"
	case score < 0.6:
		return "medium"
	default:
		return "low"
	}
}

func (v *SiteConsistencyValidator) generateConsistencyTrends(businessID string) []ConsistencyTrend {
	// This is a placeholder implementation
	// In a real implementation, you'd analyze historical consistency data
	return []ConsistencyTrend{
		{
			Period: "last_7_days",
			Score:  0.85,
			Trend:  "improving",
			Change: 0.05,
		},
		{
			Period: "last_30_days",
			Score:  0.82,
			Trend:  "stable",
			Change: 0.02,
		},
	}
}

func (v *SiteConsistencyValidator) calculateSiteConsistencyScore(siteID string, fieldResults []FieldConsistencyResult) float64 {
	var totalScore float64
	var totalFields int

	for _, fieldResult := range fieldResults {
		for _, value := range fieldResult.Values {
			if value.SiteID == siteID {
				totalScore += fieldResult.ConsistencyScore
				totalFields++
				break
			}
		}
	}

	if totalFields == 0 {
		return 0.0
	}

	return totalScore / float64(totalFields)
}

func (v *SiteConsistencyValidator) calculateSiteDataQuality(siteID string, fieldResults []FieldConsistencyResult) float64 {
	var totalConfidence float64
	var totalFields int

	for _, fieldResult := range fieldResults {
		for _, value := range fieldResult.Values {
			if value.SiteID == siteID {
				totalConfidence += value.Confidence
				totalFields++
			}
		}
	}

	if totalFields == 0 {
		return 0.0
	}

	return totalConfidence / float64(totalFields)
}

func (v *SiteConsistencyValidator) calculateSiteCoverage(siteID string, fieldResults []FieldConsistencyResult) float64 {
	var coveredFields int
	totalFields := len(fieldResults)

	for _, fieldResult := range fieldResults {
		for _, value := range fieldResult.Values {
			if value.SiteID == siteID {
				coveredFields++
				break
			}
		}
	}

	if totalFields == 0 {
		return 0.0
	}

	return float64(coveredFields) / float64(totalFields) * 100
}
