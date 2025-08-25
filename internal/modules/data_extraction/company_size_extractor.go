package data_extraction

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CompanySizeExtractor extracts company size information from business data
type CompanySizeExtractor struct {
	// Configuration
	config *CompanySizeConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Pattern matching
	employeePatterns   []*regexp.Regexp
	revenuePatterns    []*regexp.Regexp
	locationPatterns   []*regexp.Regexp
	teamSizePatterns   []*regexp.Regexp
	startupPatterns    []*regexp.Regexp
	enterprisePatterns []*regexp.Regexp
}

// CompanySizeConfig holds configuration for the company size extractor
type CompanySizeConfig struct {
	// Pattern matching settings
	CaseSensitive bool
	MaxPatterns   int

	// Confidence scoring settings
	MinConfidenceThreshold float64
	MaxConfidenceThreshold float64

	// Validation settings
	MaxEmployeeCount int
	MaxRevenueAmount int64
	MaxLocations     int

	// Processing settings
	Timeout time.Duration
}

// CompanySize represents extracted company size information
type CompanySize struct {
	// Employee information
	EmployeeCountRange string  `json:"employee_count_range"`
	EmployeeCountMin   int     `json:"employee_count_min"`
	EmployeeCountMax   int     `json:"employee_count_max"`
	EmployeeConfidence float64 `json:"employee_confidence"`

	// Revenue information
	RevenueIndicator  string  `json:"revenue_indicator"`
	RevenueRange      string  `json:"revenue_range"`
	RevenueConfidence float64 `json:"revenue_confidence"`

	// Location information
	OfficeLocationsCount int     `json:"office_locations_count"`
	LocationsConfidence  float64 `json:"locations_confidence"`

	// Team information
	TeamSizeIndicator  string  `json:"team_size_indicator"`
	TeamSizeConfidence float64 `json:"team_size_confidence"`

	// Overall assessment
	CompanySizeCategory string  `json:"company_size_category"` // startup, small, medium, large, enterprise
	OverallConfidence   float64 `json:"overall_confidence"`

	// Metadata
	ExtractedAt time.Time `json:"extracted_at"`
	DataSources []string  `json:"data_sources"`
}

// EmployeeCountRanges defines standard employee count ranges
var EmployeeCountRanges = map[string]struct {
	Min int
	Max int
}{
	"1-10":     {1, 10},
	"11-50":    {11, 50},
	"51-200":   {51, 200},
	"201-500":  {201, 500},
	"501-1000": {501, 1000},
	"1000+":    {1001, 999999},
}

// RevenueIndicators defines revenue-based company size indicators
var RevenueIndicators = map[string]string{
	"startup":    "startup",
	"small":      "small_business",
	"medium":     "medium_business",
	"large":      "large_business",
	"enterprise": "enterprise",
}

// NewCompanySizeExtractor creates a new company size extractor
func NewCompanySizeExtractor(
	config *CompanySizeConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *CompanySizeExtractor {
	// Set default configuration
	if config == nil {
		config = &CompanySizeConfig{
			CaseSensitive:          false,
			MaxPatterns:            100,
			MinConfidenceThreshold: 0.3,
			MaxConfidenceThreshold: 1.0,
			MaxEmployeeCount:       100000,
			MaxRevenueAmount:       1000000000000, // 1 trillion
			MaxLocations:           1000,
			Timeout:                30 * time.Second,
		}
	}

	extractor := &CompanySizeExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize pattern matching
	extractor.initializePatterns()

	return extractor
}

// initializePatterns initializes all pattern matching regexes
func (cse *CompanySizeExtractor) initializePatterns() {
	// Employee count patterns
	cse.employeePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\d+)\s*(?:employees?|staff|team members?)`),
		regexp.MustCompile(`(?i)(\d+)\s*(?:people|workers?)`),
		regexp.MustCompile(`(?i)team\s+of\s+(\d+)`),
		regexp.MustCompile(`(?i)employing\s+(\d+)`),
		regexp.MustCompile(`(?i)workforce\s+of\s+(\d+)`),
		regexp.MustCompile(`(?i)company\s+size[:\s]*(\d+)`),
		regexp.MustCompile(`(?i)headcount[:\s]*(\d+)`),
		regexp.MustCompile(`(?i)staff\s+size[:\s]*(\d+)`),
		regexp.MustCompile(`(?i)employee\s+count[:\s]*(\d+)`),
		regexp.MustCompile(`(?i)team\s+size[:\s]*(\d+)`),
	}

	// Revenue patterns
	cse.revenuePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(?:million|mil|m)\s*(?:dollars?|usd|revenue)`),
		regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(?:billion|bil|b)\s*(?:dollars?|usd|revenue)`),
		regexp.MustCompile(`(?i)revenue[:\s]*\$?(\d+(?:,\d{3})*(?:\.\d+)?)`),
		regexp.MustCompile(`(?i)annual\s+revenue[:\s]*\$?(\d+(?:,\d{3})*(?:\.\d+)?)`),
		regexp.MustCompile(`(?i)revenue\s+of\s+\$?(\d+(?:,\d{3})*(?:\.\d+)?)`),
		regexp.MustCompile(`(?i)generates\s+\$?(\d+(?:,\d{3})*(?:\.\d+)?)`),
		regexp.MustCompile(`(?i)earns\s+\$?(\d+(?:,\d{3})*(?:\.\d+)?)`),
	}

	// Location patterns
	cse.locationPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\d+)\s*(?:offices?|locations?|branches?)`),
		regexp.MustCompile(`(?i)offices?\s+in\s+(\d+)\s+cities?`),
		regexp.MustCompile(`(?i)locations?\s+in\s+(\d+)\s+cities?`),
		regexp.MustCompile(`(?i)presence\s+in\s+(\d+)\s+cities?`),
		regexp.MustCompile(`(?i)operating\s+in\s+(\d+)\s+cities?`),
		regexp.MustCompile(`(?i)global\s+presence[:\s]*(\d+)\s+locations?`),
		regexp.MustCompile(`(?i)worldwide\s+with\s+(\d+)\s+offices?`),
	}

	// Team size patterns
	cse.teamSizePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)small\s+team`),
		regexp.MustCompile(`(?i)startup\s+team`),
		regexp.MustCompile(`(?i)lean\s+team`),
		regexp.MustCompile(`(?i)core\s+team`),
		regexp.MustCompile(`(?i)dedicated\s+team`),
		regexp.MustCompile(`(?i)large\s+team`),
		regexp.MustCompile(`(?i)global\s+team`),
		regexp.MustCompile(`(?i)distributed\s+team`),
		regexp.MustCompile(`(?i)remote\s+team`),
		regexp.MustCompile(`(?i)cross-functional\s+team`),
	}

	// Startup patterns
	cse.startupPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)startup`),
		regexp.MustCompile(`(?i)early-stage`),
		regexp.MustCompile(`(?i)seed-stage`),
		regexp.MustCompile(`(?i)series\s+a`),
		regexp.MustCompile(`(?i)funded\s+startup`),
		regexp.MustCompile(`(?i)tech\s+startup`),
		regexp.MustCompile(`(?i)innovative\s+startup`),
		regexp.MustCompile(`(?i)disruptive\s+startup`),
	}

	// Enterprise patterns
	cse.enterprisePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)enterprise`),
		regexp.MustCompile(`(?i)fortune\s+500`),
		regexp.MustCompile(`(?i)global\s+corporation`),
		regexp.MustCompile(`(?i)multinational`),
		regexp.MustCompile(`(?i)large\s+corporation`),
		regexp.MustCompile(`(?i)established\s+company`),
		regexp.MustCompile(`(?i)industry\s+leader`),
		regexp.MustCompile(`(?i)market\s+leader`),
	}
}

// ExtractCompanySize extracts company size information from business data
func (cse *CompanySizeExtractor) ExtractCompanySize(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
) (*CompanySize, error) {
	ctx, span := cse.tracer.Start(ctx, "CompanySizeExtractor.ExtractCompanySize")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessData.BusinessName),
		attribute.String("website", businessData.WebsiteURL),
	)

	// Create result structure
	result := &CompanySize{
		ExtractedAt: time.Now(),
		DataSources: []string{"text_analysis", "pattern_matching"},
	}

	// Extract employee count information
	if err := cse.extractEmployeeCount(ctx, businessData, result); err != nil {
		cse.logger.Warn("failed to extract employee count", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract revenue information
	if err := cse.extractRevenueInfo(ctx, businessData, result); err != nil {
		cse.logger.Warn("failed to extract revenue info", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract location information
	if err := cse.extractLocationInfo(ctx, businessData, result); err != nil {
		cse.logger.Warn("failed to extract location info", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract team size information
	if err := cse.extractTeamSizeInfo(ctx, businessData, result); err != nil {
		cse.logger.Warn("failed to extract team size info", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Determine overall company size category
	cse.determineCompanySizeCategory(result)

	// Calculate overall confidence
	cse.calculateOverallConfidence(result)

	// Validate results
	if err := cse.validateResults(result); err != nil {
		cse.logger.Warn("company size validation failed", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	cse.logger.Info("company size extraction completed", map[string]interface{}{
		"business_name":          businessData.BusinessName,
		"employee_count_range":   result.EmployeeCountRange,
		"revenue_indicator":      result.RevenueIndicator,
		"office_locations_count": result.OfficeLocationsCount,
		"company_size_category":  result.CompanySizeCategory,
		"overall_confidence":     result.OverallConfidence,
	})

	return result, nil
}

// extractEmployeeCount extracts employee count information
func (cse *CompanySizeExtractor) extractEmployeeCount(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *CompanySize,
) error {
	ctx, span := cse.tracer.Start(ctx, "CompanySizeExtractor.extractEmployeeCount")
	defer span.End()

	// Combine all text for analysis
	text := cse.combineText(businessData)

	// Find employee count patterns
	var employeeCount int
	var confidence float64
	var found bool

	for _, pattern := range cse.employeePatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				if count > 0 && count <= cse.config.MaxEmployeeCount {
					employeeCount = count
					confidence = 0.8 // High confidence for explicit mentions
					found = true
					break
				}
			}
		}
	}

	// If no explicit count found, try to infer from context
	if !found {
		employeeCount, confidence = cse.inferEmployeeCount(text)
	}

	// Set employee count range
	if employeeCount > 0 {
		result.EmployeeCountMin = employeeCount
		result.EmployeeCountMax = employeeCount
		result.EmployeeCountRange = cse.getEmployeeCountRange(employeeCount)
		result.EmployeeConfidence = confidence
	}

	span.SetAttributes(
		attribute.Int("employee_count", employeeCount),
		attribute.Float64("confidence", confidence),
	)

	return nil
}

// extractRevenueInfo extracts revenue information
func (cse *CompanySizeExtractor) extractRevenueInfo(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *CompanySize,
) error {
	ctx, span := cse.tracer.Start(ctx, "CompanySizeExtractor.extractRevenueInfo")
	defer span.End()

	// Combine all text for analysis
	text := cse.combineText(businessData)

	// Find revenue patterns
	var revenue float64
	var confidence float64
	var found bool

	for _, pattern := range cse.revenuePatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			if rev, err := strconv.ParseFloat(matches[1], 64); err == nil {
				revenue = rev
				confidence = 0.7 // Medium-high confidence for explicit mentions
				found = true
				break
			}
		}
	}

	// If no explicit revenue found, try to infer from context
	if !found {
		result.RevenueIndicator, confidence = cse.inferRevenueIndicator(text)
	} else {
		result.RevenueIndicator = cse.getRevenueIndicator(revenue)
		result.RevenueRange = cse.getRevenueRange(revenue)
		result.RevenueConfidence = confidence
	}

	span.SetAttributes(
		attribute.Float64("revenue", revenue),
		attribute.String("revenue_indicator", result.RevenueIndicator),
		attribute.Float64("confidence", confidence),
	)

	return nil
}

// extractLocationInfo extracts location information
func (cse *CompanySizeExtractor) extractLocationInfo(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *CompanySize,
) error {
	ctx, span := cse.tracer.Start(ctx, "CompanySizeExtractor.extractLocationInfo")
	defer span.End()

	// Combine all text for analysis
	text := cse.combineText(businessData)

	// Find location patterns
	var locationCount int
	var confidence float64
	var found bool

	for _, pattern := range cse.locationPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				if count > 0 && count <= cse.config.MaxLocations {
					locationCount = count
					confidence = 0.8 // High confidence for explicit mentions
					found = true
					break
				}
			}
		}
	}

	// If no explicit count found, try to infer from context
	if !found {
		locationCount, confidence = cse.inferLocationCount(text)
	}

	// Set location information
	if locationCount > 0 {
		result.OfficeLocationsCount = locationCount
		result.LocationsConfidence = confidence
	}

	span.SetAttributes(
		attribute.Int("location_count", locationCount),
		attribute.Float64("confidence", confidence),
	)

	return nil
}

// extractTeamSizeInfo extracts team size information
func (cse *CompanySizeExtractor) extractTeamSizeInfo(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *CompanySize,
) error {
	ctx, span := cse.tracer.Start(ctx, "CompanySizeExtractor.extractTeamSizeInfo")
	defer span.End()

	// Combine all text for analysis
	text := cse.combineText(businessData)

	// Find team size patterns
	var teamSizeIndicator string
	var confidence float64

	// Check for specific team size indicators
	for _, pattern := range cse.teamSizePatterns {
		if pattern.MatchString(text) {
			teamSizeIndicator = cse.getTeamSizeIndicator(pattern.String())
			confidence = 0.6 // Medium confidence for pattern matches
			break
		}
	}

	// If no specific indicator found, try to infer from context
	if teamSizeIndicator == "" {
		teamSizeIndicator, confidence = cse.inferTeamSizeIndicator(text)
	}

	// Set team size information
	if teamSizeIndicator != "" {
		result.TeamSizeIndicator = teamSizeIndicator
		result.TeamSizeConfidence = confidence
	}

	span.SetAttributes(
		attribute.String("team_size_indicator", teamSizeIndicator),
		attribute.Float64("confidence", confidence),
	)

	return nil
}

// combineText combines all available text for analysis
func (cse *CompanySizeExtractor) combineText(businessData *shared.BusinessClassificationRequest) string {
	var parts []string

	// Add business name
	if businessData.BusinessName != "" {
		parts = append(parts, businessData.BusinessName)
	}

	// Add description
	if businessData.Description != "" {
		parts = append(parts, businessData.Description)
	}

	// Add keywords
	if len(businessData.Keywords) > 0 {
		parts = append(parts, strings.Join(businessData.Keywords, " "))
	}

	// Add address
	if businessData.Address != "" {
		parts = append(parts, businessData.Address)
	}

	// Combine all parts
	text := strings.Join(parts, " ")

	// Normalize text
	if !cse.config.CaseSensitive {
		text = strings.ToLower(text)
	}

	return text
}

// inferEmployeeCount infers employee count from context
func (cse *CompanySizeExtractor) inferEmployeeCount(text string) (int, float64) {
	// Check for startup indicators
	if cse.hasStartupIndicators(text) {
		return 15, 0.5 // Typical startup size
	}

	// Check for enterprise indicators
	if cse.hasEnterpriseIndicators(text) {
		return 1000, 0.6 // Typical enterprise size
	}

	// Check for small business indicators
	if cse.hasSmallBusinessIndicators(text) {
		return 25, 0.5 // Typical small business size
	}

	// Default to unknown
	return 0, 0.0
}

// inferRevenueIndicator infers revenue indicator from context
func (cse *CompanySizeExtractor) inferRevenueIndicator(text string) (string, float64) {
	// Check for startup indicators
	if cse.hasStartupIndicators(text) {
		return "startup", 0.6
	}

	// Check for enterprise indicators
	if cse.hasEnterpriseIndicators(text) {
		return "enterprise", 0.7
	}

	// Check for small business indicators
	if cse.hasSmallBusinessIndicators(text) {
		return "small", 0.5
	}

	// Default to unknown
	return "unknown", 0.0
}

// inferLocationCount infers location count from context
func (cse *CompanySizeExtractor) inferLocationCount(text string) (int, float64) {
	// Check for global/multinational indicators
	if cse.hasGlobalIndicators(text) {
		return 10, 0.6
	}

	// Check for local/single location indicators
	if cse.hasLocalIndicators(text) {
		return 1, 0.7
	}

	// Default to single location
	return 1, 0.3
}

// inferTeamSizeIndicator infers team size indicator from context
func (cse *CompanySizeExtractor) inferTeamSizeIndicator(text string) (string, float64) {
	// Check for startup indicators
	if cse.hasStartupIndicators(text) {
		return "small_team", 0.6
	}

	// Check for enterprise indicators
	if cse.hasEnterpriseIndicators(text) {
		return "large_team", 0.7
	}

	// Default to unknown
	return "unknown", 0.0
}

// Helper methods for pattern detection
func (cse *CompanySizeExtractor) hasStartupIndicators(text string) bool {
	for _, pattern := range cse.startupPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (cse *CompanySizeExtractor) hasEnterpriseIndicators(text string) bool {
	for _, pattern := range cse.enterprisePatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (cse *CompanySizeExtractor) hasSmallBusinessIndicators(text string) bool {
	smallBusinessPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)small\s+business`),
		regexp.MustCompile(`(?i)local\s+business`),
		regexp.MustCompile(`(?i)family\s+owned`),
		regexp.MustCompile(`(?i)independent`),
		regexp.MustCompile(`(?i)boutique`),
	}

	for _, pattern := range smallBusinessPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (cse *CompanySizeExtractor) hasGlobalIndicators(text string) bool {
	globalPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)global`),
		regexp.MustCompile(`(?i)worldwide`),
		regexp.MustCompile(`(?i)international`),
		regexp.MustCompile(`(?i)multinational`),
		regexp.MustCompile(`(?i)across\s+continents`),
	}

	for _, pattern := range globalPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (cse *CompanySizeExtractor) hasLocalIndicators(text string) bool {
	localPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)local`),
		regexp.MustCompile(`(?i)community`),
		regexp.MustCompile(`(?i)neighborhood`),
		regexp.MustCompile(`(?i)regional`),
	}

	for _, pattern := range localPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// Helper methods for categorization
func (cse *CompanySizeExtractor) getEmployeeCountRange(count int) string {
	for rangeName, rangeInfo := range EmployeeCountRanges {
		if count >= rangeInfo.Min && count <= rangeInfo.Max {
			return rangeName
		}
	}
	return "unknown"
}

func (cse *CompanySizeExtractor) getRevenueIndicator(revenue float64) string {
	if revenue < 1 {
		return "startup"
	} else if revenue < 10 {
		return "small"
	} else if revenue < 100 {
		return "medium"
	} else if revenue < 1000 {
		return "large"
	} else {
		return "enterprise"
	}
}

func (cse *CompanySizeExtractor) getRevenueRange(revenue float64) string {
	if revenue < 1 {
		return "< $1M"
	} else if revenue < 10 {
		return "$1M - $10M"
	} else if revenue < 100 {
		return "$10M - $100M"
	} else if revenue < 1000 {
		return "$100M - $1B"
	} else {
		return "> $1B"
	}
}

func (cse *CompanySizeExtractor) getTeamSizeIndicator(pattern string) string {
	if strings.Contains(strings.ToLower(pattern), "small") {
		return "small_team"
	} else if strings.Contains(strings.ToLower(pattern), "large") {
		return "large_team"
	} else if strings.Contains(strings.ToLower(pattern), "startup") {
		return "startup_team"
	} else if strings.Contains(strings.ToLower(pattern), "global") {
		return "global_team"
	} else if strings.Contains(strings.ToLower(pattern), "remote") {
		return "remote_team"
	}
	return "standard_team"
}

// determineCompanySizeCategory determines the overall company size category
func (cse *CompanySizeExtractor) determineCompanySizeCategory(result *CompanySize) {
	// Use employee count as primary indicator
	if result.EmployeeCountRange != "" {
		switch result.EmployeeCountRange {
		case "1-10":
			result.CompanySizeCategory = "startup"
		case "11-50":
			result.CompanySizeCategory = "small"
		case "51-200":
			result.CompanySizeCategory = "medium"
		case "201-500", "501-1000":
			result.CompanySizeCategory = "large"
		case "1000+":
			result.CompanySizeCategory = "enterprise"
		}
		return
	}

	// Fall back to revenue indicator
	if result.RevenueIndicator != "" {
		result.CompanySizeCategory = result.RevenueIndicator
		return
	}

	// Default to unknown
	result.CompanySizeCategory = "unknown"
}

// calculateOverallConfidence calculates the overall confidence score
func (cse *CompanySizeExtractor) calculateOverallConfidence(result *CompanySize) {
	var scores []float64
	var weights []float64

	// Employee confidence
	if result.EmployeeConfidence > 0 {
		scores = append(scores, result.EmployeeConfidence)
		weights = append(weights, 0.4) // 40% weight
	}

	// Revenue confidence
	if result.RevenueConfidence > 0 {
		scores = append(scores, result.RevenueConfidence)
		weights = append(weights, 0.3) // 30% weight
	}

	// Location confidence
	if result.LocationsConfidence > 0 {
		scores = append(scores, result.LocationsConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	// Team size confidence
	if result.TeamSizeConfidence > 0 {
		scores = append(scores, result.TeamSizeConfidence)
		weights = append(weights, 0.1) // 10% weight
	}

	// Calculate weighted average
	if len(scores) > 0 {
		totalWeight := 0.0
		weightedSum := 0.0

		for i, score := range scores {
			weight := weights[i]
			weightedSum += score * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			result.OverallConfidence = weightedSum / totalWeight
		}
	} else {
		result.OverallConfidence = 0.0
	}
}

// validateResults validates the extracted results
func (cse *CompanySizeExtractor) validateResults(result *CompanySize) error {
	// Validate employee count
	if result.EmployeeCountMin > 0 && result.EmployeeCountMin > cse.config.MaxEmployeeCount {
		return fmt.Errorf("employee count %d exceeds maximum allowed %d", result.EmployeeCountMin, cse.config.MaxEmployeeCount)
	}

	// Validate confidence scores
	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		return fmt.Errorf("overall confidence score %f is out of range [0,1]", result.OverallConfidence)
	}

	// Validate location count
	if result.OfficeLocationsCount > cse.config.MaxLocations {
		return fmt.Errorf("location count %d exceeds maximum allowed %d", result.OfficeLocationsCount, cse.config.MaxLocations)
	}

	return nil
}
