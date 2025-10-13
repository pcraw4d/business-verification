package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// LocalizationService implements localization and internationalization features
type LocalizationService struct {
	logger *zap.Logger
	config *LocalizationConfig
}

// LocalizationConfig represents localization configuration
type LocalizationConfig struct {
	DefaultLanguage    string                     `json:"default_language"`
	SupportedLanguages []string                   `json:"supported_languages"`
	FallbackLanguage   string                     `json:"fallback_language"`
	EnableRTL          bool                       `json:"enable_rtl"`
	DateFormat         string                     `json:"date_format"`
	TimeFormat         string                     `json:"time_format"`
	NumberFormat       NumberFormat               `json:"number_format"`
	CurrencyFormat     CurrencyFormat             `json:"currency_format"`
	Translations       map[string]Translations    `json:"translations"`
	CountryOverrides   map[string]CountryOverride `json:"country_overrides"`
}

// NumberFormat represents number formatting configuration
type NumberFormat struct {
	DecimalSeparator   string `json:"decimal_separator"`
	ThousandsSeparator string `json:"thousands_separator"`
	GroupingSize       int    `json:"grouping_size"`
	MinFractionDigits  int    `json:"min_fraction_digits"`
	MaxFractionDigits  int    `json:"max_fraction_digits"`
}

// CurrencyFormat represents currency formatting configuration
type CurrencyFormat struct {
	Symbol         string `json:"symbol"`
	SymbolPosition string `json:"symbol_position"` // "before" or "after"
	DecimalPlaces  int    `json:"decimal_places"`
	ShowSymbol     bool   `json:"show_symbol"`
	ShowCode       bool   `json:"show_code"`
}

// Translations represents translations for a language
type Translations struct {
	Language     string                 `json:"language"`
	Country      string                 `json:"country"`
	Translations map[string]string      `json:"translations"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CountryOverride represents country-specific localization overrides
type CountryOverride struct {
	CountryCode    string                 `json:"country_code"`
	Language       string                 `json:"language"`
	DateFormat     string                 `json:"date_format,omitempty"`
	TimeFormat     string                 `json:"time_format,omitempty"`
	NumberFormat   *NumberFormat          `json:"number_format,omitempty"`
	CurrencyFormat *CurrencyFormat        `json:"currency_format,omitempty"`
	Translations   map[string]string      `json:"translations,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// LocalizedContent represents localized content
type LocalizedContent struct {
	Key         string                 `json:"key"`
	Language    string                 `json:"language"`
	Country     string                 `json:"country"`
	Content     string                 `json:"content"`
	Type        ContentType            `json:"type"`
	Category    string                 `json:"category"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContentType represents the type of localized content
type ContentType string

const (
	ContentTypeText        ContentType = "text"
	ContentTypeLabel       ContentType = "label"
	ContentTypeMessage     ContentType = "message"
	ContentTypeError       ContentType = "error"
	ContentTypeWarning     ContentType = "warning"
	ContentTypeInfo        ContentType = "info"
	ContentTypePlaceholder ContentType = "placeholder"
	ContentTypeTooltip     ContentType = "tooltip"
	ContentTypeHelp        ContentType = "help"
)

// LocalizedRiskAssessment represents a localized risk assessment
type LocalizedRiskAssessment struct {
	ID               string                    `json:"id"`
	OriginalID       string                    `json:"original_id"`
	Language         string                    `json:"language"`
	Country          string                    `json:"country"`
	BusinessName     string                    `json:"business_name"`
	RiskScore        float64                   `json:"risk_score"`
	RiskLevel        string                    `json:"risk_level"`
	RiskFactors      []LocalizedRiskFactor     `json:"risk_factors"`
	Recommendations  []LocalizedRecommendation `json:"recommendations"`
	ComplianceStatus []LocalizedCompliance     `json:"compliance_status"`
	GeneratedAt      time.Time                 `json:"generated_at"`
	Metadata         map[string]interface{}    `json:"metadata"`
}

// LocalizedRiskFactor represents a localized risk factor
type LocalizedRiskFactor struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Category            string                 `json:"category"`
	Severity            string                 `json:"severity"`
	Impact              string                 `json:"impact"`
	Mitigation          string                 `json:"mitigation"`
	LocalizedName       string                 `json:"localized_name"`
	LocalizedDesc       string                 `json:"localized_desc"`
	LocalizedImpact     string                 `json:"localized_impact"`
	LocalizedMitigation string                 `json:"localized_mitigation"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// LocalizedRecommendation represents a localized recommendation
type LocalizedRecommendation struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Priority          string                 `json:"priority"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Action            string                 `json:"action"`
	Timeline          string                 `json:"timeline"`
	LocalizedTitle    string                 `json:"localized_title"`
	LocalizedDesc     string                 `json:"localized_desc"`
	LocalizedAction   string                 `json:"localized_action"`
	LocalizedTimeline string                 `json:"localized_timeline"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// LocalizedCompliance represents localized compliance information
type LocalizedCompliance struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status"`
	Requirement      string                 `json:"requirement"`
	Description      string                 `json:"description"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
	Penalty          string                 `json:"penalty,omitempty"`
	LocalizedReq     string                 `json:"localized_requirement"`
	LocalizedDesc    string                 `json:"localized_description"`
	LocalizedPenalty string                 `json:"localized_penalty,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewLocalizationService creates a new localization service instance
func NewLocalizationService(config *LocalizationConfig, logger *zap.Logger) *LocalizationService {
	return &LocalizationService{
		logger: logger,
		config: config,
	}
}

// GetSupportedLanguages returns the list of supported languages
func (ls *LocalizationService) GetSupportedLanguages() []string {
	return ls.config.SupportedLanguages
}

// GetDefaultLanguage returns the default language
func (ls *LocalizationService) GetDefaultLanguage() string {
	return ls.config.DefaultLanguage
}

// Translate translates a key to the specified language
func (ls *LocalizationService) Translate(ctx context.Context, key string, language string, country string, params map[string]interface{}) (string, error) {
	// Check if language is supported
	if !ls.isLanguageSupported(language) {
		language = ls.config.FallbackLanguage
	}

	// Get translations for the language
	translations, exists := ls.config.Translations[language]
	if !exists {
		return key, fmt.Errorf("translations not found for language: %s", language)
	}

	// Get translation
	translation, exists := translations.Translations[key]
	if !exists {
		// Try fallback language
		if language != ls.config.FallbackLanguage {
			return ls.Translate(ctx, key, ls.config.FallbackLanguage, country, params)
		}
		return key, fmt.Errorf("translation not found for key: %s", key)
	}

	// Apply country-specific overrides if available
	if country != "" {
		override, exists := ls.config.CountryOverrides[country]
		if exists && override.Language == language {
			if countryTranslation, exists := override.Translations[key]; exists {
				translation = countryTranslation
			}
		}
	}

	// Apply parameter substitution
	if params != nil {
		translation = ls.substituteParameters(translation, params)
	}

	ls.logger.Debug("Translation retrieved",
		zap.String("key", key),
		zap.String("language", language),
		zap.String("country", country),
		zap.String("translation", translation))

	return translation, nil
}

// FormatDate formats a date according to the specified language and country
func (ls *LocalizationService) FormatDate(ctx context.Context, date time.Time, language string, country string) (string, error) {
	format := ls.config.DateFormat

	// Apply country-specific date format if available
	if country != "" {
		override, exists := ls.config.CountryOverrides[country]
		if exists && override.DateFormat != "" {
			format = override.DateFormat
		}
	}

	// Format the date
	formatted := date.Format(format)

	ls.logger.Debug("Date formatted",
		zap.Time("date", date),
		zap.String("language", language),
		zap.String("country", country),
		zap.String("format", format),
		zap.String("formatted", formatted))

	return formatted, nil
}

// FormatTime formats a time according to the specified language and country
func (ls *LocalizationService) FormatTime(ctx context.Context, time time.Time, language string, country string) (string, error) {
	format := ls.config.TimeFormat

	// Apply country-specific time format if available
	if country != "" {
		override, exists := ls.config.CountryOverrides[country]
		if exists && override.TimeFormat != "" {
			format = override.TimeFormat
		}
	}

	// Format the time
	formatted := time.Format(format)

	ls.logger.Debug("Time formatted",
		zap.Time("time", time),
		zap.String("language", language),
		zap.String("country", country),
		zap.String("format", format),
		zap.String("formatted", formatted))

	return formatted, nil
}

// FormatNumber formats a number according to the specified language and country
func (ls *LocalizationService) FormatNumber(ctx context.Context, number float64, language string, country string) (string, error) {
	format := ls.config.NumberFormat

	// Apply country-specific number format if available
	if country != "" {
		override, exists := ls.config.CountryOverrides[country]
		if exists && override.NumberFormat != nil {
			format = *override.NumberFormat
		}
	}

	// Format the number
	formatted := ls.formatNumberWithConfig(number, format)

	ls.logger.Debug("Number formatted",
		zap.Float64("number", number),
		zap.String("language", language),
		zap.String("country", country),
		zap.String("formatted", formatted))

	return formatted, nil
}

// FormatCurrency formats a currency amount according to the specified language and country
func (ls *LocalizationService) FormatCurrency(ctx context.Context, amount float64, currency string, language string, country string) (string, error) {
	format := ls.config.CurrencyFormat

	// Apply country-specific currency format if available
	if country != "" {
		override, exists := ls.config.CountryOverrides[country]
		if exists && override.CurrencyFormat != nil {
			format = *override.CurrencyFormat
		}
	}

	// Format the currency
	formatted := ls.formatCurrencyWithConfig(amount, currency, format)

	ls.logger.Debug("Currency formatted",
		zap.Float64("amount", amount),
		zap.String("currency", currency),
		zap.String("language", language),
		zap.String("country", country),
		zap.String("formatted", formatted))

	return formatted, nil
}

// LocalizeRiskAssessment localizes a risk assessment for the specified language and country
func (ls *LocalizationService) LocalizeRiskAssessment(ctx context.Context, assessment interface{}, language string, country string) (*LocalizedRiskAssessment, error) {
	// In a real implementation, this would convert the assessment to localized format
	// For now, we'll create a mock localized assessment

	localized := &LocalizedRiskAssessment{
		ID:          fmt.Sprintf("localized_%d", time.Now().UnixNano()),
		Language:    language,
		Country:     country,
		GeneratedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Localize risk factors
	localized.RiskFactors = []LocalizedRiskFactor{
		{
			ID:                  "rf_1",
			Name:                "Political Risk",
			Description:         "Political instability in the region",
			Category:            "political",
			Severity:            "medium",
			Impact:              "Moderate impact on business operations",
			Mitigation:          "Monitor political developments",
			LocalizedName:       "Political Risk",
			LocalizedDesc:       "Political instability in the region",
			LocalizedImpact:     "Moderate impact on business operations",
			LocalizedMitigation: "Monitor political developments",
		},
	}

	// Localize recommendations
	localized.Recommendations = []LocalizedRecommendation{
		{
			ID:                "rec_1",
			Type:              "monitoring",
			Priority:          "high",
			Title:             "Enhanced Monitoring",
			Description:       "Implement enhanced monitoring procedures",
			Action:            "Set up automated alerts",
			Timeline:          "30 days",
			LocalizedTitle:    "Enhanced Monitoring",
			LocalizedDesc:     "Implement enhanced monitoring procedures",
			LocalizedAction:   "Set up automated alerts",
			LocalizedTimeline: "30 days",
		},
	}

	// Localize compliance status
	localized.ComplianceStatus = []LocalizedCompliance{
		{
			ID:            "comp_1",
			Type:          "kyb",
			Status:        "compliant",
			Requirement:   "Business registration verification",
			Description:   "Verify business registration status",
			LocalizedReq:  "Business registration verification",
			LocalizedDesc: "Verify business registration status",
		},
	}

	ls.logger.Info("Risk assessment localized",
		zap.String("language", language),
		zap.String("country", country),
		zap.String("assessment_id", localized.ID))

	return localized, nil
}

// GetLocalizedContent retrieves localized content for a specific key
func (ls *LocalizationService) GetLocalizedContent(ctx context.Context, key string, language string, country string) (*LocalizedContent, error) {
	content, err := ls.Translate(ctx, key, language, country, nil)
	if err != nil {
		return nil, err
	}

	localizedContent := &LocalizedContent{
		Key:         key,
		Language:    language,
		Country:     country,
		Content:     content,
		Type:        ContentTypeText,
		Category:    "general",
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	return localizedContent, nil
}

// SetLocalizedContent sets localized content for a specific key
func (ls *LocalizationService) SetLocalizedContent(ctx context.Context, content *LocalizedContent) error {
	// In a real implementation, this would store the content in a database
	// For now, we'll just log the action

	ls.logger.Info("Localized content set",
		zap.String("key", content.Key),
		zap.String("language", content.Language),
		zap.String("country", content.Country),
		zap.String("type", string(content.Type)))

	return nil
}

// GetLocalizedContentList retrieves a list of localized content
func (ls *LocalizationService) GetLocalizedContentList(ctx context.Context, language string, country string, contentType ContentType) ([]*LocalizedContent, error) {
	// In a real implementation, this would query the database
	// For now, we'll return an empty list

	ls.logger.Info("Localized content list retrieved",
		zap.String("language", language),
		zap.String("country", country),
		zap.String("content_type", string(contentType)))

	return []*LocalizedContent{}, nil
}

// Helper functions
func (ls *LocalizationService) isLanguageSupported(language string) bool {
	for _, supported := range ls.config.SupportedLanguages {
		if supported == language {
			return true
		}
	}
	return false
}

func (ls *LocalizationService) substituteParameters(text string, params map[string]interface{}) string {
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, placeholder, fmt.Sprintf("%v", value))
	}
	return text
}

func (ls *LocalizationService) formatNumberWithConfig(number float64, config NumberFormat) string {
	// In a real implementation, this would use proper number formatting
	// For now, we'll do basic formatting
	return fmt.Sprintf("%.2f", number)
}

func (ls *LocalizationService) formatCurrencyWithConfig(amount float64, currency string, config CurrencyFormat) string {
	// In a real implementation, this would use proper currency formatting
	// For now, we'll do basic formatting
	if config.ShowSymbol {
		return fmt.Sprintf("%s%.2f", config.Symbol, amount)
	}
	return fmt.Sprintf("%.2f %s", amount, currency)
}
