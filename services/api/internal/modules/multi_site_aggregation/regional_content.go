package multi_site_aggregation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RegionalContentService handles detection and processing of regional and localized content
type RegionalContentService interface {
	// DetectRegionalContent detects regional-specific content from extracted data
	DetectRegionalContent(ctx context.Context, siteData *SiteData, location *BusinessLocation) (*RegionalContent, error)

	// ExtractLocalizedData extracts localized data based on region and language
	ExtractLocalizedData(ctx context.Context, content string, region string, language string) (*LocalizedData, error)

	// NormalizeRegionalData normalizes regional data to a standard format
	NormalizeRegionalData(ctx context.Context, regionalData []*RegionalContent) (*NormalizedRegionalData, error)

	// GetSupportedRegions returns list of supported regions
	GetSupportedRegions() []RegionInfo

	// GetSupportedLanguages returns list of supported languages
	GetSupportedLanguages() []LanguageInfo
}

// RegionalContent represents region-specific content extracted from a business site
type RegionalContent struct {
	ID                 string                 `json:"id"`
	BusinessID         string                 `json:"business_id"`
	LocationID         string                 `json:"location_id"`
	Region             string                 `json:"region"`
	Country            string                 `json:"country"`
	Language           string                 `json:"language"`
	LocaleCode         string                 `json:"locale_code"` // e.g., "en-US", "fr-CA"
	ContentType        string                 `json:"content_type"`
	ExtractedContent   map[string]interface{} `json:"extracted_content"`
	LocalizedFields    map[string]interface{} `json:"localized_fields"`
	RegionalIndicators []RegionalIndicator    `json:"regional_indicators"`
	CurrencyInfo       *CurrencyInfo          `json:"currency_info,omitempty"`
	ContactInfo        *RegionalContactInfo   `json:"contact_info,omitempty"`
	BusinessHours      *RegionalBusinessHours `json:"business_hours,omitempty"`
	LegalCompliance    *LegalComplianceInfo   `json:"legal_compliance,omitempty"`
	ConfidenceScore    float64                `json:"confidence_score"`
	ExtractionMethod   string                 `json:"extraction_method"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// LocalizedData represents data that has been localized for a specific region/language
type LocalizedData struct {
	OriginalText       string                 `json:"original_text"`
	LocalizedText      string                 `json:"localized_text"`
	SourceLanguage     string                 `json:"source_language"`
	TargetLanguage     string                 `json:"target_language"`
	LocalizationType   string                 `json:"localization_type"` // "translation", "adaptation", "formatting"
	LocalizedFields    map[string]interface{} `json:"localized_fields"`
	QualityScore       float64                `json:"quality_score"`
	LocalizationMethod string                 `json:"localization_method"`
	ProcessedAt        time.Time              `json:"processed_at"`
}

// NormalizedRegionalData represents regional data normalized across different regions
type NormalizedRegionalData struct {
	BusinessID           string                            `json:"business_id"`
	PrimaryRegion        string                            `json:"primary_region"`
	SupportedRegions     []string                          `json:"supported_regions"`
	SupportedLanguages   []string                          `json:"supported_languages"`
	NormalizedContent    map[string]interface{}            `json:"normalized_content"`
	RegionalVariations   map[string]*RegionalContent       `json:"regional_variations"`
	CommonFields         map[string]interface{}            `json:"common_fields"`
	RegionSpecificFields map[string]map[string]interface{} `json:"region_specific_fields"`
	CurrencyMappings     map[string]*CurrencyInfo          `json:"currency_mappings"`
	ContactMappings      map[string]*RegionalContactInfo   `json:"contact_mappings"`
	ComplianceInfo       map[string]*LegalComplianceInfo   `json:"compliance_info"`
	NormalizationScore   float64                           `json:"normalization_score"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

// RegionalIndicator represents indicators that suggest regional content
type RegionalIndicator struct {
	Type        string  `json:"type"`        // "currency", "phone", "address", "language", "legal"
	Value       string  `json:"value"`       // The actual indicator value
	Confidence  float64 `json:"confidence"`  // Confidence in the indicator
	Region      string  `json:"region"`      // Suggested region
	Description string  `json:"description"` // Human-readable description
}

// CurrencyInfo represents currency-related information
type CurrencyInfo struct {
	CurrencyCode   string  `json:"currency_code"`   // e.g., "USD", "EUR"
	CurrencySymbol string  `json:"currency_symbol"` // e.g., "$", "€"
	Format         string  `json:"format"`          // e.g., "$1,234.56"
	DecimalPlaces  int     `json:"decimal_places"`
	Region         string  `json:"region"`
	Confidence     float64 `json:"confidence"`
}

// RegionalContactInfo represents region-specific contact information
type RegionalContactInfo struct {
	PhoneNumbers    []RegionalPhone      `json:"phone_numbers,omitempty"`
	Addresses       []RegionalAddress    `json:"addresses,omitempty"`
	EmailAddresses  []RegionalEmail      `json:"email_addresses,omitempty"`
	SocialMedia     []RegionalSocial     `json:"social_media,omitempty"`
	CustomerService *CustomerServiceInfo `json:"customer_service,omitempty"`
}

// RegionalPhone represents region-specific phone number information
type RegionalPhone struct {
	Number      string  `json:"number"`
	CountryCode string  `json:"country_code"`
	Type        string  `json:"type"` // "main", "support", "sales", "fax"
	Region      string  `json:"region"`
	IsLocal     bool    `json:"is_local"`
	IsTollFree  bool    `json:"is_toll_free"`
	Confidence  float64 `json:"confidence"`
}

// RegionalAddress represents region-specific address information
type RegionalAddress struct {
	FullAddress    string  `json:"full_address"`
	Street         string  `json:"street,omitempty"`
	City           string  `json:"city,omitempty"`
	State          string  `json:"state,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Country        string  `json:"country"`
	Region         string  `json:"region"`
	AddressType    string  `json:"address_type"` // "headquarters", "branch", "mailing"
	IsHeadquarters bool    `json:"is_headquarters"`
	Confidence     float64 `json:"confidence"`
}

// RegionalEmail represents region-specific email information
type RegionalEmail struct {
	Email      string  `json:"email"`
	Type       string  `json:"type"` // "general", "support", "sales", "info"
	Region     string  `json:"region"`
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// RegionalSocial represents region-specific social media information
type RegionalSocial struct {
	Platform   string  `json:"platform"`
	Handle     string  `json:"handle"`
	URL        string  `json:"url"`
	Region     string  `json:"region"`
	Language   string  `json:"language"`
	Followers  int     `json:"followers,omitempty"`
	Confidence float64 `json:"confidence"`
}

// CustomerServiceInfo represents customer service information
type CustomerServiceInfo struct {
	Hours      map[string]string `json:"hours,omitempty"` // day -> hours
	Languages  []string          `json:"languages,omitempty"`
	Channels   []string          `json:"channels,omitempty"` // "phone", "email", "chat"
	Region     string            `json:"region"`
	Timezone   string            `json:"timezone,omitempty"`
	Confidence float64           `json:"confidence"`
}

// RegionalBusinessHours represents region-specific business hours
type RegionalBusinessHours struct {
	Hours      map[string]string `json:"hours"`    // day -> hours (e.g., "monday" -> "9:00-17:00")
	Timezone   string            `json:"timezone"` // e.g., "America/New_York"
	Region     string            `json:"region"`
	IsLocal    bool              `json:"is_local"`
	Holidays   []string          `json:"holidays,omitempty"` // Regional holidays
	Confidence float64           `json:"confidence"`
}

// LegalComplianceInfo represents region-specific legal compliance information
type LegalComplianceInfo struct {
	Regulations     []string            `json:"regulations,omitempty"` // e.g., ["GDPR", "CCPA"]
	Licenses        []LicenseInfo       `json:"licenses,omitempty"`
	TaxInfo         *TaxInfo            `json:"tax_info,omitempty"`
	PrivacyPolicy   *PolicyInfo         `json:"privacy_policy,omitempty"`
	TermsOfService  *PolicyInfo         `json:"terms_of_service,omitempty"`
	Certifications  []CertificationInfo `json:"certifications,omitempty"`
	Region          string              `json:"region"`
	ComplianceScore float64             `json:"compliance_score"`
	LastUpdated     time.Time           `json:"last_updated"`
}

// LicenseInfo represents business license information
type LicenseInfo struct {
	LicenseNumber string    `json:"license_number"`
	LicenseType   string    `json:"license_type"`
	IssuingBody   string    `json:"issuing_body"`
	ValidFrom     time.Time `json:"valid_from,omitempty"`
	ValidTo       time.Time `json:"valid_to,omitempty"`
	Status        string    `json:"status"`
	Region        string    `json:"region"`
}

// TaxInfo represents tax-related information
type TaxInfo struct {
	TaxID       string `json:"tax_id,omitempty"`
	VATNumber   string `json:"vat_number,omitempty"`
	TaxRegion   string `json:"tax_region"`
	TaxCategory string `json:"tax_category,omitempty"`
}

// PolicyInfo represents policy document information
type PolicyInfo struct {
	URL         string    `json:"url,omitempty"`
	Language    string    `json:"language"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
	Version     string    `json:"version,omitempty"`
	Scope       string    `json:"scope,omitempty"` // "global", "regional"
}

// CertificationInfo represents certification information
type CertificationInfo struct {
	Name        string    `json:"name"`
	IssuingBody string    `json:"issuing_body"`
	Number      string    `json:"number,omitempty"`
	ValidFrom   time.Time `json:"valid_from,omitempty"`
	ValidTo     time.Time `json:"valid_to,omitempty"`
	Status      string    `json:"status"`
	Region      string    `json:"region"`
}

// RegionInfo represents information about a supported region
type RegionInfo struct {
	Code            string   `json:"code"`             // e.g., "US", "EU"
	Name            string   `json:"name"`             // e.g., "United States"
	Countries       []string `json:"countries"`        // Countries in this region
	DefaultLanguage string   `json:"default_language"` // Default language code
	CurrencyCode    string   `json:"currency_code"`    // Default currency
	TimeZones       []string `json:"time_zones"`       // Common timezones
	PhonePrefix     string   `json:"phone_prefix"`     // Country phone prefix
	PostalFormat    string   `json:"postal_format"`    // Postal code format regex
	DateFormat      string   `json:"date_format"`      // Common date format
	NumberFormat    string   `json:"number_format"`    // Number formatting pattern
}

// LanguageInfo represents information about a supported language
type LanguageInfo struct {
	Code         string   `json:"code"`          // e.g., "en", "fr"
	Name         string   `json:"name"`          // e.g., "English"
	NativeName   string   `json:"native_name"`   // e.g., "English", "Français"
	Regions      []string `json:"regions"`       // Regions where this language is common
	Direction    string   `json:"direction"`     // "ltr" or "rtl"
	Encoding     string   `json:"encoding"`      // Default encoding
	DateFormats  []string `json:"date_formats"`  // Common date formats
	TimeFormats  []string `json:"time_formats"`  // Common time formats
	NumberSystem string   `json:"number_system"` // Number system used
}

// RegionalContentConfig holds configuration for regional content processing
type RegionalContentConfig struct {
	EnableRegionalDetection    bool          `json:"enable_regional_detection"`
	EnableContentLocalization  bool          `json:"enable_content_localization"`
	EnableCurrencyDetection    bool          `json:"enable_currency_detection"`
	EnablePhoneNumberDetection bool          `json:"enable_phone_number_detection"`
	EnableAddressDetection     bool          `json:"enable_address_detection"`
	EnableLegalComplianceCheck bool          `json:"enable_legal_compliance_check"`
	DefaultRegion              string        `json:"default_region"`
	DefaultLanguage            string        `json:"default_language"`
	SupportedRegions           []string      `json:"supported_regions"`
	SupportedLanguages         []string      `json:"supported_languages"`
	RegionalDetectionTimeout   time.Duration `json:"regional_detection_timeout"`
	LocalizationTimeout        time.Duration `json:"localization_timeout"`
	MinConfidenceScore         float64       `json:"min_confidence_score"`
	EnableContentTranslation   bool          `json:"enable_content_translation"`
	CacheRegionalData          bool          `json:"cache_regional_data"`
}

// DefaultRegionalContentConfig returns default configuration for regional content
func DefaultRegionalContentConfig() *RegionalContentConfig {
	return &RegionalContentConfig{
		EnableRegionalDetection:    true,
		EnableContentLocalization:  true,
		EnableCurrencyDetection:    true,
		EnablePhoneNumberDetection: true,
		EnableAddressDetection:     true,
		EnableLegalComplianceCheck: false, // Disabled by default for performance
		DefaultRegion:              "US",
		DefaultLanguage:            "en",
		SupportedRegions:           []string{"US", "CA", "GB", "EU", "AU", "JP", "CN", "IN", "BR", "MX"},
		SupportedLanguages:         []string{"en", "es", "fr", "de", "it", "pt", "ja", "zh", "ko", "hi"},
		RegionalDetectionTimeout:   15 * time.Second,
		LocalizationTimeout:        10 * time.Second,
		MinConfidenceScore:         0.6,
		EnableContentTranslation:   false, // Disabled by default to avoid translation API costs
		CacheRegionalData:          true,
	}
}

// WebsiteRegionalContentService implements RegionalContentService interface
type WebsiteRegionalContentService struct {
	config           *RegionalContentConfig
	logger           *zap.Logger
	regionInfoMap    map[string]*RegionInfo
	languageInfoMap  map[string]*LanguageInfo
	currencyPatterns map[string]*regexp.Regexp
	phonePatterns    map[string]*regexp.Regexp
	addressPatterns  map[string]*regexp.Regexp
	datePatterns     map[string]*regexp.Regexp
}

// NewWebsiteRegionalContentService creates a new regional content service
func NewWebsiteRegionalContentService(
	config *RegionalContentConfig,
	logger *zap.Logger,
) *WebsiteRegionalContentService {
	if config == nil {
		config = DefaultRegionalContentConfig()
	}

	service := &WebsiteRegionalContentService{
		config:           config,
		logger:           logger,
		regionInfoMap:    make(map[string]*RegionInfo),
		languageInfoMap:  make(map[string]*LanguageInfo),
		currencyPatterns: make(map[string]*regexp.Regexp),
		phonePatterns:    make(map[string]*regexp.Regexp),
		addressPatterns:  make(map[string]*regexp.Regexp),
		datePatterns:     make(map[string]*regexp.Regexp),
	}

	// Initialize patterns and maps
	service.initializeRegionInfo()
	service.initializeLanguageInfo()
	service.initializePatterns()

	return service
}

// DetectRegionalContent detects regional-specific content from extracted data
func (s *WebsiteRegionalContentService) DetectRegionalContent(
	ctx context.Context,
	siteData *SiteData,
	location *BusinessLocation,
) (*RegionalContent, error) {
	if siteData == nil || location == nil {
		return nil, fmt.Errorf("site data and location are required")
	}

	s.logger.Info("detecting regional content",
		zap.String("location_id", location.ID),
		zap.String("region", location.Region))

	// Create regional content structure
	regionalContent := &RegionalContent{
		ID:               generateID(),
		BusinessID:       siteData.BusinessID,
		LocationID:       siteData.LocationID,
		Region:           location.Region,
		Country:          location.Country,
		Language:         location.Language,
		LocaleCode:       s.buildLocaleCode(location.Language, location.Country),
		ContentType:      siteData.DataType,
		ExtractedContent: make(map[string]interface{}),
		LocalizedFields:  make(map[string]interface{}),
		Metadata:         make(map[string]interface{}),
		ExtractionMethod: "regional_content_detection",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Detect regional indicators
	indicators, err := s.detectRegionalIndicators(ctx, siteData.ExtractedData, location)
	if err != nil {
		s.logger.Warn("failed to detect regional indicators",
			zap.Error(err))
		// Continue with empty indicators
		indicators = []RegionalIndicator{}
	}
	regionalContent.RegionalIndicators = indicators

	// Extract currency information
	if s.config.EnableCurrencyDetection {
		currencyInfo := s.extractCurrencyInfo(siteData.ExtractedData, location.Region)
		regionalContent.CurrencyInfo = currencyInfo
	}

	// Extract regional contact information
	contactInfo := s.extractRegionalContactInfo(siteData.ExtractedData, location)
	regionalContent.ContactInfo = contactInfo

	// Extract business hours
	businessHours := s.extractRegionalBusinessHours(siteData.ExtractedData, location)
	regionalContent.BusinessHours = businessHours

	// Extract legal compliance information
	if s.config.EnableLegalComplianceCheck {
		legalInfo := s.extractLegalComplianceInfo(siteData.ExtractedData, location)
		regionalContent.LegalCompliance = legalInfo
	}

	// Process localized content
	if s.config.EnableContentLocalization {
		localizedFields, err := s.localizeContent(ctx, siteData.ExtractedData, location)
		if err != nil {
			s.logger.Warn("failed to localize content",
				zap.Error(err))
		} else {
			regionalContent.LocalizedFields = localizedFields
		}
	}

	// Copy relevant extracted data
	regionalContent.ExtractedContent = s.filterExtractedData(siteData.ExtractedData)

	// Calculate confidence score
	regionalContent.ConfidenceScore = s.calculateRegionalConfidence(regionalContent)

	s.logger.Info("regional content detection completed",
		zap.String("location_id", location.ID),
		zap.Float64("confidence_score", regionalContent.ConfidenceScore),
		zap.Int("indicators_count", len(indicators)))

	return regionalContent, nil
}

// ExtractLocalizedData extracts localized data based on region and language
func (s *WebsiteRegionalContentService) ExtractLocalizedData(
	ctx context.Context,
	content string,
	region string,
	language string,
) (*LocalizedData, error) {
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	s.logger.Debug("extracting localized data",
		zap.String("region", region),
		zap.String("language", language),
		zap.Int("content_length", len(content)))

	localizedData := &LocalizedData{
		OriginalText:       content,
		LocalizedText:      content, // Default to original
		SourceLanguage:     s.detectLanguage(content),
		TargetLanguage:     language,
		LocalizationType:   "adaptation",
		LocalizedFields:    make(map[string]interface{}),
		LocalizationMethod: "rule_based",
		ProcessedAt:        time.Now(),
	}

	// Apply regional formatting and adaptations
	localizedText, localizedFields := s.applyRegionalFormatting(content, region, language)
	localizedData.LocalizedText = localizedText
	localizedData.LocalizedFields = localizedFields

	// Calculate quality score
	localizedData.QualityScore = s.calculateLocalizationQuality(localizedData)

	return localizedData, nil
}

// NormalizeRegionalData normalizes regional data to a standard format
func (s *WebsiteRegionalContentService) NormalizeRegionalData(
	ctx context.Context,
	regionalData []*RegionalContent,
) (*NormalizedRegionalData, error) {
	if len(regionalData) == 0 {
		return nil, fmt.Errorf("no regional data to normalize")
	}

	businessID := regionalData[0].BusinessID
	s.logger.Info("normalizing regional data",
		zap.String("business_id", businessID),
		zap.Int("regions_count", len(regionalData)))

	normalized := &NormalizedRegionalData{
		BusinessID:           businessID,
		NormalizedContent:    make(map[string]interface{}),
		RegionalVariations:   make(map[string]*RegionalContent),
		CommonFields:         make(map[string]interface{}),
		RegionSpecificFields: make(map[string]map[string]interface{}),
		CurrencyMappings:     make(map[string]*CurrencyInfo),
		ContactMappings:      make(map[string]*RegionalContactInfo),
		ComplianceInfo:       make(map[string]*LegalComplianceInfo),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Collect regions and languages
	regions := make(map[string]bool)
	languages := make(map[string]bool)
	var primaryRegion string
	highestConfidence := 0.0

	for _, data := range regionalData {
		regions[data.Region] = true
		languages[data.Language] = true
		normalized.RegionalVariations[data.Region] = data

		// Track primary region (highest confidence)
		if data.ConfidenceScore > highestConfidence {
			highestConfidence = data.ConfidenceScore
			primaryRegion = data.Region
		}

		// Map currency info
		if data.CurrencyInfo != nil {
			normalized.CurrencyMappings[data.Region] = data.CurrencyInfo
		}

		// Map contact info
		if data.ContactInfo != nil {
			normalized.ContactMappings[data.Region] = data.ContactInfo
		}

		// Map compliance info
		if data.LegalCompliance != nil {
			normalized.ComplianceInfo[data.Region] = data.LegalCompliance
		}
	}

	// Set primary region and supported lists
	normalized.PrimaryRegion = primaryRegion
	for region := range regions {
		normalized.SupportedRegions = append(normalized.SupportedRegions, region)
	}
	for language := range languages {
		normalized.SupportedLanguages = append(normalized.SupportedLanguages, language)
	}

	// Extract common fields across all regions
	normalized.CommonFields = s.extractCommonFields(regionalData)

	// Extract region-specific fields
	normalized.RegionSpecificFields = s.extractRegionSpecificFields(regionalData)

	// Create normalized content (using primary region as base)
	if primaryRegion != "" && normalized.RegionalVariations[primaryRegion] != nil {
		primaryData := normalized.RegionalVariations[primaryRegion]
		normalized.NormalizedContent = s.mergeContent(primaryData.ExtractedContent, normalized.CommonFields)
	}

	// Calculate normalization score
	normalized.NormalizationScore = s.calculateNormalizationScore(normalized)

	s.logger.Info("regional data normalization completed",
		zap.String("business_id", businessID),
		zap.String("primary_region", primaryRegion),
		zap.Int("supported_regions", len(normalized.SupportedRegions)),
		zap.Float64("normalization_score", normalized.NormalizationScore))

	return normalized, nil
}

// GetSupportedRegions returns list of supported regions
func (s *WebsiteRegionalContentService) GetSupportedRegions() []RegionInfo {
	var regions []RegionInfo
	for _, region := range s.regionInfoMap {
		regions = append(regions, *region)
	}
	return regions
}

// GetSupportedLanguages returns list of supported languages
func (s *WebsiteRegionalContentService) GetSupportedLanguages() []LanguageInfo {
	var languages []LanguageInfo
	for _, language := range s.languageInfoMap {
		languages = append(languages, *language)
	}
	return languages
}

// Helper methods

func (s *WebsiteRegionalContentService) initializeRegionInfo() {
	// Initialize common region information
	regions := []*RegionInfo{
		{
			Code:            "US",
			Name:            "United States",
			Countries:       []string{"US"},
			DefaultLanguage: "en",
			CurrencyCode:    "USD",
			TimeZones:       []string{"America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles"},
			PhonePrefix:     "+1",
			PostalFormat:    `\d{5}(-\d{4})?`,
			DateFormat:      "MM/dd/yyyy",
			NumberFormat:    "1,234.56",
		},
		{
			Code:            "CA",
			Name:            "Canada",
			Countries:       []string{"CA"},
			DefaultLanguage: "en",
			CurrencyCode:    "CAD",
			TimeZones:       []string{"America/Toronto", "America/Vancouver", "America/Edmonton"},
			PhonePrefix:     "+1",
			PostalFormat:    `[A-Z]\d[A-Z] \d[A-Z]\d`,
			DateFormat:      "dd/MM/yyyy",
			NumberFormat:    "1,234.56",
		},
		{
			Code:            "GB",
			Name:            "United Kingdom",
			Countries:       []string{"GB", "UK"},
			DefaultLanguage: "en",
			CurrencyCode:    "GBP",
			TimeZones:       []string{"Europe/London"},
			PhonePrefix:     "+44",
			PostalFormat:    `[A-Z]{1,2}\d[A-Z\d]? \d[A-Z]{2}`,
			DateFormat:      "dd/MM/yyyy",
			NumberFormat:    "1,234.56",
		},
		{
			Code:            "EU",
			Name:            "European Union",
			Countries:       []string{"DE", "FR", "IT", "ES", "NL", "BE", "AT", "FI", "PT", "IE", "GR", "LU"},
			DefaultLanguage: "en",
			CurrencyCode:    "EUR",
			TimeZones:       []string{"Europe/Berlin", "Europe/Paris", "Europe/Rome"},
			PhonePrefix:     "+49", // Default to Germany
			PostalFormat:    `\d{5}`,
			DateFormat:      "dd.MM.yyyy",
			NumberFormat:    "1.234,56",
		},
		{
			Code:            "AU",
			Name:            "Australia",
			Countries:       []string{"AU"},
			DefaultLanguage: "en",
			CurrencyCode:    "AUD",
			TimeZones:       []string{"Australia/Sydney", "Australia/Melbourne", "Australia/Perth"},
			PhonePrefix:     "+61",
			PostalFormat:    `\d{4}`,
			DateFormat:      "dd/MM/yyyy",
			NumberFormat:    "1,234.56",
		},
	}

	for _, region := range regions {
		s.regionInfoMap[region.Code] = region
	}
}

func (s *WebsiteRegionalContentService) initializeLanguageInfo() {
	// Initialize common language information
	languages := []*LanguageInfo{
		{
			Code:         "en",
			Name:         "English",
			NativeName:   "English",
			Regions:      []string{"US", "CA", "GB", "AU"},
			Direction:    "ltr",
			Encoding:     "UTF-8",
			DateFormats:  []string{"MM/dd/yyyy", "dd/MM/yyyy", "yyyy-MM-dd"},
			TimeFormats:  []string{"HH:mm", "hh:mm a"},
			NumberSystem: "decimal",
		},
		{
			Code:         "es",
			Name:         "Spanish",
			NativeName:   "Español",
			Regions:      []string{"ES", "MX", "AR", "CO"},
			Direction:    "ltr",
			Encoding:     "UTF-8",
			DateFormats:  []string{"dd/MM/yyyy", "dd-MM-yyyy"},
			TimeFormats:  []string{"HH:mm"},
			NumberSystem: "decimal",
		},
		{
			Code:         "fr",
			Name:         "French",
			NativeName:   "Français",
			Regions:      []string{"FR", "CA", "BE", "CH"},
			Direction:    "ltr",
			Encoding:     "UTF-8",
			DateFormats:  []string{"dd/MM/yyyy", "dd.MM.yyyy"},
			TimeFormats:  []string{"HH:mm"},
			NumberSystem: "decimal",
		},
		{
			Code:         "de",
			Name:         "German",
			NativeName:   "Deutsch",
			Regions:      []string{"DE", "AT", "CH"},
			Direction:    "ltr",
			Encoding:     "UTF-8",
			DateFormats:  []string{"dd.MM.yyyy", "dd-MM-yyyy"},
			TimeFormats:  []string{"HH:mm"},
			NumberSystem: "decimal",
		},
	}

	for _, language := range languages {
		s.languageInfoMap[language.Code] = language
	}
}

func (s *WebsiteRegionalContentService) initializePatterns() {
	// Initialize currency patterns
	s.currencyPatterns["USD"] = regexp.MustCompile(`\$\s*\d{1,3}(,\d{3})*(\.\d{2})?`)
	s.currencyPatterns["EUR"] = regexp.MustCompile(`€\s*\d{1,3}(\.\d{3})*(,\d{2})?`)
	s.currencyPatterns["GBP"] = regexp.MustCompile(`£\s*\d{1,3}(,\d{3})*(\.\d{2})?`)
	s.currencyPatterns["CAD"] = regexp.MustCompile(`C\$\s*\d{1,3}(,\d{3})*(\.\d{2})?`)

	// Initialize phone patterns
	s.phonePatterns["US"] = regexp.MustCompile(`\+?1?[-.\s]?\(?(\d{3})\)?[-.\s]?(\d{3})[-.\s]?(\d{4})`)
	s.phonePatterns["GB"] = regexp.MustCompile(`\+?44[-.\s]?\d{4}[-.\s]?\d{3}[-.\s]?\d{3}`)
	s.phonePatterns["DE"] = regexp.MustCompile(`\+?49[-.\s]?\d{3,4}[-.\s]?\d{3,8}`)
	s.phonePatterns["FR"] = regexp.MustCompile(`\+?33[-.\s]?\d[-.\s]?\d{2}[-.\s]?\d{2}[-.\s]?\d{2}[-.\s]?\d{2}`)

	// Initialize address patterns
	s.addressPatterns["US"] = regexp.MustCompile(`\d+\s+[\w\s]+,\s*[\w\s]+,\s*[A-Z]{2}\s+\d{5}(-\d{4})?`)
	s.addressPatterns["GB"] = regexp.MustCompile(`[\w\s]+,\s*[\w\s]+,\s*[A-Z]{1,2}\d[A-Z\d]?\s+\d[A-Z]{2}`)
	s.addressPatterns["DE"] = regexp.MustCompile(`[\w\s]+\s+\d+,\s*\d{5}\s+[\w\s]+`)

	// Initialize date patterns
	s.datePatterns["US"] = regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4}`)
	s.datePatterns["EU"] = regexp.MustCompile(`\d{1,2}\.\d{1,2}\.\d{4}`)
	s.datePatterns["GB"] = regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4}`)
}

func (s *WebsiteRegionalContentService) buildLocaleCode(language, country string) string {
	if language == "" {
		language = s.config.DefaultLanguage
	}
	if country == "" {
		country = s.config.DefaultRegion
	}
	return fmt.Sprintf("%s-%s", strings.ToLower(language), strings.ToUpper(country))
}

func (s *WebsiteRegionalContentService) detectRegionalIndicators(
	ctx context.Context,
	extractedData map[string]interface{},
	location *BusinessLocation,
) ([]RegionalIndicator, error) {
	var indicators []RegionalIndicator

	// Convert extracted data to searchable text
	searchText := s.extractTextFromData(extractedData)

	// Detect currency indicators
	if s.config.EnableCurrencyDetection {
		currencyIndicators := s.detectCurrencyIndicators(searchText, location.Region)
		indicators = append(indicators, currencyIndicators...)
	}

	// Detect phone number indicators
	if s.config.EnablePhoneNumberDetection {
		phoneIndicators := s.detectPhoneIndicators(searchText, location.Region)
		indicators = append(indicators, phoneIndicators...)
	}

	// Detect address indicators
	if s.config.EnableAddressDetection {
		addressIndicators := s.detectAddressIndicators(searchText, location.Region)
		indicators = append(indicators, addressIndicators...)
	}

	// Detect language indicators
	languageIndicators := s.detectLanguageIndicators(searchText, location.Language)
	indicators = append(indicators, languageIndicators...)

	return indicators, nil
}

func (s *WebsiteRegionalContentService) detectCurrencyIndicators(text, region string) []RegionalIndicator {
	var indicators []RegionalIndicator

	for currency, pattern := range s.currencyPatterns {
		matches := pattern.FindAllString(text, -1)
		if len(matches) > 0 {
			confidence := s.calculateCurrencyConfidence(currency, region, len(matches))
			indicator := RegionalIndicator{
				Type:        "currency",
				Value:       currency,
				Confidence:  confidence,
				Region:      s.getCurrencyRegion(currency),
				Description: fmt.Sprintf("Found %d %s currency references", len(matches), currency),
			}
			indicators = append(indicators, indicator)
		}
	}

	return indicators
}

func (s *WebsiteRegionalContentService) detectPhoneIndicators(text, region string) []RegionalIndicator {
	var indicators []RegionalIndicator

	for phoneRegion, pattern := range s.phonePatterns {
		matches := pattern.FindAllString(text, -1)
		if len(matches) > 0 {
			confidence := s.calculatePhoneConfidence(phoneRegion, region, len(matches))
			indicator := RegionalIndicator{
				Type:        "phone",
				Value:       phoneRegion,
				Confidence:  confidence,
				Region:      phoneRegion,
				Description: fmt.Sprintf("Found %d phone numbers with %s format", len(matches), phoneRegion),
			}
			indicators = append(indicators, indicator)
		}
	}

	return indicators
}

func (s *WebsiteRegionalContentService) detectAddressIndicators(text, region string) []RegionalIndicator {
	var indicators []RegionalIndicator

	for addressRegion, pattern := range s.addressPatterns {
		matches := pattern.FindAllString(text, -1)
		if len(matches) > 0 {
			confidence := s.calculateAddressConfidence(addressRegion, region, len(matches))
			indicator := RegionalIndicator{
				Type:        "address",
				Value:       addressRegion,
				Confidence:  confidence,
				Region:      addressRegion,
				Description: fmt.Sprintf("Found %d addresses with %s format", len(matches), addressRegion),
			}
			indicators = append(indicators, indicator)
		}
	}

	return indicators
}

func (s *WebsiteRegionalContentService) detectLanguageIndicators(text, language string) []RegionalIndicator {
	var indicators []RegionalIndicator

	// Simple language detection based on common words
	detectedLanguage := s.detectLanguage(text)
	if detectedLanguage != "" {
		confidence := 0.7 // Base confidence
		if detectedLanguage == language {
			confidence = 0.9 // Higher if matches expected
		}

		indicator := RegionalIndicator{
			Type:        "language",
			Value:       detectedLanguage,
			Confidence:  confidence,
			Region:      s.getLanguageRegion(detectedLanguage),
			Description: fmt.Sprintf("Detected language: %s", detectedLanguage),
		}
		indicators = append(indicators, indicator)
	}

	return indicators
}

func (s *WebsiteRegionalContentService) extractCurrencyInfo(extractedData map[string]interface{}, region string) *CurrencyInfo {
	text := s.extractTextFromData(extractedData)

	// Try to find currency information
	for currency, pattern := range s.currencyPatterns {
		matches := pattern.FindAllString(text, -1)
		if len(matches) > 0 {
			// Get currency symbol
			symbol := s.getCurrencySymbol(currency)

			// Calculate confidence based on region match and frequency
			confidence := s.calculateCurrencyConfidence(currency, region, len(matches))

			return &CurrencyInfo{
				CurrencyCode:   currency,
				CurrencySymbol: symbol,
				Format:         matches[0], // Use first match as example format
				DecimalPlaces:  s.getCurrencyDecimalPlaces(currency),
				Region:         region,
				Confidence:     confidence,
			}
		}
	}

	// Return default currency for region if no specific currency found
	if regionInfo, exists := s.regionInfoMap[region]; exists {
		return &CurrencyInfo{
			CurrencyCode:   regionInfo.CurrencyCode,
			CurrencySymbol: s.getCurrencySymbol(regionInfo.CurrencyCode),
			Format:         regionInfo.NumberFormat,
			DecimalPlaces:  2, // Default
			Region:         region,
			Confidence:     0.5, // Lower confidence for default
		}
	}

	return nil
}

func (s *WebsiteRegionalContentService) extractRegionalContactInfo(extractedData map[string]interface{}, location *BusinessLocation) *RegionalContactInfo {
	contactInfo := &RegionalContactInfo{}
	text := s.extractTextFromData(extractedData)

	// Extract phone numbers
	if pattern, exists := s.phonePatterns[location.Region]; exists {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			phone := RegionalPhone{
				Number:      match,
				CountryCode: s.getCountryCodeForRegion(location.Region),
				Type:        "main", // Default type
				Region:      location.Region,
				IsLocal:     true,
				IsTollFree:  s.isTollFree(match),
				Confidence:  0.8,
			}
			contactInfo.PhoneNumbers = append(contactInfo.PhoneNumbers, phone)
		}
	}

	// Extract addresses
	if pattern, exists := s.addressPatterns[location.Region]; exists {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			address := RegionalAddress{
				FullAddress:    match,
				Country:        location.Country,
				Region:         location.Region,
				AddressType:    "main",
				IsHeadquarters: len(contactInfo.Addresses) == 0, // First address is HQ
				Confidence:     0.8,
			}
			contactInfo.Addresses = append(contactInfo.Addresses, address)
		}
	}

	// Extract email addresses (basic pattern)
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emailMatches := emailPattern.FindAllString(text, -1)
	for _, match := range emailMatches {
		email := RegionalEmail{
			Email:      match,
			Type:       "general",
			Region:     location.Region,
			Language:   location.Language,
			Confidence: 0.9,
		}
		contactInfo.EmailAddresses = append(contactInfo.EmailAddresses, email)
	}

	return contactInfo
}

func (s *WebsiteRegionalContentService) extractRegionalBusinessHours(extractedData map[string]interface{}, location *BusinessLocation) *RegionalBusinessHours {
	text := s.extractTextFromData(extractedData)

	// Simple business hours extraction (this could be much more sophisticated)
	hoursPattern := regexp.MustCompile(`(?i)(monday|tuesday|wednesday|thursday|friday|saturday|sunday)[:\s-]+([\d:apm\s-]+)`)
	matches := hoursPattern.FindAllStringSubmatch(text, -1)

	hours := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			day := strings.ToLower(match[1])
			time := strings.TrimSpace(match[2])
			hours[day] = time
		}
	}

	if len(hours) == 0 {
		return nil
	}

	// Get timezone for region
	timezone := s.getTimezoneForRegion(location.Region)

	return &RegionalBusinessHours{
		Hours:      hours,
		Timezone:   timezone,
		Region:     location.Region,
		IsLocal:    true,
		Holidays:   s.getRegionalHolidays(location.Region),
		Confidence: 0.7,
	}
}

func (s *WebsiteRegionalContentService) extractLegalComplianceInfo(extractedData map[string]interface{}, location *BusinessLocation) *LegalComplianceInfo {
	text := s.extractTextFromData(extractedData)

	// Look for common compliance terms
	regulations := []string{}
	if strings.Contains(strings.ToLower(text), "gdpr") {
		regulations = append(regulations, "GDPR")
	}
	if strings.Contains(strings.ToLower(text), "ccpa") {
		regulations = append(regulations, "CCPA")
	}
	if strings.Contains(strings.ToLower(text), "hipaa") {
		regulations = append(regulations, "HIPAA")
	}

	// Look for privacy policy
	var privacyPolicy *PolicyInfo
	privacyPattern := regexp.MustCompile(`(?i)privacy\s+policy`)
	if privacyPattern.MatchString(text) {
		privacyPolicy = &PolicyInfo{
			Language:    location.Language,
			LastUpdated: time.Now(), // Would need better detection
			Scope:       "regional",
		}
	}

	// Look for terms of service
	var termsOfService *PolicyInfo
	termsPattern := regexp.MustCompile(`(?i)terms\s+(of\s+)?(service|use)`)
	if termsPattern.MatchString(text) {
		termsOfService = &PolicyInfo{
			Language:    location.Language,
			LastUpdated: time.Now(), // Would need better detection
			Scope:       "regional",
		}
	}

	if len(regulations) == 0 && privacyPolicy == nil && termsOfService == nil {
		return nil
	}

	complianceScore := float64(len(regulations)) * 0.3
	if privacyPolicy != nil {
		complianceScore += 0.2
	}
	if termsOfService != nil {
		complianceScore += 0.2
	}
	if complianceScore > 1.0 {
		complianceScore = 1.0
	}

	return &LegalComplianceInfo{
		Regulations:     regulations,
		PrivacyPolicy:   privacyPolicy,
		TermsOfService:  termsOfService,
		Region:          location.Region,
		ComplianceScore: complianceScore,
		LastUpdated:     time.Now(),
	}
}

func (s *WebsiteRegionalContentService) localizeContent(ctx context.Context, extractedData map[string]interface{}, location *BusinessLocation) (map[string]interface{}, error) {
	localizedFields := make(map[string]interface{})

	// Apply region-specific formatting to relevant fields
	for key, value := range extractedData {
		if valueStr, ok := value.(string); ok {
			localizedValue := s.applyRegionalFormattingToString(valueStr, location.Region, location.Language)
			if localizedValue != valueStr {
				localizedFields[key] = localizedValue
			}
		}
	}

	return localizedFields, nil
}

func (s *WebsiteRegionalContentService) filterExtractedData(extractedData map[string]interface{}) map[string]interface{} {
	// Filter and clean extracted data for regional processing
	filtered := make(map[string]interface{})
	for key, value := range extractedData {
		// Include all data for now, could add filtering logic here
		filtered[key] = value
	}
	return filtered
}

func (s *WebsiteRegionalContentService) calculateRegionalConfidence(content *RegionalContent) float64 {
	score := 0.0
	totalChecks := 0.0

	// Base score from indicators
	if len(content.RegionalIndicators) > 0 {
		indicatorScore := 0.0
		for _, indicator := range content.RegionalIndicators {
			indicatorScore += indicator.Confidence
		}
		score += indicatorScore / float64(len(content.RegionalIndicators)) * 0.3
		totalChecks += 0.3
	}

	// Currency information confidence
	if content.CurrencyInfo != nil {
		score += content.CurrencyInfo.Confidence * 0.2
		totalChecks += 0.2
	}

	// Contact information confidence
	if content.ContactInfo != nil {
		contactScore := 0.0
		count := 0
		if len(content.ContactInfo.PhoneNumbers) > 0 {
			for _, phone := range content.ContactInfo.PhoneNumbers {
				contactScore += phone.Confidence
				count++
			}
		}
		if len(content.ContactInfo.EmailAddresses) > 0 {
			for _, email := range content.ContactInfo.EmailAddresses {
				contactScore += email.Confidence
				count++
			}
		}
		if count > 0 {
			score += (contactScore / float64(count)) * 0.3
			totalChecks += 0.3
		}
	}

	// Business hours confidence
	if content.BusinessHours != nil {
		score += content.BusinessHours.Confidence * 0.1
		totalChecks += 0.1
	}

	// Legal compliance confidence
	if content.LegalCompliance != nil {
		score += content.LegalCompliance.ComplianceScore * 0.1
		totalChecks += 0.1
	}

	if totalChecks == 0 {
		return 0.5 // Default confidence
	}

	return score / totalChecks
}

func (s *WebsiteRegionalContentService) detectLanguage(text string) string {
	// Simple language detection based on common words
	text = strings.ToLower(text)

	// English indicators
	englishWords := []string{"the", "and", "or", "for", "with", "about", "contact", "home", "services"}
	englishCount := 0
	for _, word := range englishWords {
		if strings.Contains(text, word) {
			englishCount++
		}
	}

	// Spanish indicators
	spanishWords := []string{"el", "la", "y", "de", "con", "para", "contacto", "inicio", "servicios"}
	spanishCount := 0
	for _, word := range spanishWords {
		if strings.Contains(text, word) {
			spanishCount++
		}
	}

	// French indicators
	frenchWords := []string{"le", "la", "et", "de", "avec", "pour", "contact", "accueil", "services"}
	frenchCount := 0
	for _, word := range frenchWords {
		if strings.Contains(text, word) {
			frenchCount++
		}
	}

	// German indicators
	germanWords := []string{"der", "die", "das", "und", "mit", "für", "kontakt", "home", "dienste"}
	germanCount := 0
	for _, word := range germanWords {
		if strings.Contains(text, word) {
			germanCount++
		}
	}

	// Determine most likely language
	maxCount := englishCount
	language := "en"

	if spanishCount > maxCount {
		maxCount = spanishCount
		language = "es"
	}
	if frenchCount > maxCount {
		maxCount = frenchCount
		language = "fr"
	}
	if germanCount > maxCount {
		maxCount = germanCount
		language = "de"
	}

	// Require minimum confidence
	if maxCount < 2 {
		return ""
	}

	return language
}

func (s *WebsiteRegionalContentService) applyRegionalFormatting(content, region, language string) (string, map[string]interface{}) {
	localizedText := content
	localizedFields := make(map[string]interface{})

	// Apply region-specific formatting
	if regionInfo, exists := s.regionInfoMap[region]; exists {
		// Apply date formatting
		if s.datePatterns["US"] != nil {
			matches := s.datePatterns["US"].FindAllString(content, -1)
			for _, match := range matches {
				if regionInfo.DateFormat == "dd/MM/yyyy" && strings.Contains(match, "/") {
					// Convert MM/dd/yyyy to dd/MM/yyyy
					parts := strings.Split(match, "/")
					if len(parts) == 3 {
						converted := fmt.Sprintf("%s/%s/%s", parts[1], parts[0], parts[2])
						localizedText = strings.Replace(localizedText, match, converted, -1)
						localizedFields["date_format_applied"] = true
					}
				}
			}
		}

		// Apply number formatting
		numberPattern := regexp.MustCompile(`\d{1,3}(,\d{3})*(\.\d{2})?`)
		if regionInfo.NumberFormat == "1.234,56" {
			matches := numberPattern.FindAllString(content, -1)
			for _, match := range matches {
				if strings.Contains(match, ",") && strings.Contains(match, ".") {
					// Convert 1,234.56 to 1.234,56
					converted := strings.Replace(match, ",", "TEMP", -1)
					converted = strings.Replace(converted, ".", ",", -1)
					converted = strings.Replace(converted, "TEMP", ".", -1)
					localizedText = strings.Replace(localizedText, match, converted, -1)
					localizedFields["number_format_applied"] = true
				}
			}
		}
	}

	return localizedText, localizedFields
}

func (s *WebsiteRegionalContentService) applyRegionalFormattingToString(value, region, language string) string {
	// Apply basic regional formatting to a string value
	localizedValue, _ := s.applyRegionalFormatting(value, region, language)
	return localizedValue
}

func (s *WebsiteRegionalContentService) calculateLocalizationQuality(data *LocalizedData) float64 {
	score := 0.5 // Base score

	// Higher score if languages are different (actual localization occurred)
	if data.SourceLanguage != data.TargetLanguage {
		score += 0.3
	}

	// Higher score if localized fields were populated
	if len(data.LocalizedFields) > 0 {
		score += 0.2
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (s *WebsiteRegionalContentService) extractCommonFields(regionalData []*RegionalContent) map[string]interface{} {
	if len(regionalData) == 0 {
		return make(map[string]interface{})
	}

	// Find fields that exist in all regions
	commonFields := make(map[string]interface{})
	firstData := regionalData[0].ExtractedContent

	for key, value := range firstData {
		isCommon := true
		for i := 1; i < len(regionalData); i++ {
			if _, exists := regionalData[i].ExtractedContent[key]; !exists {
				isCommon = false
				break
			}
			// Check if values are similar (basic check)
			if fmt.Sprintf("%v", regionalData[i].ExtractedContent[key]) != fmt.Sprintf("%v", value) {
				isCommon = false
				break
			}
		}
		if isCommon {
			commonFields[key] = value
		}
	}

	return commonFields
}

func (s *WebsiteRegionalContentService) extractRegionSpecificFields(regionalData []*RegionalContent) map[string]map[string]interface{} {
	regionSpecific := make(map[string]map[string]interface{})

	for _, data := range regionalData {
		regionFields := make(map[string]interface{})

		// Add fields that are unique to this region or have different values
		for key, value := range data.ExtractedContent {
			isSpecific := false
			for _, otherData := range regionalData {
				if otherData.Region == data.Region {
					continue
				}
				if otherValue, exists := otherData.ExtractedContent[key]; exists {
					if fmt.Sprintf("%v", otherValue) != fmt.Sprintf("%v", value) {
						isSpecific = true
						break
					}
				} else {
					isSpecific = true
					break
				}
			}
			if isSpecific {
				regionFields[key] = value
			}
		}

		if len(regionFields) > 0 {
			regionSpecific[data.Region] = regionFields
		}
	}

	return regionSpecific
}

func (s *WebsiteRegionalContentService) mergeContent(primary, common map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Add common fields first
	for key, value := range common {
		merged[key] = value
	}

	// Add primary region fields (may override common fields)
	for key, value := range primary {
		merged[key] = value
	}

	return merged
}

func (s *WebsiteRegionalContentService) calculateNormalizationScore(normalized *NormalizedRegionalData) float64 {
	score := 0.0
	totalChecks := 0.0

	// Score based on number of common fields
	if len(normalized.CommonFields) > 0 {
		score += 0.3
		totalChecks += 0.3
	}

	// Score based on number of supported regions
	if len(normalized.SupportedRegions) > 1 {
		regionScore := float64(len(normalized.SupportedRegions)) / 5.0 // Normalize to max 5 regions
		if regionScore > 1.0 {
			regionScore = 1.0
		}
		score += regionScore * 0.3
		totalChecks += 0.3
	}

	// Score based on data completeness
	if len(normalized.NormalizedContent) > 0 {
		score += 0.2
		totalChecks += 0.2
	}

	// Score based on currency mappings
	if len(normalized.CurrencyMappings) > 0 {
		score += 0.1
		totalChecks += 0.1
	}

	// Score based on contact mappings
	if len(normalized.ContactMappings) > 0 {
		score += 0.1
		totalChecks += 0.1
	}

	if totalChecks == 0 {
		return 0.5
	}

	return score / totalChecks
}

// Helper utility methods

func (s *WebsiteRegionalContentService) extractTextFromData(data map[string]interface{}) string {
	var textParts []string
	for _, value := range data {
		if str, ok := value.(string); ok {
			textParts = append(textParts, str)
		}
	}
	return strings.Join(textParts, " ")
}

func (s *WebsiteRegionalContentService) calculateCurrencyConfidence(currency, region string, matchCount int) float64 {
	baseConfidence := 0.5

	// Higher confidence for more matches
	frequencyBonus := float64(matchCount) * 0.1
	if frequencyBonus > 0.3 {
		frequencyBonus = 0.3
	}

	// Higher confidence if currency matches region
	regionBonus := 0.0
	if s.currencyMatchesRegion(currency, region) {
		regionBonus = 0.2
	}

	return baseConfidence + frequencyBonus + regionBonus
}

func (s *WebsiteRegionalContentService) calculatePhoneConfidence(phoneRegion, actualRegion string, matchCount int) float64 {
	baseConfidence := 0.5

	// Higher confidence for more matches
	frequencyBonus := float64(matchCount) * 0.1
	if frequencyBonus > 0.3 {
		frequencyBonus = 0.3
	}

	// Higher confidence if phone region matches actual region
	regionBonus := 0.0
	if phoneRegion == actualRegion {
		regionBonus = 0.2
	}

	return baseConfidence + frequencyBonus + regionBonus
}

func (s *WebsiteRegionalContentService) calculateAddressConfidence(addressRegion, actualRegion string, matchCount int) float64 {
	baseConfidence := 0.5

	// Higher confidence for more matches
	frequencyBonus := float64(matchCount) * 0.1
	if frequencyBonus > 0.3 {
		frequencyBonus = 0.3
	}

	// Higher confidence if address region matches actual region
	regionBonus := 0.0
	if addressRegion == actualRegion {
		regionBonus = 0.2
	}

	return baseConfidence + frequencyBonus + regionBonus
}

func (s *WebsiteRegionalContentService) currencyMatchesRegion(currency, region string) bool {
	regionCurrencyMap := map[string]string{
		"US": "USD",
		"CA": "CAD",
		"GB": "GBP",
		"EU": "EUR",
		"AU": "AUD",
	}

	expectedCurrency, exists := regionCurrencyMap[region]
	return exists && expectedCurrency == currency
}

func (s *WebsiteRegionalContentService) getCurrencyRegion(currency string) string {
	currencyRegionMap := map[string]string{
		"USD": "US",
		"CAD": "CA",
		"GBP": "GB",
		"EUR": "EU",
		"AUD": "AU",
	}

	if region, exists := currencyRegionMap[currency]; exists {
		return region
	}
	return ""
}

func (s *WebsiteRegionalContentService) getLanguageRegion(language string) string {
	languageRegionMap := map[string]string{
		"en": "US",
		"es": "ES",
		"fr": "FR",
		"de": "DE",
		"it": "IT",
		"pt": "BR",
	}

	if region, exists := languageRegionMap[language]; exists {
		return region
	}
	return ""
}

func (s *WebsiteRegionalContentService) getCurrencySymbol(currency string) string {
	currencySymbolMap := map[string]string{
		"USD": "$",
		"CAD": "C$",
		"GBP": "£",
		"EUR": "€",
		"AUD": "A$",
		"JPY": "¥",
	}

	if symbol, exists := currencySymbolMap[currency]; exists {
		return symbol
	}
	return currency
}

func (s *WebsiteRegionalContentService) getCurrencyDecimalPlaces(currency string) int {
	// Most currencies use 2 decimal places
	switch currency {
	case "JPY", "KRW":
		return 0 // These currencies typically don't use decimal places
	default:
		return 2
	}
}

func (s *WebsiteRegionalContentService) getCountryCodeForRegion(region string) string {
	regionCountryCodeMap := map[string]string{
		"US": "+1",
		"CA": "+1",
		"GB": "+44",
		"DE": "+49",
		"FR": "+33",
		"IT": "+39",
		"ES": "+34",
		"AU": "+61",
	}

	if code, exists := regionCountryCodeMap[region]; exists {
		return code
	}
	return ""
}

func (s *WebsiteRegionalContentService) isTollFree(phoneNumber string) bool {
	// Simple toll-free detection for US/CA numbers
	tollFreePatterns := []string{"800", "888", "877", "866", "855", "844", "833", "822"}

	for _, pattern := range tollFreePatterns {
		if strings.Contains(phoneNumber, pattern) {
			return true
		}
	}

	return false
}

func (s *WebsiteRegionalContentService) getTimezoneForRegion(region string) string {
	regionTimezoneMap := map[string]string{
		"US": "America/New_York",
		"CA": "America/Toronto",
		"GB": "Europe/London",
		"DE": "Europe/Berlin",
		"FR": "Europe/Paris",
		"IT": "Europe/Rome",
		"ES": "Europe/Madrid",
		"AU": "Australia/Sydney",
	}

	if timezone, exists := regionTimezoneMap[region]; exists {
		return timezone
	}
	return "UTC"
}

func (s *WebsiteRegionalContentService) getRegionalHolidays(region string) []string {
	// Return common holidays for the region
	regionalHolidaysMap := map[string][]string{
		"US": {"New Year's Day", "Independence Day", "Thanksgiving", "Christmas"},
		"CA": {"New Year's Day", "Canada Day", "Thanksgiving", "Christmas"},
		"GB": {"New Year's Day", "Easter", "Christmas", "Boxing Day"},
		"DE": {"New Year's Day", "Easter", "Christmas", "German Unity Day"},
		"FR": {"New Year's Day", "Easter", "Christmas", "Bastille Day"},
	}

	if holidays, exists := regionalHolidaysMap[region]; exists {
		return holidays
	}
	return []string{}
}
