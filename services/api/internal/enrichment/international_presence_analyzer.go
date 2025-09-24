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

// InternationalPresenceAnalyzer analyzes international presence and localization features
type InternationalPresenceAnalyzer struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *InternationalPresenceAnalyzerConfig
}

// InternationalPresenceAnalyzerConfig contains configuration for international presence analysis
type InternationalPresenceAnalyzerConfig struct {
	// Country codes and names mapping
	CountryMapping map[string]string
	// Language codes and names mapping
	LanguageMapping map[string]string
	// International business indicators
	InternationalIndicators []string
	// Localization features to detect
	LocalizationFeatures []string
	// Confidence thresholds
	MinConfidenceScore float64
	// Maximum countries to detect
	MaxCountries int
}

// InternationalPresence represents international business presence
type InternationalPresence struct {
	Type            string    `json:"type"`             // global, regional, local, multinational
	Countries       []string  `json:"countries"`        // Countries where business operates
	Regions         []string  `json:"regions"`          // Geographic regions (NA, EU, APAC, etc.)
	Languages       []string  `json:"languages"`        // Supported languages
	Localization    []string  `json:"localization"`     // Localization features
	ConfidenceScore float64   `json:"confidence_score"` // Confidence in analysis
	ExtractedAt     time.Time `json:"extracted_at"`     // When this was extracted
	Evidence        []string  `json:"evidence"`         // Evidence supporting analysis
}

// LocalizationFeature represents a specific localization feature
type LocalizationFeature struct {
	Type            string  `json:"type"`             // language, currency, timezone, culture
	Name            string  `json:"name"`             // Feature name
	Value           string  `json:"value"`            // Feature value
	ConfidenceScore float64 `json:"confidence_score"` // Confidence in detection
}

// InternationalPresenceResult contains the results of international presence analysis
type InternationalPresenceResult struct {
	Presence             *InternationalPresence `json:"presence"`              // Overall international presence
	LocalizationFeatures []LocalizationFeature  `json:"localization_features"` // Detected localization features
	Countries            []string               `json:"countries"`             // All detected countries
	Languages            []string               `json:"languages"`             // All supported languages
	Regions              []string               `json:"regions"`               // All detected regions
	IsInternational      bool                   `json:"is_international"`      // Whether business is international
	GlobalReach          string                 `json:"global_reach"`          // Global reach classification
	ConfidenceScore      float64                `json:"confidence_score"`      // Overall confidence score
	Evidence             []string               `json:"evidence"`              // Evidence supporting analysis
	ProcessingTime       time.Duration          `json:"processing_time"`       // Time taken to process
}

// NewInternationalPresenceAnalyzer creates a new international presence analyzer
func NewInternationalPresenceAnalyzer(logger *zap.Logger, config *InternationalPresenceAnalyzerConfig) *InternationalPresenceAnalyzer {
	if config == nil {
		config = &InternationalPresenceAnalyzerConfig{
			CountryMapping: map[string]string{
				"usa": "United States", "us": "United States", "united states": "United States",
				"uk": "United Kingdom", "united kingdom": "United Kingdom", "england": "United Kingdom",
				"canada": "Canada", "ca": "Canada",
				"australia": "Australia", "au": "Australia",
				"germany": "Germany", "de": "Germany",
				"france": "France", "fr": "France",
				"spain": "Spain", "es": "Spain",
				"italy": "Italy", "it": "Italy",
				"japan": "Japan", "jp": "Japan",
				"china": "China", "cn": "China",
				"india": "India", "in": "India",
				"brazil": "Brazil", "br": "Brazil",
				"mexico": "Mexico", "mx": "Mexico",
				"singapore": "Singapore", "sg": "Singapore",
				"hong kong": "Hong Kong", "hk": "Hong Kong",
				"dubai": "UAE", "uae": "UAE", "united arab emirates": "UAE",
			},
			LanguageMapping: map[string]string{
				"en": "English", "english": "English",
				"es": "Spanish", "spanish": "Spanish", "español": "Spanish",
				"fr": "French", "french": "French", "français": "French",
				"de": "German", "german": "German", "deutsch": "German",
				"it": "Italian", "italian": "Italian", "italiano": "Italian",
				"pt": "Portuguese", "portuguese": "Portuguese", "português": "Portuguese",
				"ru": "Russian", "russian": "Russian", "русский": "Russian",
				"zh": "Chinese", "chinese": "Chinese", "中文": "Chinese",
				"ja": "Japanese", "japanese": "Japanese", "日本語": "Japanese",
				"ko": "Korean", "korean": "Korean", "한국어": "Korean",
				"ar": "Arabic", "arabic": "Arabic", "العربية": "Arabic",
				"hi": "Hindi", "hindi": "Hindi", "हिन्दी": "Hindi",
			},
			InternationalIndicators: []string{
				"global", "international", "worldwide", "world-wide", "multinational",
				"europe", "asia", "africa", "americas", "north america", "south america",
				"european", "asian", "african", "american", "international presence",
				"global presence", "worldwide presence", "operates in", "serving",
				"locations in", "offices in", "branches in", "presence in",
				"available in", "serving customers in", "global reach", "international reach",
			},
			LocalizationFeatures: []string{
				"language", "currency", "timezone", "date format", "number format",
				"localization", "l10n", "internationalization", "i18n", "translation",
				"local content", "regional content", "country-specific", "localized",
				"multi-language", "multi-currency", "multi-timezone", "cultural adaptation",
			},
			MinConfidenceScore: 0.3,
			MaxCountries:       20,
		}
	}

	return &InternationalPresenceAnalyzer{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("international_presence_analyzer"),
		config: config,
	}
}

// AnalyzeInternationalPresence analyzes international presence and localization from website content
func (ipa *InternationalPresenceAnalyzer) AnalyzeInternationalPresence(ctx context.Context, content string) (*InternationalPresenceResult, error) {
	ctx, span := ipa.tracer.Start(ctx, "international_presence_analyzer.analyze")
	defer span.End()

	startTime := time.Now()
	lowerContent := strings.ToLower(content)

	// Extract countries
	countries := ipa.extractCountries(lowerContent)

	// Extract languages
	languages := ipa.extractLanguages(lowerContent)

	// Extract regions
	regions := ipa.extractRegions(lowerContent)

	// Extract localization features
	localizationFeatures := ipa.extractLocalizationFeatures(lowerContent)

	// Determine international presence type
	presenceType := ipa.determinePresenceType(countries, regions, lowerContent)

	// Calculate confidence scores
	confidenceScore := ipa.calculateConfidenceScore(countries, languages, regions, localizationFeatures, lowerContent)

	// Determine global reach
	globalReach := ipa.determineGlobalReach(countries, regions)

	// Collect evidence
	evidence := ipa.collectEvidence(countries, languages, regions, localizationFeatures, lowerContent)

	// Create international presence
	presence := &InternationalPresence{
		Type:            presenceType,
		Countries:       countries,
		Regions:         regions,
		Languages:       languages,
		Localization:    ipa.getLocalizationTypes(localizationFeatures),
		ConfidenceScore: confidenceScore,
		ExtractedAt:     time.Now(),
		Evidence:        evidence,
	}

	result := &InternationalPresenceResult{
		Presence:             presence,
		LocalizationFeatures: localizationFeatures,
		Countries:            countries,
		Languages:            languages,
		Regions:              regions,
		IsInternational:      len(countries) > 1 || len(regions) > 1,
		GlobalReach:          globalReach,
		ConfidenceScore:      confidenceScore,
		Evidence:             evidence,
		ProcessingTime:       time.Since(startTime),
	}

	ipa.logger.Info("international presence analysis completed",
		zap.String("presence_type", presenceType),
		zap.Strings("countries", countries),
		zap.Strings("languages", languages),
		zap.Float64("confidence_score", confidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// extractCountries extracts country information from content
func (ipa *InternationalPresenceAnalyzer) extractCountries(content string) []string {
	var countries []string
	seen := make(map[string]bool)

	// Extract countries using country mapping
	for code, name := range ipa.config.CountryMapping {
		if strings.Contains(content, code) || strings.Contains(content, name) {
			if !seen[name] {
				countries = append(countries, name)
				seen[name] = true
			}
		}
	}

	// Extract countries using regex patterns
	countryPatterns := []string{
		`(?:serving|operating|presence|offices|branches)\s+(?:in|at)\s+([A-Za-z\s]+)`,
		`(?:available|serving)\s+(?:in|at)\s+([A-Za-z\s]+)`,
		`(?:headquarters|office|branch)\s+(?:in|at)\s+([A-Za-z\s]+)`,
	}

	for _, pattern := range countryPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				country := strings.TrimSpace(match[1])
				if country != "" && !seen[country] {
					// Check if it's a known country
					for _, knownCountry := range ipa.config.CountryMapping {
						if strings.Contains(strings.ToLower(knownCountry), strings.ToLower(country)) ||
							strings.Contains(strings.ToLower(country), strings.ToLower(knownCountry)) {
							countries = append(countries, knownCountry)
							seen[knownCountry] = true
							break
						}
					}
				}
			}
		}
	}

	// Limit to max countries
	if len(countries) > ipa.config.MaxCountries {
		countries = countries[:ipa.config.MaxCountries]
	}

	return countries
}

// extractLanguages extracts language information from content
func (ipa *InternationalPresenceAnalyzer) extractLanguages(content string) []string {
	var languages []string
	seen := make(map[string]bool)

	// Extract languages using language mapping
	for code, name := range ipa.config.LanguageMapping {
		if strings.Contains(content, code) || strings.Contains(content, name) {
			if !seen[name] {
				languages = append(languages, name)
				seen[name] = true
			}
		}
	}

	// Extract languages using regex patterns
	languagePatterns := []string{
		`(?:language|lang|translation)\s*[:\-]?\s*([A-Za-z\s]+)`,
		`(?:available\s+in|supported\s+in)\s+([A-Za-z\s]+)`,
		`(?:multi-language|multilingual|bilingual)\s+([A-Za-z\s]+)`,
	}

	for _, pattern := range languagePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				language := strings.TrimSpace(match[1])
				if language != "" && !seen[language] {
					// Check if it's a known language
					for _, knownLanguage := range ipa.config.LanguageMapping {
						if strings.Contains(strings.ToLower(knownLanguage), strings.ToLower(language)) ||
							strings.Contains(strings.ToLower(language), strings.ToLower(knownLanguage)) {
							languages = append(languages, knownLanguage)
							seen[knownLanguage] = true
							break
						}
					}
				}
			}
		}
	}

	return languages
}

// extractRegions extracts geographic regions from content
func (ipa *InternationalPresenceAnalyzer) extractRegions(content string) []string {
	var regions []string
	seen := make(map[string]bool)

	regionPatterns := map[string][]string{
		"North America": {"north america", "north american", "na", "us & canada", "united states and canada"},
		"Europe":        {"europe", "european", "eu", "european union"},
		"Asia Pacific":  {"asia pacific", "apac", "asia-pacific", "asian pacific"},
		"Asia":          {"asia", "asian"},
		"Latin America": {"latin america", "latin american", "latam", "south america", "south american"},
		"Middle East":   {"middle east", "middle eastern", "mena", "gcc"},
		"Africa":        {"africa", "african"},
		"Oceania":       {"oceania", "australasia", "australia and new zealand"},
	}

	for region, patterns := range regionPatterns {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) && !seen[region] {
				regions = append(regions, region)
				seen[region] = true
				break
			}
		}
	}

	return regions
}

// extractLocalizationFeatures extracts localization features from content
func (ipa *InternationalPresenceAnalyzer) extractLocalizationFeatures(content string) []LocalizationFeature {
	var features []LocalizationFeature

	// Language features
	languageFeatures := ipa.extractLanguageFeatures(content)
	features = append(features, languageFeatures...)

	// Currency features
	currencyFeatures := ipa.extractCurrencyFeatures(content)
	features = append(features, currencyFeatures...)

	// Timezone features
	timezoneFeatures := ipa.extractTimezoneFeatures(content)
	features = append(features, timezoneFeatures...)

	// Cultural features
	culturalFeatures := ipa.extractCulturalFeatures(content)
	features = append(features, culturalFeatures...)

	return features
}

// extractLanguageFeatures extracts language-related localization features
func (ipa *InternationalPresenceAnalyzer) extractLanguageFeatures(content string) []LocalizationFeature {
	var features []LocalizationFeature

	languagePatterns := []string{
		`(?:language|lang)\s+selector`,
		`(?:language|lang)\s+switcher`,
		`(?:translation|translate)\s+(?:service|tool)`,
		`(?:multi-language|multilingual|bilingual)`,
		`(?:localized|localization)\s+content`,
	}

	for _, pattern := range languagePatterns {
		if regexp.MustCompile(pattern).MatchString(content) {
			features = append(features, LocalizationFeature{
				Type:            "language",
				Name:            "Multi-language Support",
				Value:           "Detected",
				ConfidenceScore: 0.8,
			})
			break
		}
	}

	return features
}

// extractCurrencyFeatures extracts currency-related localization features
func (ipa *InternationalPresenceAnalyzer) extractCurrencyFeatures(content string) []LocalizationFeature {
	var features []LocalizationFeature

	currencyPatterns := []string{
		`(?:currency|currencies)\s+supported`,
		`(?:multi-currency|multiple\s+currencies)`,
		`(?:usd|eur|gbp|jpy|cad|aud|chf|sek|nok|dkk|pln|czk|huf|ron|bgn|hrk|rub|try|brl|mxn|zar|inr|cny|krw|sgd|hkd|nzd)`,
		`(?:dollar|euro|pound|yen|franc|krona|koruna|forint|leu|lev|kuna|ruble|lira|real|peso|rand|rupee|yuan|won|dinar)`,
	}

	for _, pattern := range currencyPatterns {
		if regexp.MustCompile(pattern).MatchString(content) {
			features = append(features, LocalizationFeature{
				Type:            "currency",
				Name:            "Multi-currency Support",
				Value:           "Detected",
				ConfidenceScore: 0.7,
			})
			break
		}
	}

	return features
}

// extractTimezoneFeatures extracts timezone-related localization features
func (ipa *InternationalPresenceAnalyzer) extractTimezoneFeatures(content string) []LocalizationFeature {
	var features []LocalizationFeature

	timezonePatterns := []string{
		`(?:timezone|time\s+zone)\s+(?:support|handling)`,
		`(?:local\s+time|time\s+conversion)`,
		`(?:utc|gmt|est|cst|mst|pst|cet|eet|jst|aest|nzst)`,
	}

	for _, pattern := range timezonePatterns {
		if regexp.MustCompile(pattern).MatchString(content) {
			features = append(features, LocalizationFeature{
				Type:            "timezone",
				Name:            "Timezone Support",
				Value:           "Detected",
				ConfidenceScore: 0.6,
			})
			break
		}
	}

	return features
}

// extractCulturalFeatures extracts cultural adaptation features
func (ipa *InternationalPresenceAnalyzer) extractCulturalFeatures(content string) []LocalizationFeature {
	var features []LocalizationFeature

	culturalPatterns := []string{
		`(?:cultural|culture)\s+(?:adaptation|localization)`,
		`(?:local\s+customs|cultural\s+sensitivity)`,
		`(?:date\s+format|number\s+format)`,
		`(?:local\s+content|regional\s+content)`,
	}

	for _, pattern := range culturalPatterns {
		if regexp.MustCompile(pattern).MatchString(content) {
			features = append(features, LocalizationFeature{
				Type:            "culture",
				Name:            "Cultural Adaptation",
				Value:           "Detected",
				ConfidenceScore: 0.5,
			})
			break
		}
	}

	return features
}

// determinePresenceType determines the type of international presence
func (ipa *InternationalPresenceAnalyzer) determinePresenceType(countries []string, regions []string, content string) string {
	// Check for global indicators
	globalIndicators := []string{"global", "worldwide", "world-wide", "international"}
	for _, indicator := range globalIndicators {
		if strings.Contains(content, indicator) {
			return "global"
		}
	}

	// Check number of countries and regions
	if len(countries) > 5 || len(regions) > 3 {
		return "global"
	} else if len(countries) > 2 || len(regions) > 1 {
		return "multinational"
	} else if len(countries) > 1 || len(regions) > 0 {
		return "regional"
	} else {
		return "local"
	}
}

// calculateConfidenceScore calculates the overall confidence score
func (ipa *InternationalPresenceAnalyzer) calculateConfidenceScore(countries []string, languages []string, regions []string, features []LocalizationFeature, content string) float64 {
	score := 0.0
	totalFactors := 0.0

	// Country detection confidence
	if len(countries) > 0 {
		countryScore := minFloat64(1.0, float64(len(countries))*0.2)
		score += countryScore
		totalFactors += 1.0
	}

	// Language detection confidence
	if len(languages) > 0 {
		languageScore := minFloat64(1.0, float64(len(languages))*0.15)
		score += languageScore
		totalFactors += 1.0
	}

	// Region detection confidence
	if len(regions) > 0 {
		regionScore := minFloat64(1.0, float64(len(regions))*0.25)
		score += regionScore
		totalFactors += 1.0
	}

	// Localization features confidence
	if len(features) > 0 {
		featureScore := minFloat64(1.0, float64(len(features))*0.1)
		score += featureScore
		totalFactors += 1.0
	}

	// International indicators confidence
	indicatorCount := 0
	for _, indicator := range ipa.config.InternationalIndicators {
		if strings.Contains(content, indicator) {
			indicatorCount++
		}
	}
	if indicatorCount > 0 {
		indicatorScore := minFloat64(1.0, float64(indicatorCount)*0.1)
		score += indicatorScore
		totalFactors += 1.0
	}

	if totalFactors == 0 {
		return 0.0
	}

	return score / totalFactors
}

// determineGlobalReach determines the global reach classification
func (ipa *InternationalPresenceAnalyzer) determineGlobalReach(countries []string, regions []string) string {
	if len(countries) > 10 || len(regions) > 4 {
		return "global"
	} else if len(countries) > 5 || len(regions) > 2 {
		return "international"
	} else if len(countries) > 2 || len(regions) > 1 {
		return "regional"
	} else if len(countries) > 1 {
		return "multi-country"
	} else {
		return "local"
	}
}

// collectEvidence collects evidence supporting the analysis
func (ipa *InternationalPresenceAnalyzer) collectEvidence(countries []string, languages []string, regions []string, features []LocalizationFeature, content string) []string {
	var evidence []string

	// Add country evidence
	for _, country := range countries {
		evidence = append(evidence, fmt.Sprintf("Detected presence in: %s", country))
	}

	// Add language evidence
	for _, language := range languages {
		evidence = append(evidence, fmt.Sprintf("Supports language: %s", language))
	}

	// Add region evidence
	for _, region := range regions {
		evidence = append(evidence, fmt.Sprintf("Operates in region: %s", region))
	}

	// Add localization feature evidence
	for _, feature := range features {
		evidence = append(evidence, fmt.Sprintf("Localization feature: %s - %s", feature.Name, feature.Value))
	}

	// Add international indicators evidence
	for _, indicator := range ipa.config.InternationalIndicators {
		if strings.Contains(content, indicator) {
			evidence = append(evidence, fmt.Sprintf("International indicator: %s", indicator))
		}
	}

	return evidence
}

// getLocalizationTypes extracts the types of localization features
func (ipa *InternationalPresenceAnalyzer) getLocalizationTypes(features []LocalizationFeature) []string {
	types := make(map[string]bool)
	for _, feature := range features {
		types[feature.Type] = true
	}

	var result []string
	for t := range types {
		result = append(result, t)
	}
	return result
}
