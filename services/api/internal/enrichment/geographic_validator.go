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

// GeographicValidator validates and standardizes geographic data
type GeographicValidator struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *GeographicValidatorConfig
}

// GeographicValidatorConfig contains configuration for geographic validation
type GeographicValidatorConfig struct {
	// Country mappings and validations
	CountryMappings    map[string]string   // Country name standardization
	CountryCodeMapping map[string]string   // ISO country codes to names
	ValidCountries     map[string]bool     // Valid country names
	ValidRegions       map[string][]string // Valid regions and their countries

	// Address validation patterns
	PostalCodePatterns map[string]string   // Postal code patterns by country
	PhonePatterns      map[string]string   // Phone number patterns by country
	AddressPatterns    map[string][]string // Address format patterns by country

	// Coordinate validation
	MinLatitude  float64 // Minimum valid latitude
	MaxLatitude  float64 // Maximum valid latitude
	MinLongitude float64 // Minimum valid longitude
	MaxLongitude float64 // Maximum valid longitude

	// Validation thresholds
	MinConfidenceScore  float64 // Minimum confidence for validation
	MaxValidationErrors int     // Maximum validation errors allowed
}

// ValidationError represents a geographic validation error
type ValidationError struct {
	Field      string  `json:"field"`      // Field that failed validation
	Value      string  `json:"value"`      // Invalid value
	Error      string  `json:"error"`      // Error description
	Severity   string  `json:"severity"`   // error, warning, info
	Suggestion string  `json:"suggestion"` // Suggested correction
	Confidence float64 `json:"confidence"` // Confidence in suggestion
}

// StandardizedLocation represents a validated and standardized location
type StandardizedLocation struct {
	// Original data
	OriginalAddress string `json:"original_address"`
	OriginalCountry string `json:"original_country"`

	// Standardized data
	Address     string `json:"address"`      // Standardized address
	Street      string `json:"street"`       // Street address
	City        string `json:"city"`         // City name
	State       string `json:"state"`        // State/province (standardized)
	Country     string `json:"country"`      // Country (standardized)
	CountryCode string `json:"country_code"` // ISO country code
	PostalCode  string `json:"postal_code"`  // Postal code (validated)
	Region      string `json:"region"`       // Geographic region

	// Coordinates (if available)
	Latitude  *float64 `json:"latitude"`  // Latitude coordinate
	Longitude *float64 `json:"longitude"` // Longitude coordinate

	// Validation metadata
	IsValid          bool              `json:"is_valid"`          // Whether location is valid
	ConfidenceScore  float64           `json:"confidence_score"`  // Validation confidence
	ValidationErrors []ValidationError `json:"validation_errors"` // Validation errors
	StandardizedAt   time.Time         `json:"standardized_at"`   // When standardized
}

// StandardizedGeography represents validated geographic data
type StandardizedGeography struct {
	Countries    []string               `json:"countries"`     // Standardized country names
	CountryCodes []string               `json:"country_codes"` // ISO country codes
	Regions      []string               `json:"regions"`       // Standardized regions
	ServiceAreas []StandardizedLocation `json:"service_areas"` // Validated service areas
	Locations    []StandardizedLocation `json:"locations"`     // Validated locations

	// Validation summary
	IsValid          bool              `json:"is_valid"`          // Overall validity
	ConfidenceScore  float64           `json:"confidence_score"`  // Overall confidence
	ValidationErrors []ValidationError `json:"validation_errors"` // All validation errors
	ProcessingTime   time.Duration     `json:"processing_time"`   // Processing time
}

// GeographicValidationResult contains the results of geographic validation
type GeographicValidationResult struct {
	StandardizedGeography *StandardizedGeography `json:"standardized_geography"` // Standardized data
	OriginalData          interface{}            `json:"original_data"`          // Original input data
	ValidationSummary     *ValidationSummary     `json:"validation_summary"`     // Validation summary
	ProcessingTime        time.Duration          `json:"processing_time"`        // Time taken
}

// ValidationSummary provides an overview of validation results
type ValidationSummary struct {
	TotalFields     int            `json:"total_fields"`    // Total fields validated
	ValidFields     int            `json:"valid_fields"`    // Valid fields count
	InvalidFields   int            `json:"invalid_fields"`  // Invalid fields count
	WarningFields   int            `json:"warning_fields"`  // Fields with warnings
	ErrorsByType    map[string]int `json:"errors_by_type"`  // Error counts by type
	OverallScore    float64        `json:"overall_score"`   // Overall validation score
	Recommendations []string       `json:"recommendations"` // Improvement recommendations
}

// NewGeographicValidator creates a new geographic validator
func NewGeographicValidator(logger *zap.Logger, config *GeographicValidatorConfig) *GeographicValidator {
	if config == nil {
		config = getDefaultGeographicValidatorConfig()
	}

	return &GeographicValidator{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("geographic_validator"),
		config: config,
	}
}

// ValidateLocation validates and standardizes a single location
func (gv *GeographicValidator) ValidateLocation(ctx context.Context, location Location) (*StandardizedLocation, error) {
	ctx, span := gv.tracer.Start(ctx, "geographic_validator.validate_location")
	defer span.End()

	standardized := &StandardizedLocation{
		OriginalAddress: location.Address,
		OriginalCountry: location.Country,
		StandardizedAt:  time.Now(),
	}

	var errors []ValidationError

	// Validate and standardize country
	country, countryCode, err := gv.validateCountry(location.Country)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:      "country",
			Value:      location.Country,
			Error:      err.Error(),
			Severity:   "error",
			Suggestion: gv.suggestCountry(location.Country),
			Confidence: 0.6,
		})
	} else {
		standardized.Country = country
		standardized.CountryCode = countryCode
	}

	// Validate postal code
	if location.PostalCode != "" {
		isValid, suggestion := gv.validatePostalCode(location.PostalCode, countryCode)
		if !isValid {
			errors = append(errors, ValidationError{
				Field:      "postal_code",
				Value:      location.PostalCode,
				Error:      "Invalid postal code format",
				Severity:   "warning",
				Suggestion: suggestion,
				Confidence: 0.7,
			})
		} else {
			standardized.PostalCode = gv.standardizePostalCode(location.PostalCode, countryCode)
		}
	}

	// Validate phone number
	if location.Phone != "" {
		isValid, standardizedPhone := gv.validatePhoneNumber(location.Phone, countryCode)
		if !isValid {
			errors = append(errors, ValidationError{
				Field:      "phone",
				Value:      location.Phone,
				Error:      "Invalid phone number format",
				Severity:   "warning",
				Suggestion: standardizedPhone,
				Confidence: 0.5,
			})
		}
	}

	// Standardize address components
	standardized.Address = gv.standardizeAddress(location.Address)
	standardized.City = gv.standardizeCity(location.City)
	standardized.State = gv.standardizeState(location.State, countryCode)

	// Determine region
	if standardized.Country != "" {
		standardized.Region = gv.getRegionForCountry(standardized.Country)
	}

	// Calculate confidence score
	standardized.ConfidenceScore = gv.calculateLocationConfidence(standardized, errors)
	standardized.IsValid = len(errors) == 0 || gv.hasOnlyWarnings(errors)
	standardized.ValidationErrors = errors

	gv.logger.Debug("location validated",
		zap.String("original_address", location.Address),
		zap.String("standardized_country", standardized.Country),
		zap.Float64("confidence", standardized.ConfidenceScore),
		zap.Int("errors", len(errors)))

	return standardized, nil
}

// ValidateGeography validates comprehensive geographic data from multiple extractors
func (gv *GeographicValidator) ValidateGeography(ctx context.Context, data interface{}) (*GeographicValidationResult, error) {
	ctx, span := gv.tracer.Start(ctx, "geographic_validator.validate_geography")
	defer span.End()

	startTime := time.Now()

	standardized := &StandardizedGeography{
		Countries:    []string{},
		CountryCodes: []string{},
		Regions:      []string{},
		ServiceAreas: []StandardizedLocation{},
		Locations:    []StandardizedLocation{},
	}

	var allErrors []ValidationError

	// Handle different input types
	switch d := data.(type) {
	case *LocationResult:
		err := gv.validateLocationResult(ctx, d, standardized, &allErrors)
		if err != nil {
			return nil, fmt.Errorf("failed to validate location result: %w", err)
		}

	case *InternationalPresenceResult:
		err := gv.validateInternationalPresenceResult(ctx, d, standardized, &allErrors)
		if err != nil {
			return nil, fmt.Errorf("failed to validate international presence result: %w", err)
		}

	case *MarketCoverageResult:
		err := gv.validateMarketCoverageResult(ctx, d, standardized, &allErrors)
		if err != nil {
			return nil, fmt.Errorf("failed to validate market coverage result: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	// Remove duplicates and standardize
	standardized.Countries = gv.removeDuplicateStrings(standardized.Countries)
	standardized.CountryCodes = gv.removeDuplicateStrings(standardized.CountryCodes)
	standardized.Regions = gv.removeDuplicateStrings(standardized.Regions)

	// Calculate overall validation scores
	standardized.ConfidenceScore = gv.calculateOverallConfidence(standardized, allErrors)
	standardized.IsValid = len(allErrors) == 0 || gv.hasOnlyWarnings(allErrors)
	standardized.ValidationErrors = allErrors
	standardized.ProcessingTime = time.Since(startTime)

	// Create validation summary
	summary := gv.createValidationSummary(standardized, allErrors)

	result := &GeographicValidationResult{
		StandardizedGeography: standardized,
		OriginalData:          data,
		ValidationSummary:     summary,
		ProcessingTime:        time.Since(startTime),
	}

	gv.logger.Info("geographic validation completed",
		zap.Int("countries", len(standardized.Countries)),
		zap.Int("locations", len(standardized.Locations)),
		zap.Float64("confidence", standardized.ConfidenceScore),
		zap.Int("errors", len(allErrors)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// validateCountry validates and standardizes a country name
func (gv *GeographicValidator) validateCountry(country string) (string, string, error) {
	if country == "" {
		return "", "", fmt.Errorf("country is empty")
	}

	normalized := strings.ToLower(strings.TrimSpace(country))

	// Check direct mapping
	if standardized, exists := gv.config.CountryMappings[normalized]; exists {
		if code, codeExists := gv.config.CountryCodeMapping[standardized]; codeExists {
			return standardized, code, nil
		}
		return standardized, "", nil
	}

	// Check if already standardized
	if gv.config.ValidCountries[country] {
		if code, exists := gv.config.CountryCodeMapping[country]; exists {
			return country, code, nil
		}
		return country, "", nil
	}

	return "", "", fmt.Errorf("invalid or unrecognized country: %s", country)
}

// validatePostalCode validates postal code format for a specific country
func (gv *GeographicValidator) validatePostalCode(postalCode, countryCode string) (bool, string) {
	if postalCode == "" {
		return true, "" // Empty is valid
	}

	pattern, exists := gv.config.PostalCodePatterns[countryCode]
	if !exists {
		return true, postalCode // No pattern available, assume valid
	}

	matched, _ := regexp.MatchString(pattern, postalCode)
	if matched {
		return true, postalCode
	}

	// Try to suggest a correction
	suggestion := gv.suggestPostalCodeCorrection(postalCode, countryCode)
	return false, suggestion
}

// validatePhoneNumber validates phone number format
func (gv *GeographicValidator) validatePhoneNumber(phone, countryCode string) (bool, string) {
	if phone == "" {
		return true, ""
	}

	// Remove common separators and spaces
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	pattern, exists := gv.config.PhonePatterns[countryCode]
	if !exists {
		// Generic validation - must start with + and have 7-15 digits
		if matched, _ := regexp.MatchString(`^\+?\d{7,15}$`, cleaned); matched {
			return true, cleaned
		}
		return false, cleaned
	}

	matched, _ := regexp.MatchString(pattern, cleaned)
	return matched, cleaned
}

// validateCoordinates validates latitude and longitude coordinates
func (gv *GeographicValidator) validateCoordinates(lat, lng float64) error {
	if lat < gv.config.MinLatitude || lat > gv.config.MaxLatitude {
		return fmt.Errorf("invalid latitude: %f (must be between %f and %f)",
			lat, gv.config.MinLatitude, gv.config.MaxLatitude)
	}

	if lng < gv.config.MinLongitude || lng > gv.config.MaxLongitude {
		return fmt.Errorf("invalid longitude: %f (must be between %f and %f)",
			lng, gv.config.MinLongitude, gv.config.MaxLongitude)
	}

	return nil
}

// Helper functions for validation of different result types
func (gv *GeographicValidator) validateLocationResult(ctx context.Context, result *LocationResult, standardized *StandardizedGeography, allErrors *[]ValidationError) error {
	for _, location := range result.Locations {
		validatedLocation, err := gv.ValidateLocation(ctx, location)
		if err != nil {
			return fmt.Errorf("failed to validate location: %w", err)
		}

		standardized.Locations = append(standardized.Locations, *validatedLocation)
		*allErrors = append(*allErrors, validatedLocation.ValidationErrors...)

		if validatedLocation.Country != "" {
			standardized.Countries = append(standardized.Countries, validatedLocation.Country)
		}
		if validatedLocation.CountryCode != "" {
			standardized.CountryCodes = append(standardized.CountryCodes, validatedLocation.CountryCode)
		}
		if validatedLocation.Region != "" {
			standardized.Regions = append(standardized.Regions, validatedLocation.Region)
		}
	}

	return nil
}

func (gv *GeographicValidator) validateInternationalPresenceResult(ctx context.Context, result *InternationalPresenceResult, standardized *StandardizedGeography, allErrors *[]ValidationError) error {
	// Validate countries
	for _, country := range result.Countries {
		validatedCountry, countryCode, err := gv.validateCountry(country)
		if err != nil {
			*allErrors = append(*allErrors, ValidationError{
				Field:      "country",
				Value:      country,
				Error:      err.Error(),
				Severity:   "warning",
				Suggestion: gv.suggestCountry(country),
				Confidence: 0.6,
			})
		} else {
			standardized.Countries = append(standardized.Countries, validatedCountry)
			if countryCode != "" {
				standardized.CountryCodes = append(standardized.CountryCodes, countryCode)
			}
		}
	}

	// Validate regions
	for _, region := range result.Regions {
		if gv.isValidRegion(region) {
			standardized.Regions = append(standardized.Regions, region)
		} else {
			*allErrors = append(*allErrors, ValidationError{
				Field:      "region",
				Value:      region,
				Error:      "Invalid or unrecognized region",
				Severity:   "warning",
				Suggestion: gv.suggestRegion(region),
				Confidence: 0.5,
			})
		}
	}

	return nil
}

func (gv *GeographicValidator) validateMarketCoverageResult(ctx context.Context, result *MarketCoverageResult, standardized *StandardizedGeography, allErrors *[]ValidationError) error {
	for _, serviceArea := range result.ServiceAreas {
		// Convert ServiceArea to Location for validation
		location := Location{
			Address:    serviceArea.Description,
			City:       "", // ServiceArea doesn't have structured address
			State:      "",
			Country:    "",
			PostalCode: "",
		}

		// Extract country from service area if available
		for _, country := range serviceArea.Countries {
			location.Country = country
			break // Use first country for validation
		}

		validatedLocation, err := gv.ValidateLocation(ctx, location)
		if err != nil {
			return fmt.Errorf("failed to validate service area: %w", err)
		}

		standardized.ServiceAreas = append(standardized.ServiceAreas, *validatedLocation)
		*allErrors = append(*allErrors, validatedLocation.ValidationErrors...)

		// Add countries from service area
		for _, country := range serviceArea.Countries {
			validatedCountry, countryCode, err := gv.validateCountry(country)
			if err != nil {
				*allErrors = append(*allErrors, ValidationError{
					Field:      "service_area_country",
					Value:      country,
					Error:      err.Error(),
					Severity:   "warning",
					Suggestion: gv.suggestCountry(country),
					Confidence: 0.6,
				})
			} else {
				standardized.Countries = append(standardized.Countries, validatedCountry)
				if countryCode != "" {
					standardized.CountryCodes = append(standardized.CountryCodes, countryCode)
				}
			}
		}
	}

	return nil
}

// Helper functions continue...
