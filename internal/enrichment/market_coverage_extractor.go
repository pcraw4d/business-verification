package enrichment

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MarketCoverageExtractor extracts market coverage and service area information from website content
type MarketCoverageExtractor struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *MarketCoverageExtractorConfig
}

// MarketCoverageExtractorConfig contains configuration for market coverage extraction
type MarketCoverageExtractorConfig struct {
	// Service area patterns for different types of businesses
	ServiceAreaPatterns map[string][]string
	// Market coverage indicators
	MarketCoverageIndicators []string
	// Geographic scope patterns
	GeographicScopePatterns []string
	// Confidence thresholds
	MinConfidenceScore float64
	// Maximum service areas to extract
	MaxServiceAreas int
}

// ServiceArea represents a geographic service area
type ServiceArea struct {
	Type            string    `json:"type"`             // local, regional, national, international
	Name            string    `json:"name"`             // Area name (e.g., "New York Metro")
	Description     string    `json:"description"`      // Detailed description
	Countries       []string  `json:"countries"`        // Countries covered
	States          []string  `json:"states"`           // States/provinces covered
	Cities          []string  `json:"cities"`           // Cities covered
	Radius          *int      `json:"radius"`           // Service radius in miles/km
	RadiusUnit      string    `json:"radius_unit"`      // miles, km
	ConfidenceScore float64   `json:"confidence_score"` // Extraction accuracy
	ExtractedAt     time.Time `json:"extracted_at"`     // Timestamp
	Source          string    `json:"source"`           // Data source
}

// MarketCoverage represents market coverage information
type MarketCoverage struct {
	Type            string    `json:"type"`             // local, regional, national, international
	Description     string    `json:"description"`      // Market coverage description
	GeographicScope string    `json:"geographic_scope"` // Geographic scope
	TargetMarkets   []string  `json:"target_markets"`   // Target market segments
	ServiceAreas    []string  `json:"service_areas"`    // Service area names
	ConfidenceScore float64   `json:"confidence_score"` // Extraction accuracy
	ExtractedAt     time.Time `json:"extracted_at"`     // Timestamp
	Source          string    `json:"source"`           // Data source
}

// MarketCoverageResult contains the results of market coverage extraction
type MarketCoverageResult struct {
	ServiceAreas    []ServiceArea   `json:"service_areas"`    // All extracted service areas
	MarketCoverage  *MarketCoverage `json:"market_coverage"`  // Overall market coverage
	GeographicScope string          `json:"geographic_scope"` // Primary geographic scope
	TargetMarkets   []string        `json:"target_markets"`   // Target market segments
	CoverageType    string          `json:"coverage_type"`    // local, regional, national, international
	ConfidenceScore float64         `json:"confidence_score"` // Overall confidence score
	Evidence        []string        `json:"evidence"`         // Supporting evidence
	ProcessingTime  time.Duration   `json:"processing_time"`  // Time taken to process
}

// NewMarketCoverageExtractor creates a new market coverage extractor
func NewMarketCoverageExtractor(logger *zap.Logger, config *MarketCoverageExtractorConfig) *MarketCoverageExtractor {
	if config == nil {
		config = &MarketCoverageExtractorConfig{
			MinConfidenceScore: 0.3,
			MaxServiceAreas:    10,
			ServiceAreaPatterns: map[string][]string{
				"local": {
					`within\s+(\d+)\s*(mile|km|kilometer)s?`,
					`serving\s+([^,]+)`,
					`local\s+service\s+area`,
					`neighborhood\s+service`,
				},
				"regional": {
					`serving\s+([A-Za-z\s]+)\s+region`,
					`regional\s+coverage`,
					`throughout\s+([A-Za-z\s]+)`,
					`across\s+([A-Za-z\s]+)`,
				},
				"national": {
					`nationwide\s+service`,
					`serving\s+all\s+50\s+states`,
					`across\s+the\s+country`,
					`national\s+coverage`,
				},
				"international": {
					`international\s+service`,
					`serving\s+(\d+)\s+countries`,
					`global\s+coverage`,
					`worldwide\s+service`,
				},
			},
			MarketCoverageIndicators: []string{
				"service area", "coverage area", "serving", "available in", "operating in",
				"locations", "regions", "markets", "territories", "jurisdictions",
			},
			GeographicScopePatterns: []string{
				`local`, `regional`, `national`, `international`, `global`,
				`worldwide`, `domestic`, `overseas`, `cross-border`,
			},
		}
	}

	return &MarketCoverageExtractor{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("market_coverage_extractor"),
		config: config,
	}
}

// ExtractMarketCoverage extracts market coverage and service areas from website content
func (mce *MarketCoverageExtractor) ExtractMarketCoverage(ctx context.Context, content string) (*MarketCoverageResult, error) {
	ctx, span := mce.tracer.Start(ctx, "ExtractMarketCoverage")
	defer span.End()

	startTime := time.Now()
	mce.logger.Info("Starting market coverage extraction", zap.String("content_length", fmt.Sprintf("%d", len(content))))

	result := &MarketCoverageResult{
		ServiceAreas:    []ServiceArea{},
		TargetMarkets:   []string{},
		Evidence:        []string{},
		ProcessingTime:  0,
		ConfidenceScore: 0.0,
	}

	// Extract service areas
	serviceAreas, err := mce.extractServiceAreas(ctx, content)
	if err != nil {
		mce.logger.Error("Failed to extract service areas", zap.Error(err))
		return nil, fmt.Errorf("service area extraction failed: %w", err)
	}
	result.ServiceAreas = serviceAreas

	// Extract market coverage
	marketCoverage, err := mce.extractMarketCoverage(ctx, content)
	if err != nil {
		mce.logger.Error("Failed to extract market coverage", zap.Error(err))
		return nil, fmt.Errorf("market coverage extraction failed: %w", err)
	}
	result.MarketCoverage = marketCoverage

	// Determine geographic scope
	result.GeographicScope = mce.determineGeographicScope(serviceAreas, marketCoverage)

	// Extract target markets
	result.TargetMarkets = mce.extractTargetMarkets(content)

	// Determine coverage type
	result.CoverageType = mce.determineCoverageType(serviceAreas, marketCoverage)

	// Calculate overall confidence
	result.ConfidenceScore = mce.calculateOverallConfidence(result)

	// Add evidence
	result.Evidence = mce.collectEvidence(content, serviceAreas, marketCoverage)

	result.ProcessingTime = time.Since(startTime)

	mce.logger.Info("Market coverage extraction completed",
		zap.Int("service_areas", len(result.ServiceAreas)),
		zap.String("coverage_type", result.CoverageType),
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// extractServiceAreas extracts service areas from content
func (mce *MarketCoverageExtractor) extractServiceAreas(ctx context.Context, content string) ([]ServiceArea, error) {
	var serviceAreas []ServiceArea

	// Extract service areas by type
	for areaType, patterns := range mce.config.ServiceAreaPatterns {
		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				mce.logger.Warn("Invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
				continue
			}

			matches := re.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					serviceArea := ServiceArea{
						Type:        areaType,
						Name:        mce.extractServiceAreaName(match, areaType),
						Description: mce.extractServiceAreaDescription(match, areaType),
						ExtractedAt: time.Now(),
						Source:      "pattern_matching",
					}

					// Extract radius if present
					if radius := mce.extractRadius(match); radius != nil {
						serviceArea.Radius = radius
						serviceArea.RadiusUnit = mce.extractRadiusUnit(match)
					}

					// Extract geographic information
					serviceArea.Countries = mce.extractCountries(match[0])
					serviceArea.States = mce.extractStates(match[0])
					serviceArea.Cities = mce.extractCities(match[0])

					// Calculate confidence
					serviceArea.ConfidenceScore = mce.calculateServiceAreaConfidence(serviceArea, match)

					serviceAreas = append(serviceAreas, serviceArea)
				}
			}
		}
	}

	// Extract service areas from market coverage indicators
	indicatorAreas := mce.extractFromIndicators(content)
	serviceAreas = append(serviceAreas, indicatorAreas...)

	// Deduplicate and validate
	serviceAreas = mce.deduplicateServiceAreas(serviceAreas)
	serviceAreas = mce.validateServiceAreas(serviceAreas)

	// Limit results
	if len(serviceAreas) > mce.config.MaxServiceAreas {
		serviceAreas = serviceAreas[:mce.config.MaxServiceAreas]
	}

	return serviceAreas, nil
}

// extractMarketCoverage extracts overall market coverage information
func (mce *MarketCoverageExtractor) extractMarketCoverage(ctx context.Context, content string) (*MarketCoverage, error) {
	marketCoverage := &MarketCoverage{
		ExtractedAt: time.Now(),
		Source:      "content_analysis",
	}

	// Determine coverage type
	marketCoverage.Type = mce.determineMarketCoverageType(content)

	// Extract description
	marketCoverage.Description = mce.extractMarketCoverageDescription(content)

	// Extract geographic scope
	marketCoverage.GeographicScope = mce.extractGeographicScope(content)

	// Extract target markets
	marketCoverage.TargetMarkets = mce.extractTargetMarkets(content)

	// Extract service areas
	marketCoverage.ServiceAreas = mce.extractServiceAreaNames(content)

	// Calculate confidence
	marketCoverage.ConfidenceScore = mce.calculateMarketCoverageConfidence(marketCoverage, content)

	return marketCoverage, nil
}

// extractServiceAreaName extracts the name of a service area from regex matches
func (mce *MarketCoverageExtractor) extractServiceAreaName(match []string, areaType string) string {
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return areaType + " service area"
}

// extractServiceAreaDescription extracts a description for a service area
func (mce *MarketCoverageExtractor) extractServiceAreaDescription(match []string, areaType string) string {
	if len(match) > 1 {
		return fmt.Sprintf("Service area covering %s", strings.TrimSpace(match[1]))
	}
	return fmt.Sprintf("%s service area", strings.Title(areaType))
}

// extractRadius extracts radius information from matches
func (mce *MarketCoverageExtractor) extractRadius(match []string) *int {
	if len(match) > 0 {
		// Look for numeric radius in the match
		radiusPattern := regexp.MustCompile(`(\d+)`)
		if radiusMatch := radiusPattern.FindString(match[0]); radiusMatch != "" {
			var radius int
			if _, err := fmt.Sscanf(radiusMatch, "%d", &radius); err == nil {
				return &radius
			}
		}
	}
	return nil
}

// extractRadiusUnit extracts the unit for radius (miles, km)
func (mce *MarketCoverageExtractor) extractRadiusUnit(match []string) string {
	if len(match) > 0 {
		if strings.Contains(strings.ToLower(match[0]), "km") || strings.Contains(strings.ToLower(match[0]), "kilometer") {
			return "km"
		}
		if strings.Contains(strings.ToLower(match[0]), "mile") {
			return "miles"
		}
	}
	return "miles" // Default
}

// extractCountries extracts countries from text
func (mce *MarketCoverageExtractor) extractCountries(text string) []string {
	// Simple country extraction - can be enhanced with more sophisticated logic
	countries := []string{}
	countryPatterns := map[string]string{
		"united states":  "us",
		"usa":            "us",
		"canada":         "ca",
		"united kingdom": "uk",
		"australia":      "au",
		"germany":        "de",
		"france":         "fr",
		"japan":          "jp",
		"china":          "cn",
		"india":          "in",
		"brazil":         "br",
	}

	textLower := strings.ToLower(text)
	for countryName, countryCode := range countryPatterns {
		if strings.Contains(textLower, countryName) {
			countries = append(countries, countryCode)
		}
	}

	return countries
}

// extractStates extracts states/provinces from text
func (mce *MarketCoverageExtractor) extractStates(text string) []string {
	// Simple state extraction - can be enhanced
	states := []string{}
	statePattern := regexp.MustCompile(`\b([A-Z]{2})\b`)
	matches := statePattern.FindAllString(text, -1)

	for _, match := range matches {
		if match != "US" && match != "UK" && match != "CA" {
			states = append(states, match)
		}
	}

	return states
}

// extractCities extracts cities from text
func (mce *MarketCoverageExtractor) extractCities(text string) []string {
	// Simple city extraction - can be enhanced with more sophisticated logic
	cities := []string{}
	cityPattern := regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\b`)
	matches := cityPattern.FindAllString(text, -1)

	for _, match := range matches {
		if len(match) > 3 && !strings.Contains(strings.ToLower(match), "service") {
			cities = append(cities, match)
		}
	}

	return cities
}

// extractFromIndicators extracts service areas from market coverage indicators
func (mce *MarketCoverageExtractor) extractFromIndicators(content string) []ServiceArea {
	var serviceAreas []ServiceArea

	for _, indicator := range mce.config.MarketCoverageIndicators {
		if strings.Contains(strings.ToLower(content), strings.ToLower(indicator)) {
			serviceArea := ServiceArea{
				Type:            "general",
				Name:            indicator + " area",
				Description:     fmt.Sprintf("Service area identified from '%s' indicator", indicator),
				ExtractedAt:     time.Now(),
				Source:          "indicator_extraction",
				ConfidenceScore: 0.4,
			}

			serviceAreas = append(serviceAreas, serviceArea)
		}
	}

	return serviceAreas
}

// extractContext extracts context around a keyword
func (mce *MarketCoverageExtractor) extractContext(content, keyword string, contextSize int) string {
	index := strings.Index(strings.ToLower(content), strings.ToLower(keyword))
	if index == -1 {
		return ""
	}

	start := max(0, index-contextSize)
	end := min(len(content), index+len(keyword)+contextSize)

	return content[start:end]
}

// determineMarketCoverageType determines the type of market coverage
func (mce *MarketCoverageExtractor) determineMarketCoverageType(content string) string {
	contentLower := strings.ToLower(content)

	if strings.Contains(contentLower, "global") || strings.Contains(contentLower, "worldwide") {
		return "international"
	}
	if strings.Contains(contentLower, "national") || strings.Contains(contentLower, "nationwide") {
		return "national"
	}
	if strings.Contains(contentLower, "regional") {
		return "regional"
	}
	if strings.Contains(contentLower, "local") {
		return "local"
	}

	return "unknown"
}

// extractMarketCoverageDescription extracts market coverage description
func (mce *MarketCoverageExtractor) extractMarketCoverageDescription(content string) string {
	// Look for sentences containing coverage indicators
	coveragePattern := regexp.MustCompile(`[^.]*(?:service area|coverage area|serving|available in)[^.]*\.`)
	if match := coveragePattern.FindString(content); match != "" {
		return strings.TrimSpace(match)
	}

	return "Market coverage information extracted from website content"
}

// extractGeographicScope extracts geographic scope from content
func (mce *MarketCoverageExtractor) extractGeographicScope(content string) string {
	contentLower := strings.ToLower(content)

	for _, pattern := range mce.config.GeographicScopePatterns {
		if strings.Contains(contentLower, pattern) {
			return pattern
		}
	}

	return "unknown"
}

// extractTargetMarkets extracts target market segments
func (mce *MarketCoverageExtractor) extractTargetMarkets(content string) []string {
	targetMarkets := []string{}

	// Common target market patterns
	marketPatterns := []string{
		`small business`, `enterprise`, `startup`, `sme`, `mid-market`,
		`consumer`, `b2b`, `b2c`, `government`, `healthcare`, `education`,
		`retail`, `manufacturing`, `technology`, `financial`,
	}

	contentLower := strings.ToLower(content)
	for _, pattern := range marketPatterns {
		if strings.Contains(contentLower, pattern) {
			targetMarkets = append(targetMarkets, pattern)
		}
	}

	return targetMarkets
}

// extractServiceAreaNames extracts service area names from content
func (mce *MarketCoverageExtractor) extractServiceAreaNames(content string) []string {
	serviceAreas := []string{}

	// Look for geographic names that might be service areas
	geoPattern := regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\b`)
	matches := geoPattern.FindAllString(content, -1)

	for _, match := range matches {
		if len(match) > 3 && !strings.Contains(strings.ToLower(match), "service") {
			serviceAreas = append(serviceAreas, match)
		}
	}

	return serviceAreas
}

// determineGeographicScope determines the primary geographic scope
func (mce *MarketCoverageExtractor) determineGeographicScope(serviceAreas []ServiceArea, marketCoverage *MarketCoverage) string {
	if marketCoverage != nil && marketCoverage.GeographicScope != "unknown" {
		return marketCoverage.GeographicScope
	}

	// Determine from service areas
	scopeCounts := make(map[string]int)
	for _, area := range serviceAreas {
		scopeCounts[area.Type]++
	}

	// Return the most common scope
	maxCount := 0
	primaryScope := "unknown"
	for scope, count := range scopeCounts {
		if count > maxCount {
			maxCount = count
			primaryScope = scope
		}
	}

	return primaryScope
}

// determineCoverageType determines the overall coverage type
func (mce *MarketCoverageExtractor) determineCoverageType(serviceAreas []ServiceArea, marketCoverage *MarketCoverage) string {
	if marketCoverage != nil && marketCoverage.Type != "unknown" {
		return marketCoverage.Type
	}

	// If no service areas, return unknown
	if len(serviceAreas) == 0 {
		return "unknown"
	}

	// Determine from service areas
	coverageCounts := make(map[string]int)
	for _, area := range serviceAreas {
		coverageCounts[area.Type]++
	}

	// Return the most common coverage type
	maxCount := 0
	primaryCoverage := "unknown"
	for coverage, count := range coverageCounts {
		if count > maxCount {
			maxCount = count
			primaryCoverage = coverage
		}
	}

	return primaryCoverage
}

// calculateServiceAreaConfidence calculates confidence for a service area
func (mce *MarketCoverageExtractor) calculateServiceAreaConfidence(area ServiceArea, match []string) float64 {
	confidence := 0.5 // Base confidence

	// Boost confidence based on match quality
	if len(match) > 1 {
		confidence += 0.2
	}

	// Boost confidence based on area completeness
	if area.Name != "" {
		confidence += 0.1
	}
	if area.Description != "" {
		confidence += 0.1
	}
	if len(area.Countries) > 0 || len(area.States) > 0 || len(area.Cities) > 0 {
		confidence += 0.1
	}

	return minFloat64(confidence, 1.0)
}

// calculateMarketCoverageConfidence calculates confidence for market coverage
func (mce *MarketCoverageExtractor) calculateMarketCoverageConfidence(coverage *MarketCoverage, content string) float64 {
	confidence := 0.3 // Base confidence

	// Boost confidence based on coverage completeness
	if coverage.Type != "unknown" {
		confidence += 0.2
	}
	if coverage.Description != "" {
		confidence += 0.2
	}
	if coverage.GeographicScope != "unknown" {
		confidence += 0.1
	}
	if len(coverage.TargetMarkets) > 0 {
		confidence += 0.1
	}
	if len(coverage.ServiceAreas) > 0 {
		confidence += 0.1
	}

	return minFloat64(confidence, 1.0)
}

// calculateOverallConfidence calculates overall confidence for the result
func (mce *MarketCoverageExtractor) calculateOverallConfidence(result *MarketCoverageResult) float64 {
	if len(result.ServiceAreas) == 0 && result.MarketCoverage == nil {
		return 0.0
	}

	totalConfidence := 0.0
	count := 0

	// Average confidence from service areas
	for _, area := range result.ServiceAreas {
		totalConfidence += area.ConfidenceScore
		count++
	}

	// Add market coverage confidence
	if result.MarketCoverage != nil {
		totalConfidence += result.MarketCoverage.ConfidenceScore
		count++
	}

	if count == 0 {
		return 0.0
	}

	avgConfidence := totalConfidence / float64(count)

	// Boost confidence based on result completeness
	completenessBonus := 0.0
	if len(result.ServiceAreas) > 0 {
		completenessBonus += 0.1
	}
	if result.MarketCoverage != nil {
		completenessBonus += 0.1
	}
	if result.GeographicScope != "unknown" {
		completenessBonus += 0.1
	}
	if len(result.TargetMarkets) > 0 {
		completenessBonus += 0.1
	}

	// For empty content, return 0 confidence
	if result.GeographicScope == "unknown" && len(result.ServiceAreas) == 0 {
		return 0.0
	}

	return minFloat64(avgConfidence+completenessBonus, 1.0)
}

// collectEvidence collects supporting evidence for the extraction
func (mce *MarketCoverageExtractor) collectEvidence(content string, serviceAreas []ServiceArea, marketCoverage *MarketCoverage) []string {
	evidence := []string{}

	// Add evidence from service areas
	for _, area := range serviceAreas {
		if area.Name != "" {
			evidence = append(evidence, fmt.Sprintf("Service area: %s", area.Name))
		}
	}

	// Add evidence from market coverage
	if marketCoverage != nil && marketCoverage.Description != "" {
		evidence = append(evidence, fmt.Sprintf("Market coverage: %s", marketCoverage.Description))
	}

	// Add evidence from indicators found
	for _, indicator := range mce.config.MarketCoverageIndicators {
		if strings.Contains(strings.ToLower(content), strings.ToLower(indicator)) {
			evidence = append(evidence, fmt.Sprintf("Found indicator: %s", indicator))
		}
	}

	return evidence
}

// deduplicateServiceAreas removes duplicate service areas
func (mce *MarketCoverageExtractor) deduplicateServiceAreas(areas []ServiceArea) []ServiceArea {
	seen := make(map[string]bool)
	var unique []ServiceArea

	for _, area := range areas {
		key := area.Type + ":" + area.Name
		if !seen[key] {
			seen[key] = true
			unique = append(unique, area)
		}
	}

	return unique
}

// validateServiceAreas validates and filters service areas
func (mce *MarketCoverageExtractor) validateServiceAreas(areas []ServiceArea) []ServiceArea {
	var valid []ServiceArea

	for _, area := range areas {
		if area.ConfidenceScore >= mce.config.MinConfidenceScore {
			valid = append(valid, area)
		}
	}

	return valid
}
