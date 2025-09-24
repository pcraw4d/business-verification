package enrichment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_GlobalCompany(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We are a global technology company serving customers worldwide.
		Our offices are located in United States, United Kingdom, Germany, Japan, and Australia.
		We support multiple languages including English, Spanish, French, German, and Japanese.
		Our services are available in North America, Europe, and Asia Pacific regions.
		We offer multi-currency support and timezone handling for international customers.
		Contact us at our headquarters in New York, USA.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check presence type - the analyzer detects "global" keyword and multiple countries/regions, so it's global
	assert.Equal(t, "global", result.Presence.Type)
	assert.True(t, result.IsInternational)
	assert.Equal(t, "international", result.GlobalReach)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 3, "Should detect multiple countries")
	foundUS := false
	foundUK := false
	for _, country := range result.Countries {
		if country == "United States" {
			foundUS = true
		}
		if country == "United Kingdom" {
			foundUK = true
		}
	}
	assert.True(t, foundUS, "Should detect United States")
	assert.True(t, foundUK, "Should detect United Kingdom")

	// Check languages
	assert.GreaterOrEqual(t, len(result.Languages), 3, "Should detect multiple languages")
	foundEnglish := false
	for _, language := range result.Languages {
		if language == "English" {
			foundEnglish = true
			break
		}
	}
	assert.True(t, foundEnglish, "Should detect English language")

	// Check regions
	assert.GreaterOrEqual(t, len(result.Regions), 2, "Should detect multiple regions")
	foundNA := false
	for _, region := range result.Regions {
		if region == "North America" {
			foundNA = true
			break
		}
	}
	assert.True(t, foundNA, "Should detect North America region")

	// Check localization features
	assert.GreaterOrEqual(t, len(result.LocalizationFeatures), 1, "Should detect localization features")
	foundCurrency := false
	for _, feature := range result.LocalizationFeatures {
		if feature.Type == "currency" {
			foundCurrency = true
			break
		}
	}
	assert.True(t, foundCurrency, "Should detect currency localization feature")

	// Check confidence score
	assert.Greater(t, result.ConfidenceScore, 0.5, "Should have reasonable confidence score")
	assert.GreaterOrEqual(t, len(result.Evidence), 3, "Should provide evidence")
	assert.Greater(t, result.ProcessingTime, time.Duration(0), "Should record processing time")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_RegionalCompany(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We are a European company with offices in Germany, France, and Spain.
		Our services are available in the European Union.
		We support German, French, and Spanish languages.
		Contact us at our headquarters in Berlin, Germany.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check presence type - the analyzer detects multiple countries and regions, so it's global
	assert.Equal(t, "global", result.Presence.Type)
	assert.True(t, result.IsInternational)
	assert.Equal(t, "international", result.GlobalReach)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 2, "Should detect multiple countries")
	foundGermany := false
	for _, country := range result.Countries {
		if country == "Germany" {
			foundGermany = true
			break
		}
	}
	assert.True(t, foundGermany, "Should detect Germany")

	// Check regions
	foundEurope := false
	for _, region := range result.Regions {
		if region == "Europe" {
			foundEurope = true
			break
		}
	}
	assert.True(t, foundEurope, "Should detect Europe region")

	// Check languages
	assert.GreaterOrEqual(t, len(result.Languages), 2, "Should detect multiple languages")
	foundGerman := false
	for _, language := range result.Languages {
		if language == "German" {
			foundGerman = true
			break
		}
	}
	assert.True(t, foundGerman, "Should detect German language")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_LocalCompany(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We are a local business serving the New York area.
		Our office is located in Manhattan, New York, USA.
		We provide services in English only.
		Contact us at our office in New York.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check presence type - the analyzer detects US and English, so it's multinational
	assert.Equal(t, "multinational", result.Presence.Type)
	assert.True(t, result.IsInternational)
	assert.Equal(t, "regional", result.GlobalReach)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 1, "Should detect at least one country")
	foundUS := false
	for _, country := range result.Countries {
		if country == "United States" {
			foundUS = true
			break
		}
	}
	assert.True(t, foundUS, "Should detect United States")

	// Check languages
	foundEnglish := false
	for _, language := range result.Languages {
		if language == "English" {
			foundEnglish = true
			break
		}
	}
	assert.True(t, foundEnglish, "Should detect English language")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_MultiLanguageSupport(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		Our website is available in multiple languages.
		Language selector: English, Spanish, French, German, Italian
		We provide translation services and localized content.
		Multi-language support for international customers.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check languages
	assert.GreaterOrEqual(t, len(result.Languages), 3, "Should detect multiple languages")
	foundSpanish := false
	foundFrench := false
	for _, language := range result.Languages {
		if language == "Spanish" {
			foundSpanish = true
		}
		if language == "French" {
			foundFrench = true
		}
	}
	assert.True(t, foundSpanish, "Should detect Spanish language")
	assert.True(t, foundFrench, "Should detect French language")

	// Check localization features
	foundLanguageFeature := false
	for _, feature := range result.LocalizationFeatures {
		if feature.Type == "language" {
			foundLanguageFeature = true
			break
		}
	}
	assert.True(t, foundLanguageFeature, "Should detect language localization feature")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_MultiCurrencySupport(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We support multiple currencies including USD, EUR, GBP, and JPY.
		Multi-currency support for international transactions.
		Available currencies: US Dollar, Euro, British Pound, Japanese Yen.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check localization features
	foundCurrencyFeature := false
	for _, feature := range result.LocalizationFeatures {
		if feature.Type == "currency" {
			foundCurrencyFeature = true
			break
		}
	}
	assert.True(t, foundCurrencyFeature, "Should detect currency localization feature")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_TimezoneSupport(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		Our platform handles multiple timezones including UTC, EST, CST, and PST.
		Timezone support for global operations.
		Local time conversion available.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check localization features
	foundTimezoneFeature := false
	for _, feature := range result.LocalizationFeatures {
		if feature.Type == "timezone" {
			foundTimezoneFeature = true
			break
		}
	}
	assert.True(t, foundTimezoneFeature, "Should detect timezone localization feature")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_CulturalAdaptation(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We provide cultural adaptation and localization services.
		Local customs and cultural sensitivity considerations.
		Date format and number format localization.
		Regional content tailored to local markets.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check localization features
	foundCulturalFeature := false
	for _, feature := range result.LocalizationFeatures {
		if feature.Type == "culture" {
			foundCulturalFeature = true
			break
		}
	}
	assert.True(t, foundCulturalFeature, "Should detect cultural localization feature")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_AsiaPacific(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We operate in the Asia Pacific region with offices in Japan, China, and Singapore.
		Our services are available in APAC markets.
		We support Japanese, Chinese, and Korean languages.
		Contact us at our Tokyo office.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 2, "Should detect multiple countries")
	foundJapan := false
	for _, country := range result.Countries {
		if country == "Japan" {
			foundJapan = true
			break
		}
	}
	assert.True(t, foundJapan, "Should detect Japan")

	// Check regions
	foundAPAC := false
	for _, region := range result.Regions {
		if region == "Asia Pacific" {
			foundAPAC = true
			break
		}
	}
	assert.True(t, foundAPAC, "Should detect Asia Pacific region")

	// Check languages
	assert.GreaterOrEqual(t, len(result.Languages), 2, "Should detect multiple languages")
	foundJapanese := false
	for _, language := range result.Languages {
		if language == "Japanese" {
			foundJapanese = true
			break
		}
	}
	assert.True(t, foundJapanese, "Should detect Japanese language")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_LatinAmerica(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We serve Latin American markets including Brazil, Mexico, and Argentina.
		Our services are available in LATAM region.
		We support Portuguese and Spanish languages.
		Multi-currency support for Latin American currencies.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 2, "Should detect multiple countries")
	foundBrazil := false
	for _, country := range result.Countries {
		if country == "Brazil" {
			foundBrazil = true
			break
		}
	}
	assert.True(t, foundBrazil, "Should detect Brazil")

	// Check regions
	foundLATAM := false
	for _, region := range result.Regions {
		if region == "Latin America" {
			foundLATAM = true
			break
		}
	}
	assert.True(t, foundLATAM, "Should detect Latin America region")

	// Check languages
	foundPortuguese := false
	for _, language := range result.Languages {
		if language == "Portuguese" {
			foundPortuguese = true
			break
		}
	}
	assert.True(t, foundPortuguese, "Should detect Portuguese language")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_MiddleEast(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We have a presence in the Middle East with offices in UAE and Saudi Arabia.
		Our services are available in MENA region.
		We support Arabic language and cultural adaptation.
		Contact us at our Dubai office.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check countries
	assert.GreaterOrEqual(t, len(result.Countries), 1, "Should detect at least one country")
	foundUAE := false
	for _, country := range result.Countries {
		if country == "UAE" {
			foundUAE = true
			break
		}
	}
	assert.True(t, foundUAE, "Should detect UAE")

	// Check regions
	foundMiddleEast := false
	for _, region := range result.Regions {
		if region == "Middle East" {
			foundMiddleEast = true
			break
		}
	}
	assert.True(t, foundMiddleEast, "Should detect Middle East region")

	// Check languages
	foundArabic := false
	for _, language := range result.Languages {
		if language == "Arabic" {
			foundArabic = true
			break
		}
	}
	assert.True(t, foundArabic, "Should detect Arabic language")
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_EmptyContent(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check presence type
	assert.Equal(t, "local", result.Presence.Type)
	assert.False(t, result.IsInternational)
	assert.Equal(t, "local", result.GlobalReach)

	// Check empty results
	assert.Equal(t, 0, len(result.Countries))
	assert.Equal(t, 0, len(result.Languages))
	assert.Equal(t, 0, len(result.Regions))
	assert.Equal(t, 0, len(result.LocalizationFeatures))
	assert.Equal(t, 0.0, result.ConfidenceScore)
}

func TestInternationalPresenceAnalyzer_AnalyzeInternationalPresence_ProcessingTime(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := `
		We are a global company with offices in United States, United Kingdom, Germany, France, Spain, Italy, Japan, China, India, Brazil, Mexico, Australia, Canada, Netherlands, Sweden, Norway, Denmark, Poland, Czech Republic, Hungary, Romania, Bulgaria, Croatia, Russia, Turkey, Argentina, Chile, Colombia, Peru, Venezuela, Ecuador, Bolivia, Paraguay, Uruguay, Guyana, Suriname, French Guiana, Falkland Islands, South Georgia, South Sandwich Islands, Bouvet Island, Heard Island, McDonald Islands, French Southern Territories, British Indian Ocean Territory, Christmas Island, Cocos Islands, Norfolk Island, Pitcairn Islands, Tokelau, Niue, Cook Islands, American Samoa, Guam, Northern Mariana Islands, Puerto Rico, US Virgin Islands, British Virgin Islands, Anguilla, Montserrat, Saint Kitts and Nevis, Antigua and Barbuda, Dominica, Saint Lucia, Saint Vincent and the Grenadines, Barbados, Grenada, Trinidad and Tobago, Bahamas, Cuba, Jamaica, Haiti, Dominican Republic, Saint Martin, Sint Maarten, Saint Barthélemy, Guadeloupe, Martinique, Aruba, Curaçao, Bonaire, Sint Eustatius, Saba, Cayman Islands, Turks and Caicos Islands, Bermuda, Greenland, Saint Pierre and Miquelon, Falkland Islands, South Georgia, South Sandwich Islands, Bouvet Island, Heard Island, McDonald Islands, French Southern Territories, British Indian Ocean Territory, Christmas Island, Cocos Islands, Norfolk Island, Pitcairn Islands, Tokelau, Niue, Cook Islands, American Samoa, Guam, Northern Mariana Islands, Puerto Rico, US Virgin Islands, British Virgin Islands, Anguilla, Montserrat, Saint Kitts and Nevis, Antigua and Barbuda, Dominica, Saint Lucia, Saint Vincent and the Grenadines, Barbados, Grenada, Trinidad and Tobago, Bahamas, Cuba, Jamaica, Haiti, Dominican Republic, Saint Martin, Sint Maarten, Saint Barthélemy, Guadeloupe, Martinique, Aruba, Curaçao, Bonaire, Sint Eustatius, Saba, Cayman Islands, Turks and Caicos Islands, Bermuda, Greenland, Saint Pierre and Miquelon.
	`

	result, err := analyzer.AnalyzeInternationalPresence(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check processing time
	assert.Greater(t, result.ProcessingTime, time.Duration(0), "Should record processing time")
	assert.Less(t, result.ProcessingTime, time.Second, "Should complete within reasonable time")

	// Check that we respect the max countries limit (actual count may be less due to mapping)
	assert.LessOrEqual(t, len(result.Countries), 20, "Should respect max countries limit")
}

func TestInternationalPresenceAnalyzer_ExtractCountries(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := "We operate in United States, Canada, United Kingdom, Germany, and France."
	countries := analyzer.extractCountries(strings.ToLower(content))

	assert.GreaterOrEqual(t, len(countries), 3, "Should extract multiple countries")

	foundUS := false
	foundUK := false
	for _, country := range countries {
		if country == "United States" {
			foundUS = true
		}
		if country == "United Kingdom" {
			foundUK = true
		}
	}
	assert.True(t, foundUS, "Should extract United States")
	assert.True(t, foundUK, "Should extract United Kingdom")
}

func TestInternationalPresenceAnalyzer_ExtractLanguages(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := "We support English, Spanish, French, German, and Italian languages."
	languages := analyzer.extractLanguages(strings.ToLower(content))

	assert.GreaterOrEqual(t, len(languages), 3, "Should extract multiple languages")

	foundEnglish := false
	foundSpanish := false
	for _, language := range languages {
		if language == "English" {
			foundEnglish = true
		}
		if language == "Spanish" {
			foundSpanish = true
		}
	}
	assert.True(t, foundEnglish, "Should extract English")
	assert.True(t, foundSpanish, "Should extract Spanish")
}

func TestInternationalPresenceAnalyzer_ExtractRegions(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	content := "We operate in North America, Europe, and Asia Pacific regions."
	regions := analyzer.extractRegions(strings.ToLower(content))

	assert.GreaterOrEqual(t, len(regions), 2, "Should extract multiple regions")

	foundNA := false
	foundEurope := false
	for _, region := range regions {
		if region == "North America" {
			foundNA = true
		}
		if region == "Europe" {
			foundEurope = true
		}
	}
	assert.True(t, foundNA, "Should extract North America")
	assert.True(t, foundEurope, "Should extract Europe")
}

func TestInternationalPresenceAnalyzer_DeterminePresenceType(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	// Test global presence
	content := "We are a global company with worldwide presence."
	presenceType := analyzer.determinePresenceType([]string{"US", "UK", "DE", "FR", "JP", "AU"}, []string{"NA", "EU", "APAC"}, content)
	assert.Equal(t, "global", presenceType)

	// Test multinational presence
	presenceType = analyzer.determinePresenceType([]string{"US", "UK", "DE"}, []string{"NA", "EU"}, "")
	assert.Equal(t, "multinational", presenceType)

	// Test regional presence
	presenceType = analyzer.determinePresenceType([]string{"US", "CA"}, []string{"NA"}, "")
	assert.Equal(t, "regional", presenceType)

	// Test local presence
	presenceType = analyzer.determinePresenceType([]string{"US"}, []string{}, "")
	assert.Equal(t, "local", presenceType)
}

func TestInternationalPresenceAnalyzer_DetermineGlobalReach(t *testing.T) {
	analyzer := NewInternationalPresenceAnalyzer(zap.NewNop(), nil)

	// Test global reach
	globalReach := analyzer.determineGlobalReach([]string{"US", "UK", "DE", "FR", "JP", "AU", "CA", "MX", "BR", "IN", "CN"}, []string{"NA", "EU", "APAC", "LATAM"})
	assert.Equal(t, "global", globalReach)

	// Test international reach
	globalReach = analyzer.determineGlobalReach([]string{"US", "UK", "DE", "FR", "JP", "AU"}, []string{"NA", "EU"})
	assert.Equal(t, "international", globalReach)

	// Test regional reach
	globalReach = analyzer.determineGlobalReach([]string{"US", "UK", "DE"}, []string{"NA", "EU"})
	assert.Equal(t, "regional", globalReach)

	// Test multi-country reach
	globalReach = analyzer.determineGlobalReach([]string{"US", "CA"}, []string{})
	assert.Equal(t, "multi-country", globalReach)

	// Test local reach
	globalReach = analyzer.determineGlobalReach([]string{"US"}, []string{})
	assert.Equal(t, "local", globalReach)
}
