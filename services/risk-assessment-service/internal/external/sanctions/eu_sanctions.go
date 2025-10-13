package sanctions

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EUSanctionsClient provides EU sanctions screening
type EUSanctionsClient struct {
	logger *zap.Logger
	config *EUSanctionsConfig
}

// EUSanctionsConfig holds configuration for EU sanctions API
type EUSanctionsConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// EUSanctionsList represents an EU sanctions list entry
type EUSanctionsList struct {
	EntityID         string     `json:"entity_id"`
	EntityName       string     `json:"entity_name"`
	EntityType       string     `json:"entity_type"` // "individual", "entity", "vessel", "aircraft"
	Country          string     `json:"country"`
	Nationality      string     `json:"nationality,omitempty"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	PlaceOfBirth     string     `json:"place_of_birth,omitempty"`
	PassportNumber   string     `json:"passport_number,omitempty"`
	NationalID       string     `json:"national_id,omitempty"`
	Address          string     `json:"address,omitempty"`
	SanctionsProgram string     `json:"sanctions_program"`
	ProgramList      string     `json:"program_list"`
	Title            string     `json:"title,omitempty"`
	Remarks          string     `json:"remarks,omitempty"`
	EffectiveDate    time.Time  `json:"effective_date"`
	LastUpdated      time.Time  `json:"last_updated"`
	MatchScore       float64    `json:"match_score"`
	RiskLevel        string     `json:"risk_level"` // "high", "medium", "low"
	Source           string     `json:"source"`
	EUMemberState    string     `json:"eu_member_state,omitempty"`
}

// EUSanctionsSearchResult represents the result of an EU sanctions search
type EUSanctionsSearchResult struct {
	Query        string            `json:"query"`
	TotalMatches int               `json:"total_matches"`
	Matches      []EUSanctionsList `json:"matches"`
	SearchTime   time.Duration     `json:"search_time"`
	LastChecked  time.Time         `json:"last_checked"`
	DataQuality  string            `json:"data_quality"`
	Source       string            `json:"source"`
}

// NewEUSanctionsClient creates a new EU sanctions client
func NewEUSanctionsClient(config *EUSanctionsConfig, logger *zap.Logger) *EUSanctionsClient {
	return &EUSanctionsClient{
		logger: logger,
		config: config,
	}
}

// SearchSanctions searches for entities in EU sanctions lists
func (eu *EUSanctionsClient) SearchSanctions(ctx context.Context, entityName, country string) (*EUSanctionsSearchResult, error) {
	startTime := time.Now()
	eu.logger.Info("Searching EU sanctions lists (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(400)+200) * time.Millisecond)

	// Generate mock EU sanctions search results
	matches := eu.generateEUSanctionsMatches(entityName, country)

	result := &EUSanctionsSearchResult{
		Query:        entityName,
		TotalMatches: len(matches),
		Matches:      matches,
		SearchTime:   time.Since(startTime),
		LastChecked:  time.Now(),
		DataQuality:  eu.generateDataQuality(),
		Source:       "European Union",
	}

	eu.logger.Info("EU sanctions search completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", result.TotalMatches),
		zap.Duration("search_time", result.SearchTime))

	return result, nil
}

// GetSanctionsList retrieves the complete EU sanctions list
func (eu *EUSanctionsClient) GetSanctionsList(ctx context.Context) ([]EUSanctionsList, error) {
	eu.logger.Info("Retrieving complete EU sanctions list (mock)")

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	// Generate mock complete EU sanctions list
	var sanctionsList []EUSanctionsList
	numEntries := rand.Intn(800) + 400 // 400-1200 entries

	for i := 0; i < numEntries; i++ {
		entry := EUSanctionsList{
			EntityID:         fmt.Sprintf("EU_%d", rand.Intn(100000)),
			EntityName:       eu.generateEntityName(),
			EntityType:       eu.generateEntityType(),
			Country:          eu.generateCountry(),
			Nationality:      eu.generateNationality(),
			SanctionsProgram: eu.generateSanctionsProgram(),
			ProgramList:      eu.generateProgramList(),
			Title:            eu.generateTitle(),
			Remarks:          eu.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       1.0,
			RiskLevel:        "high",
			Source:           "European Union",
			EUMemberState:    eu.generateEUMemberState(),
		}

		// Add optional fields for individuals
		if entry.EntityType == "individual" {
			if rand.Float64() > 0.3 {
				dob := time.Now().Add(-time.Duration(rand.Intn(36500)) * 24 * time.Hour)
				entry.DateOfBirth = &dob
			}
			if rand.Float64() > 0.5 {
				entry.PlaceOfBirth = eu.generatePlaceOfBirth()
			}
			if rand.Float64() > 0.7 {
				entry.PassportNumber = eu.generatePassportNumber()
			}
		}

		// Add address for entities
		if entry.EntityType == "entity" && rand.Float64() > 0.4 {
			entry.Address = eu.generateAddress(entry.Country)
		}

		sanctionsList = append(sanctionsList, entry)
	}

	eu.logger.Info("EU sanctions list retrieved (mock)",
		zap.Int("total_entries", len(sanctionsList)))

	return sanctionsList, nil
}

// GetSanctionsByCountry retrieves sanctions entries for a specific country
func (eu *EUSanctionsClient) GetSanctionsByCountry(ctx context.Context, country string) ([]EUSanctionsList, error) {
	eu.logger.Info("Retrieving EU sanctions by country (mock)",
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(600)+300) * time.Millisecond)

	// Generate mock country-specific EU sanctions
	var sanctionsList []EUSanctionsList
	numEntries := rand.Intn(40) + 5 // 5-45 entries per country

	for i := 0; i < numEntries; i++ {
		entry := EUSanctionsList{
			EntityID:         fmt.Sprintf("EU_%s_%d", country, rand.Intn(10000)),
			EntityName:       eu.generateEntityName(),
			EntityType:       eu.generateEntityType(),
			Country:          country,
			Nationality:      eu.generateNationality(),
			SanctionsProgram: eu.generateSanctionsProgram(),
			ProgramList:      eu.generateProgramList(),
			Title:            eu.generateTitle(),
			Remarks:          eu.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       1.0,
			RiskLevel:        "high",
			Source:           "European Union",
			EUMemberState:    eu.generateEUMemberState(),
		}

		sanctionsList = append(sanctionsList, entry)
	}

	eu.logger.Info("EU sanctions by country retrieved (mock)",
		zap.String("country", country),
		zap.Int("total_entries", len(sanctionsList)))

	return sanctionsList, nil
}

// GetSanctionsByEUMemberState retrieves sanctions entries for a specific EU member state
func (eu *EUSanctionsClient) GetSanctionsByEUMemberState(ctx context.Context, memberState string) ([]EUSanctionsList, error) {
	eu.logger.Info("Retrieving EU sanctions by member state (mock)",
		zap.String("member_state", memberState))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(500)+250) * time.Millisecond)

	// Generate mock member state-specific EU sanctions
	var sanctionsList []EUSanctionsList
	numEntries := rand.Intn(30) + 5 // 5-35 entries per member state

	for i := 0; i < numEntries; i++ {
		entry := EUSanctionsList{
			EntityID:         fmt.Sprintf("EU_%s_%d", memberState, rand.Intn(10000)),
			EntityName:       eu.generateEntityName(),
			EntityType:       eu.generateEntityType(),
			Country:          eu.generateCountry(),
			Nationality:      eu.generateNationality(),
			SanctionsProgram: eu.generateSanctionsProgram(),
			ProgramList:      eu.generateProgramList(),
			Title:            eu.generateTitle(),
			Remarks:          eu.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       1.0,
			RiskLevel:        "high",
			Source:           "European Union",
			EUMemberState:    memberState,
		}

		sanctionsList = append(sanctionsList, entry)
	}

	eu.logger.Info("EU sanctions by member state retrieved (mock)",
		zap.String("member_state", memberState),
		zap.Int("total_entries", len(sanctionsList)))

	return sanctionsList, nil
}

// IsHealthy checks if the EU sanctions service is healthy
func (eu *EUSanctionsClient) IsHealthy(ctx context.Context) error {
	eu.logger.Info("Checking EU sanctions service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}

// GenerateRiskFactors generates risk factors from EU sanctions data
func (eu *EUSanctionsClient) GenerateRiskFactors(result *EUSanctionsSearchResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// EU sanctions risk factor
	sanctionsRisk := 0.1 // Base risk
	if result.TotalMatches > 0 {
		sanctionsRisk = 0.9 // High risk if EU sanctions matches found
	}

	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryCompliance,
		Subcategory: "eu_sanctions",
		Name:        "eu_sanctions_risk",
		Score:       sanctionsRisk,
		Weight:      0.45,
		Description: "Risk associated with EU sanctions list matches",
		Source:      "eu_sanctions",
		Confidence:  0.95,
		Impact:      "EU sanctions violations can result in severe penalties and business restrictions",
		Mitigation:  "Immediate compliance review and potential business relationship termination",
		LastUpdated: &now,
	})

	return riskFactors
}

// Helper methods for generating mock data

func (eu *EUSanctionsClient) generateEUSanctionsMatches(entityName, country string) []EUSanctionsList {
	var matches []EUSanctionsList

	// 97% of entities are not on EU sanctions lists
	if rand.Float64() > 0.03 {
		return matches
	}

	// Generate 1-2 matches for high-risk entities
	numMatches := rand.Intn(2) + 1
	for i := 0; i < numMatches; i++ {
		match := EUSanctionsList{
			EntityID:         fmt.Sprintf("EU_%d", rand.Intn(100000)),
			EntityName:       eu.generateSimilarName(entityName),
			EntityType:       eu.generateEntityType(),
			Country:          country,
			Nationality:      eu.generateNationality(),
			SanctionsProgram: eu.generateSanctionsProgram(),
			ProgramList:      eu.generateProgramList(),
			Title:            eu.generateTitle(),
			Remarks:          eu.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       rand.Float64()*0.3 + 0.7, // 0.7-1.0
			RiskLevel:        "high",
			Source:           "European Union",
			EUMemberState:    eu.generateEUMemberState(),
		}

		// Add optional fields for individuals
		if match.EntityType == "individual" {
			if rand.Float64() > 0.3 {
				dob := time.Now().Add(-time.Duration(rand.Intn(36500)) * 24 * time.Hour)
				match.DateOfBirth = &dob
			}
			if rand.Float64() > 0.5 {
				match.PlaceOfBirth = eu.generatePlaceOfBirth()
			}
		}

		matches = append(matches, match)
	}

	return matches
}

func (eu *EUSanctionsClient) generateEntityName() string {
	names := []string{
		"John Smith", "Jane Doe", "Ahmed Hassan", "Maria Garcia",
		"Vladimir Petrov", "Chen Wei", "Mohammed Al-Rashid",
		"Acme Corporation", "Global Trading Ltd", "International Holdings",
		"Smith & Associates", "Johnson Enterprises", "Williams Group",
		"European Trading Co", "Continental Holdings", "Mediterranean Corp",
	}
	return names[rand.Intn(len(names))]
}

func (eu *EUSanctionsClient) generateSimilarName(originalName string) string {
	variations := []string{
		originalName + " Ltd",
		originalName + " Inc",
		originalName + " Corp",
		"The " + originalName,
		originalName + " Group",
		originalName + " Holdings",
		originalName + " International",
		originalName + " Europe",
		originalName + " Continental",
	}
	return variations[rand.Intn(len(variations))]
}

func (eu *EUSanctionsClient) generateEntityType() string {
	types := []string{"individual", "entity", "vessel", "aircraft"}
	weights := []float64{0.6, 0.3, 0.05, 0.05} // 60% individual, 30% entity, 5% vessel, 5% aircraft

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return types[i]
		}
	}
	return "individual"
}

func (eu *EUSanctionsClient) generateCountry() string {
	countries := []string{
		"US", "UK", "CA", "AU", "DE", "FR", "IT", "ES", "NL", "SE",
		"RU", "CN", "JP", "KR", "IN", "BR", "MX", "AR", "ZA", "NG",
		"EG", "SA", "AE", "TR", "IR", "IQ", "AF", "PK", "BD", "ID",
		"BE", "AT", "FI", "DK", "NO", "CH", "PL", "CZ", "HU", "RO",
		"BG", "HR", "SI", "SK", "LT", "LV", "EE", "CY", "MT", "LU",
		"IE", "PT", "GR", "LV", "EE", "LT", "PL", "CZ", "HU", "RO",
	}
	return countries[rand.Intn(len(countries))]
}

func (eu *EUSanctionsClient) generateEUMemberState() string {
	memberStates := []string{
		"Germany", "France", "Italy", "Spain", "Netherlands", "Sweden",
		"Belgium", "Austria", "Finland", "Denmark", "Poland", "Czech Republic",
		"Hungary", "Romania", "Bulgaria", "Croatia", "Slovenia", "Slovakia",
		"Lithuania", "Latvia", "Estonia", "Cyprus", "Malta", "Luxembourg",
		"Ireland", "Portugal", "Greece",
	}
	return memberStates[rand.Intn(len(memberStates))]
}

func (eu *EUSanctionsClient) generateNationality() string {
	nationalities := []string{
		"American", "British", "Canadian", "Australian", "German", "French",
		"Italian", "Spanish", "Dutch", "Swedish", "Russian", "Chinese",
		"Japanese", "Korean", "Indian", "Brazilian", "Mexican", "Argentine",
		"South African", "Nigerian", "Egyptian", "Saudi", "Emirati",
		"Turkish", "Iranian", "Iraqi", "Afghan", "Pakistani", "Bangladeshi",
		"Indonesian", "Belgian", "Austrian", "Finnish", "Danish", "Norwegian",
		"Swiss", "Polish", "Czech", "Hungarian", "Romanian", "Bulgarian",
		"Croatian", "Slovenian", "Slovak", "Lithuanian", "Latvian", "Estonian",
		"Cypriot", "Maltese", "Luxembourgish", "Irish", "Portuguese", "Greek",
	}
	return nationalities[rand.Intn(len(nationalities))]
}

func (eu *EUSanctionsClient) generateSanctionsProgram() string {
	programs := []string{
		"Russia Sanctions",
		"Belarus Sanctions",
		"Iran Sanctions",
		"North Korea Sanctions",
		"Syria Sanctions",
		"Libya Sanctions",
		"Somalia Sanctions",
		"Central African Republic Sanctions",
		"Democratic Republic of the Congo Sanctions",
		"Mali Sanctions",
		"South Sudan Sanctions",
		"Yemen Sanctions",
		"Iraq Sanctions",
		"Lebanon Sanctions",
		"Guinea-Bissau Sanctions",
		"Guinea Sanctions",
		"Côte d'Ivoire Sanctions",
		"Liberia Sanctions",
		"Sierra Leone Sanctions",
		"Eritrea Sanctions",
		"Ethiopia Sanctions",
		"Burundi Sanctions",
		"Zimbabwe Sanctions",
		"Madagascar Sanctions",
		"Comoros Sanctions",
		"Rwanda Sanctions",
		"Uganda Sanctions",
		"Kenya Sanctions",
		"Venezuela Sanctions",
		"Nicaragua Sanctions",
		"Turkey Sanctions",
		"China Sanctions",
		"Myanmar Sanctions",
		"Afghanistan Sanctions",
		"Terrorism Sanctions",
		"Cyber Sanctions",
		"Chemical Weapons Sanctions",
		"Nuclear Proliferation Sanctions",
		"Human Rights Sanctions",
	}
	return programs[rand.Intn(len(programs))]
}

func (eu *EUSanctionsClient) generateProgramList() string {
	lists := []string{
		"EU Russia Sanctions List",
		"EU Belarus Sanctions List",
		"EU Iran Sanctions List",
		"EU North Korea Sanctions List",
		"EU Syria Sanctions List",
		"EU Libya Sanctions List",
		"EU Somalia Sanctions List",
		"EU Central African Republic Sanctions List",
		"EU Democratic Republic of the Congo Sanctions List",
		"EU Mali Sanctions List",
		"EU South Sudan Sanctions List",
		"EU Yemen Sanctions List",
		"EU Iraq Sanctions List",
		"EU Lebanon Sanctions List",
		"EU Guinea-Bissau Sanctions List",
		"EU Guinea Sanctions List",
		"EU Côte d'Ivoire Sanctions List",
		"EU Liberia Sanctions List",
		"EU Sierra Leone Sanctions List",
		"EU Eritrea Sanctions List",
		"EU Ethiopia Sanctions List",
		"EU Burundi Sanctions List",
		"EU Zimbabwe Sanctions List",
		"EU Madagascar Sanctions List",
		"EU Comoros Sanctions List",
		"EU Rwanda Sanctions List",
		"EU Uganda Sanctions List",
		"EU Kenya Sanctions List",
		"EU Venezuela Sanctions List",
		"EU Nicaragua Sanctions List",
		"EU Turkey Sanctions List",
		"EU China Sanctions List",
		"EU Myanmar Sanctions List",
		"EU Afghanistan Sanctions List",
		"EU Terrorism Sanctions List",
		"EU Cyber Sanctions List",
		"EU Chemical Weapons Sanctions List",
		"EU Nuclear Proliferation Sanctions List",
		"EU Human Rights Sanctions List",
	}
	return lists[rand.Intn(len(lists))]
}

func (eu *EUSanctionsClient) generateTitle() string {
	titles := []string{
		"Chief Executive Officer",
		"President",
		"Director",
		"Manager",
		"Owner",
		"Minister",
		"Ambassador",
		"General",
		"Colonel",
		"Captain",
		"Commissioner",
		"Secretary",
		"",
	}
	return titles[rand.Intn(len(titles))]
}

func (eu *EUSanctionsClient) generateRemarks() string {
	remarks := []string{
		"Subject to EU sanctions",
		"Designated for sanctions violations",
		"Associated with sanctioned activities",
		"Blocked pursuant to EU Council Decision",
		"Subject to EU restrictive measures",
		"",
	}
	return remarks[rand.Intn(len(remarks))]
}

func (eu *EUSanctionsClient) generatePlaceOfBirth() string {
	places := []string{
		"New York, USA", "London, UK", "Toronto, Canada", "Sydney, Australia",
		"Berlin, Germany", "Paris, France", "Rome, Italy", "Madrid, Spain",
		"Moscow, Russia", "Beijing, China", "Tokyo, Japan", "Seoul, South Korea",
		"Mumbai, India", "São Paulo, Brazil", "Mexico City, Mexico",
		"Buenos Aires, Argentina", "Cape Town, South Africa", "Lagos, Nigeria",
		"Cairo, Egypt", "Riyadh, Saudi Arabia", "Dubai, UAE", "Istanbul, Turkey",
		"Tehran, Iran", "Baghdad, Iraq", "Kabul, Afghanistan", "Islamabad, Pakistan",
		"Dhaka, Bangladesh", "Jakarta, Indonesia", "Brussels, Belgium", "Vienna, Austria",
		"Helsinki, Finland", "Copenhagen, Denmark", "Warsaw, Poland", "Prague, Czech Republic",
		"Budapest, Hungary", "Bucharest, Romania", "Sofia, Bulgaria", "Zagreb, Croatia",
		"Ljubljana, Slovenia", "Bratislava, Slovakia", "Vilnius, Lithuania", "Riga, Latvia",
		"Tallinn, Estonia", "Nicosia, Cyprus", "Valletta, Malta", "Luxembourg City, Luxembourg",
		"Dublin, Ireland", "Lisbon, Portugal", "Athens, Greece",
	}
	return places[rand.Intn(len(places))]
}

func (eu *EUSanctionsClient) generatePassportNumber() string {
	return fmt.Sprintf("P%d%d%d%d%d%d%d%d%d",
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10),
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))
}

func (eu *EUSanctionsClient) generateAddress(country string) string {
	addresses := map[string][]string{
		"US": {"123 Main St, New York, NY 10001", "456 Business Ave, San Francisco, CA 94105"},
		"UK": {"10 Downing Street, London SW1A 2AA", "25 Business Park, Manchester M1 1AA"},
		"CA": {"100 Bay Street, Toronto, ON M5H 2Y2", "2000 McGill College, Montreal, QC H3A 3H3"},
		"DE": {"Unter den Linden 1, 10117 Berlin", "Maximilianstraße 1, 80539 München"},
		"FR": {"1 Place de la Concorde, 75001 Paris", "1 Cours Mirabeau, 13100 Aix-en-Provence"},
		"IT": {"Via del Corso 1, 00186 Roma", "Piazza del Duomo 1, 20122 Milano"},
		"ES": {"Plaza Mayor 1, 28012 Madrid", "Plaça de Catalunya 1, 08002 Barcelona"},
		"NL": {"Dam 1, 1012 JS Amsterdam", "Coolsingel 1, 3011 AD Rotterdam"},
	}

	if countryAddresses, exists := addresses[country]; exists {
		return countryAddresses[rand.Intn(len(countryAddresses))]
	}
	return "123 International Blvd, Global City, GC 12345"
}

func (eu *EUSanctionsClient) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}
