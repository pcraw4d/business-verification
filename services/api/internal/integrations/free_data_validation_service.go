package integrations

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FreeDataValidationService provides free data validation by cross-referencing
// government APIs and validating business information consistency
type FreeDataValidationService struct {
	config           FreeDataValidationConfig
	logger           *zap.Logger
	governmentAPIs   *BusinessDataAPIService
	validationCache  map[string]*ValidationResult
	cacheMutex       sync.RWMutex
	qualityThreshold float64
}

// FreeDataValidationConfig holds configuration for free data validation
type FreeDataValidationConfig struct {
	// Quality thresholds
	MinQualityScore    float64 `json:"min_quality_score"`   // Minimum quality score to accept
	ConsistencyWeight  float64 `json:"consistency_weight"`  // Weight for consistency checks
	CompletenessWeight float64 `json:"completeness_weight"` // Weight for data completeness
	AccuracyWeight     float64 `json:"accuracy_weight"`     // Weight for accuracy checks
	FreshnessWeight    float64 `json:"freshness_weight"`    // Weight for data freshness

	// Validation settings
	EnableCrossReference   bool `json:"enable_cross_reference"`   // Enable cross-referencing
	MaxValidationTime      int  `json:"max_validation_time"`      // Max validation time in seconds
	CacheValidationResults bool `json:"cache_validation_results"` // Cache validation results

	// Cost control
	MaxAPICallsPerValidation int `json:"max_api_calls_per_validation"` // Max API calls per validation
	RateLimitDelay           int `json:"rate_limit_delay"`             // Delay between API calls in ms
}

// ValidationResult represents the result of free data validation
type ValidationResult struct {
	BusinessID            string                 `json:"business_id"`
	IsValid               bool                   `json:"is_valid"`
	QualityScore          float64                `json:"quality_score"`
	ConsistencyScore      float64                `json:"consistency_score"`
	CompletenessScore     float64                `json:"completeness_score"`
	AccuracyScore         float64                `json:"accuracy_score"`
	FreshnessScore        float64                `json:"freshness_score"`
	CrossReferenceResults map[string]interface{} `json:"cross_reference_results"`
	ValidationErrors      []ValidationError      `json:"validation_errors"`
	ValidationWarnings    []ValidationWarning    `json:"validation_warnings"`
	DataSources           []DataSourceInfo       `json:"data_sources"`
	ValidatedAt           time.Time              `json:"validated_at"`
	ValidationTime        time.Duration          `json:"validation_time"`
	Cost                  float64                `json:"cost"` // Always 0.0 for free validation
}

// ValidationError represents a validation error
type ValidationError struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Code     string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Code     string `json:"code"`
}

// DataSourceInfo represents information about a data source used in validation
type DataSourceInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	IsFree      bool      `json:"is_free"`
	Cost        float64   `json:"cost"`
	LastUpdated time.Time `json:"last_updated"`
	Reliability float64   `json:"reliability"`
}

// BusinessDataForValidation represents business data to be validated
type BusinessDataForValidation struct {
	BusinessID         string `json:"business_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Address            string `json:"address"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Website            string `json:"website"`
	Industry           string `json:"industry"`
	Country            string `json:"country"`
	RegistrationNumber string `json:"registration_number"`
	TaxID              string `json:"tax_id"`
}

// NewFreeDataValidationService creates a new free data validation service
func NewFreeDataValidationService(config FreeDataValidationConfig, logger *zap.Logger, governmentAPIs *BusinessDataAPIService) *FreeDataValidationService {
	return &FreeDataValidationService{
		config:           config,
		logger:           logger,
		governmentAPIs:   governmentAPIs,
		validationCache:  make(map[string]*ValidationResult),
		qualityThreshold: config.MinQualityScore,
	}
}

// GetDefaultFreeDataValidationConfig returns default configuration for free data validation
func GetDefaultFreeDataValidationConfig() FreeDataValidationConfig {
	return FreeDataValidationConfig{
		MinQualityScore:          0.7,  // 70% minimum quality score
		ConsistencyWeight:        0.3,  // 30% weight for consistency
		CompletenessWeight:       0.25, // 25% weight for completeness
		AccuracyWeight:           0.25, // 25% weight for accuracy
		FreshnessWeight:          0.2,  // 20% weight for freshness
		EnableCrossReference:     true, // Enable cross-referencing
		MaxValidationTime:        30,   // 30 seconds max validation time
		CacheValidationResults:   true, // Cache validation results
		MaxAPICallsPerValidation: 10,   // Max 10 API calls per validation
		RateLimitDelay:           100,  // 100ms delay between API calls
	}
}

// ValidateBusinessData validates business data using free government APIs
func (s *FreeDataValidationService) ValidateBusinessData(ctx context.Context, data BusinessDataForValidation) (*ValidationResult, error) {
	startTime := time.Now()

	// Check cache first
	if s.config.CacheValidationResults {
		if cached := s.getCachedValidation(data.BusinessID); cached != nil {
			s.logger.Debug("Using cached validation result",
				zap.String("business_id", data.BusinessID),
				zap.Float64("quality_score", cached.QualityScore))
			return cached, nil
		}
	}

	// Create validation result
	result := &ValidationResult{
		BusinessID:            data.BusinessID,
		IsValid:               true,
		QualityScore:          0.0,
		ConsistencyScore:      0.0,
		CompletenessScore:     0.0,
		AccuracyScore:         0.0,
		FreshnessScore:        0.0,
		CrossReferenceResults: make(map[string]interface{}),
		ValidationErrors:      []ValidationError{},
		ValidationWarnings:    []ValidationWarning{},
		DataSources:           []DataSourceInfo{},
		ValidatedAt:           time.Now(),
		Cost:                  0.0, // Always free
	}

	// Set timeout for validation
	validationCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.MaxValidationTime)*time.Second)
	defer cancel()

	// Perform validation steps
	if err := s.validateCompleteness(data, result); err != nil {
		s.logger.Error("Completeness validation failed", zap.Error(err))
	}

	if err := s.validateAccuracy(validationCtx, data, result); err != nil {
		s.logger.Error("Accuracy validation failed", zap.Error(err))
	}

	if s.config.EnableCrossReference {
		if err := s.validateCrossReference(validationCtx, data, result); err != nil {
			s.logger.Error("Cross-reference validation failed", zap.Error(err))
		}
	}

	if err := s.validateConsistency(data, result); err != nil {
		s.logger.Error("Consistency validation failed", zap.Error(err))
	}

	if err := s.validateFreshness(data, result); err != nil {
		s.logger.Error("Freshness validation failed", zap.Error(err))
	}

	// Calculate overall quality score
	s.calculateQualityScore(result)

	// Determine if validation passed
	result.IsValid = result.QualityScore >= s.qualityThreshold && len(result.ValidationErrors) == 0

	// Record validation time
	result.ValidationTime = time.Since(startTime)

	// Cache result if enabled
	if s.config.CacheValidationResults {
		s.cacheValidationResult(data.BusinessID, result)
	}

	s.logger.Info("Free data validation completed",
		zap.String("business_id", data.BusinessID),
		zap.Bool("is_valid", result.IsValid),
		zap.Float64("quality_score", result.QualityScore),
		zap.Duration("validation_time", result.ValidationTime),
		zap.Float64("cost", result.Cost))

	return result, nil
}

// validateCompleteness validates data completeness
func (s *FreeDataValidationService) validateCompleteness(data BusinessDataForValidation, result *ValidationResult) error {
	completenessScore := 0.0

	// Check required fields
	requiredFields := map[string]string{
		"name":        data.Name,
		"description": data.Description,
		"address":     data.Address,
		"country":     data.Country,
	}

	optionalFields := map[string]string{
		"phone":               data.Phone,
		"email":               data.Email,
		"website":             data.Website,
		"registration_number": data.RegistrationNumber,
	}

	// Check required fields
	for field, value := range requiredFields {
		if strings.TrimSpace(value) != "" {
			completenessScore += 0.5 // Required fields worth 50% of score
		} else {
			result.ValidationErrors = append(result.ValidationErrors, ValidationError{
				Field:    field,
				Message:  fmt.Sprintf("Required field %s is missing", field),
				Severity: "high",
				Source:   "completeness_validation",
				Code:     "MISSING_REQUIRED_FIELD",
			})
		}
	}

	// Check optional fields
	for field, value := range optionalFields {
		if strings.TrimSpace(value) != "" {
			completenessScore += 0.125 // Optional fields worth 12.5% each
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    field,
				Message:  fmt.Sprintf("Optional field %s is missing", field),
				Severity: "low",
				Source:   "completeness_validation",
				Code:     "MISSING_OPTIONAL_FIELD",
			})
		}
	}

	result.CompletenessScore = completenessScore
	return nil
}

// validateAccuracy validates data accuracy using free government APIs
func (s *FreeDataValidationService) validateAccuracy(ctx context.Context, data BusinessDataForValidation, result *ValidationResult) error {
	accuracyScore := 0.0
	apiCalls := 0

	// Validate business name against government registries
	if data.Name != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if err := s.validateBusinessName(ctx, data, result); err != nil {
			s.logger.Warn("Business name validation failed", zap.Error(err))
		} else {
			accuracyScore += 0.3
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	// Validate registration number if provided
	if data.RegistrationNumber != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if err := s.validateRegistrationNumber(ctx, data, result); err != nil {
			s.logger.Warn("Registration number validation failed", zap.Error(err))
		} else {
			accuracyScore += 0.3
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	// Validate domain/website if provided
	if data.Website != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if err := s.validateWebsite(ctx, data, result); err != nil {
			s.logger.Warn("Website validation failed", zap.Error(err))
		} else {
			accuracyScore += 0.2
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	// Validate address format
	if data.Address != "" {
		if s.isValidAddressFormat(data.Address) {
			accuracyScore += 0.2
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    "address",
				Message:  "Address format appears to be invalid",
				Severity: "medium",
				Source:   "accuracy_validation",
				Code:     "INVALID_ADDRESS_FORMAT",
			})
		}
	}

	result.AccuracyScore = accuracyScore
	return nil
}

// validateCrossReference cross-references data across multiple free government sources
func (s *FreeDataValidationService) validateCrossReference(ctx context.Context, data BusinessDataForValidation, result *ValidationResult) error {
	consistencyScore := 0.0
	apiCalls := 0

	// Cross-reference with SEC EDGAR (US companies)
	if data.Country == "US" && data.Name != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if secData, err := s.crossReferenceSECEdgar(ctx, data); err == nil {
			result.CrossReferenceResults["sec_edgar"] = secData
			if s.isDataConsistent(data, secData, "sec_edgar") {
				consistencyScore += 0.4
			}
			result.DataSources = append(result.DataSources, DataSourceInfo{
				Name:        "SEC EDGAR",
				Type:        "government_registry",
				IsFree:      true,
				Cost:        0.0,
				LastUpdated: time.Now(),
				Reliability: 0.95,
			})
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	// Cross-reference with Companies House (UK companies)
	if data.Country == "UK" && data.Name != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if chData, err := s.crossReferenceCompaniesHouse(ctx, data); err == nil {
			result.CrossReferenceResults["companies_house"] = chData
			if s.isDataConsistent(data, chData, "companies_house") {
				consistencyScore += 0.4
			}
			result.DataSources = append(result.DataSources, DataSourceInfo{
				Name:        "Companies House",
				Type:        "government_registry",
				IsFree:      true,
				Cost:        0.0,
				LastUpdated: time.Now(),
				Reliability: 0.95,
			})
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	// Cross-reference with OpenCorporates (global companies)
	if data.Name != "" && apiCalls < s.config.MaxAPICallsPerValidation {
		if ocData, err := s.crossReferenceOpenCorporates(ctx, data); err == nil {
			result.CrossReferenceResults["opencorporates"] = ocData
			if s.isDataConsistent(data, ocData, "opencorporates") {
				consistencyScore += 0.2
			}
			result.DataSources = append(result.DataSources, DataSourceInfo{
				Name:        "OpenCorporates",
				Type:        "business_registry",
				IsFree:      true,
				Cost:        0.0,
				LastUpdated: time.Now(),
				Reliability: 0.85,
			})
		}
		apiCalls++
		time.Sleep(time.Duration(s.config.RateLimitDelay) * time.Millisecond)
	}

	result.ConsistencyScore = consistencyScore
	return nil
}

// validateConsistency validates internal data consistency
func (s *FreeDataValidationService) validateConsistency(data BusinessDataForValidation, result *ValidationResult) error {
	consistencyScore := 0.0

	// Check email domain consistency with website
	if data.Email != "" && data.Website != "" {
		if s.isEmailDomainConsistent(data.Email, data.Website) {
			consistencyScore += 0.3
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    "email",
				Message:  "Email domain does not match website domain",
				Severity: "medium",
				Source:   "consistency_validation",
				Code:     "EMAIL_DOMAIN_MISMATCH",
			})
		}
	}

	// Check phone number format consistency
	if data.Phone != "" {
		if s.isValidPhoneFormat(data.Phone) {
			consistencyScore += 0.2
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    "phone",
				Message:  "Phone number format appears to be invalid",
				Severity: "medium",
				Source:   "consistency_validation",
				Code:     "INVALID_PHONE_FORMAT",
			})
		}
	}

	// Check business name and description consistency
	if data.Name != "" && data.Description != "" {
		if s.isNameDescriptionConsistent(data.Name, data.Description) {
			consistencyScore += 0.3
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    "description",
				Message:  "Business description does not appear to match business name",
				Severity: "low",
				Source:   "consistency_validation",
				Code:     "NAME_DESCRIPTION_MISMATCH",
			})
		}
	}

	// Check address and country consistency
	if data.Address != "" && data.Country != "" {
		if s.isAddressCountryConsistent(data.Address, data.Country) {
			consistencyScore += 0.2
		} else {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Field:    "address",
				Message:  "Address does not appear to be in the specified country",
				Severity: "medium",
				Source:   "consistency_validation",
				Code:     "ADDRESS_COUNTRY_MISMATCH",
			})
		}
	}

	result.ConsistencyScore = consistencyScore
	return nil
}

// validateFreshness validates data freshness
func (s *FreeDataValidationService) validateFreshness(data BusinessDataForValidation, result *ValidationResult) error {
	freshnessScore := 1.0 // Start with perfect freshness

	// For free validation, we assume data is fresh since we're validating in real-time
	// In a real implementation, you might check when the data was last updated
	// or when the business was last verified

	result.FreshnessScore = freshnessScore
	return nil
}

// calculateQualityScore calculates the overall quality score
func (s *FreeDataValidationService) calculateQualityScore(result *ValidationResult) {
	result.QualityScore =
		result.ConsistencyScore*s.config.ConsistencyWeight +
			result.CompletenessScore*s.config.CompletenessWeight +
			result.AccuracyScore*s.config.AccuracyWeight +
			result.FreshnessScore*s.config.FreshnessWeight
}

// Helper methods for validation

func (s *FreeDataValidationService) validateBusinessName(ctx context.Context, data BusinessDataForValidation, result *ValidationResult) error {
	// This would use the government APIs to validate business name
	// For now, we'll implement a basic validation
	if len(data.Name) < 2 {
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Field:    "name",
			Message:  "Business name is too short",
			Severity: "high",
			Source:   "accuracy_validation",
			Code:     "NAME_TOO_SHORT",
		})
		return fmt.Errorf("business name too short")
	}
	return nil
}

func (s *FreeDataValidationService) validateRegistrationNumber(ctx context.Context, data BusinessDataForValidation, result *ValidationResult) error {
	// This would use the government APIs to validate registration number
	// For now, we'll implement basic format validation
	if len(data.RegistrationNumber) < 5 {
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Field:    "registration_number",
			Message:  "Registration number format appears invalid",
			Severity: "medium",
			Source:   "accuracy_validation",
			Code:     "INVALID_REGISTRATION_FORMAT",
		})
		return fmt.Errorf("invalid registration number format")
	}
	return nil
}

func (s *FreeDataValidationService) validateWebsite(ctx context.Context, data BusinessDataForValidation, result *ValidationResult) error {
	// This would use WHOIS and other free tools to validate website
	// For now, we'll implement basic URL validation
	if !strings.HasPrefix(data.Website, "http://") && !strings.HasPrefix(data.Website, "https://") {
		result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
			Field:    "website",
			Message:  "Website URL should start with http:// or https://",
			Severity: "low",
			Source:   "accuracy_validation",
			Code:     "INVALID_URL_FORMAT",
		})
		return fmt.Errorf("invalid URL format")
	}
	return nil
}

func (s *FreeDataValidationService) crossReferenceSECEdgar(ctx context.Context, data BusinessDataForValidation) (interface{}, error) {
	// This would use the SEC EDGAR API to cross-reference data
	// For now, return mock data
	return map[string]interface{}{
		"cik":   "0001234567",
		"name":  data.Name,
		"sic":   "5812",
		"state": "DE",
	}, nil
}

func (s *FreeDataValidationService) crossReferenceCompaniesHouse(ctx context.Context, data BusinessDataForValidation) (interface{}, error) {
	// This would use the Companies House API to cross-reference data
	// For now, return mock data
	return map[string]interface{}{
		"company_number": "12345678",
		"name":           data.Name,
		"status":         "active",
		"country":        "England",
	}, nil
}

func (s *FreeDataValidationService) crossReferenceOpenCorporates(ctx context.Context, data BusinessDataForValidation) (interface{}, error) {
	// This would use the OpenCorporates API to cross-reference data
	// For now, return mock data
	return map[string]interface{}{
		"company_number": "12345678",
		"name":           data.Name,
		"status":         "active",
		"jurisdiction":   data.Country,
	}, nil
}

func (s *FreeDataValidationService) isDataConsistent(data BusinessDataForValidation, referenceData interface{}, source string) bool {
	// This would implement logic to check if data is consistent with reference data
	// For now, return true as a placeholder
	return true
}

func (s *FreeDataValidationService) isValidAddressFormat(address string) bool {
	// Basic address format validation
	return len(address) > 10 && strings.Contains(address, " ")
}

func (s *FreeDataValidationService) isEmailDomainConsistent(email, website string) bool {
	// Extract domain from email and website
	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return false
	}
	emailDomain := strings.ToLower(emailParts[1])

	websiteDomain := strings.ToLower(website)
	websiteDomain = strings.TrimPrefix(websiteDomain, "http://")
	websiteDomain = strings.TrimPrefix(websiteDomain, "https://")
	websiteDomain = strings.TrimPrefix(websiteDomain, "www.")
	websiteDomain = strings.Split(websiteDomain, "/")[0]

	return emailDomain == websiteDomain
}

func (s *FreeDataValidationService) isValidPhoneFormat(phone string) bool {
	// Basic phone format validation
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// Handle international format with + prefix
	if strings.HasPrefix(phone, "+") {
		phone = phone[1:] // Remove + prefix
	}

	// Check if it's all digits (after removing common separators)
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	return len(phone) >= 10 && len(phone) <= 15
}

func (s *FreeDataValidationService) isNameDescriptionConsistent(name, description string) bool {
	// Basic consistency check between name and description
	nameWords := strings.Fields(strings.ToLower(name))
	descriptionWords := strings.Fields(strings.ToLower(description))

	// Check if any significant words from name appear in description
	for _, nameWord := range nameWords {
		if len(nameWord) > 3 { // Only check words longer than 3 characters
			for _, descWord := range descriptionWords {
				if nameWord == descWord {
					return true
				}
			}
		}
	}
	return false
}

func (s *FreeDataValidationService) isAddressCountryConsistent(address, country string) bool {
	// Basic address-country consistency check
	addressLower := strings.ToLower(address)
	countryLower := strings.ToLower(country)

	// Check if country name appears in address
	return strings.Contains(addressLower, countryLower)
}

// Cache management methods

func (s *FreeDataValidationService) getCachedValidation(businessID string) *ValidationResult {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if result, exists := s.validationCache[businessID]; exists {
		// Check if cache is still valid (e.g., not older than 1 hour)
		if time.Since(result.ValidatedAt) < time.Hour {
			return result
		}
		// Remove expired cache entry
		delete(s.validationCache, businessID)
	}
	return nil
}

func (s *FreeDataValidationService) cacheValidationResult(businessID string, result *ValidationResult) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.validationCache[businessID] = result
}

// GetValidationStats returns statistics about validation performance
func (s *FreeDataValidationService) GetValidationStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	totalValidations := len(s.validationCache)
	validCount := 0
	totalQualityScore := 0.0

	for _, result := range s.validationCache {
		if result.IsValid {
			validCount++
		}
		totalQualityScore += result.QualityScore
	}

	avgQualityScore := 0.0
	if totalValidations > 0 {
		avgQualityScore = totalQualityScore / float64(totalValidations)
	}

	return map[string]interface{}{
		"total_validations":     totalValidations,
		"valid_count":           validCount,
		"invalid_count":         totalValidations - validCount,
		"average_quality_score": avgQualityScore,
		"cache_size":            len(s.validationCache),
		"cost_per_validation":   0.0, // Always free
	}
}
