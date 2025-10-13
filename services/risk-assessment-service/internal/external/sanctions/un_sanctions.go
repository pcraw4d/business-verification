package sanctions

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// UNSanctionsClient provides UN Security Council sanctions screening
type UNSanctionsClient struct {
	logger *zap.Logger
	config *UNSanctionsConfig
}

// UNSanctionsConfig holds configuration for UN sanctions API
type UNSanctionsConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// UNSanctionsList represents a UN sanctions list entry
type UNSanctionsList struct {
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
}

// UNSanctionsSearchResult represents the result of a UN sanctions search
type UNSanctionsSearchResult struct {
	Query        string            `json:"query"`
	TotalMatches int               `json:"total_matches"`
	Matches      []UNSanctionsList `json:"matches"`
	SearchTime   time.Duration     `json:"search_time"`
	LastChecked  time.Time         `json:"last_checked"`
	DataQuality  string            `json:"data_quality"`
	Source       string            `json:"source"`
}

// NewUNSanctionsClient creates a new UN sanctions client
func NewUNSanctionsClient(config *UNSanctionsConfig, logger *zap.Logger) *UNSanctionsClient {
	return &UNSanctionsClient{
		logger: logger,
		config: config,
	}
}

// SearchSanctions searches for entities in UN sanctions lists
func (un *UNSanctionsClient) SearchSanctions(ctx context.Context, entityName, country string) (*UNSanctionsSearchResult, error) {
	startTime := time.Now()
	un.logger.Info("Searching UN sanctions lists (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(400)+200) * time.Millisecond)

	// Generate mock UN sanctions search results
	matches := un.generateUNSanctionsMatches(entityName, country)

	result := &UNSanctionsSearchResult{
		Query:        entityName,
		TotalMatches: len(matches),
		Matches:      matches,
		SearchTime:   time.Since(startTime),
		LastChecked:  time.Now(),
		DataQuality:  un.generateDataQuality(),
		Source:       "UN Security Council",
	}

	un.logger.Info("UN sanctions search completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", result.TotalMatches),
		zap.Duration("search_time", result.SearchTime))

	return result, nil
}

// GetSanctionsList retrieves the complete UN sanctions list
func (un *UNSanctionsClient) GetSanctionsList(ctx context.Context) ([]UNSanctionsList, error) {
	un.logger.Info("Retrieving complete UN sanctions list (mock)")

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	// Generate mock complete sanctions list
	var sanctionsList []UNSanctionsList
	numEntries := rand.Intn(1000) + 500 // 500-1500 entries

	for i := 0; i < numEntries; i++ {
		entry := UNSanctionsList{
			EntityID:         fmt.Sprintf("UN_%d", rand.Intn(100000)),
			EntityName:       un.generateEntityName(),
			EntityType:       un.generateEntityType(),
			Country:          un.generateCountry(),
			Nationality:      un.generateNationality(),
			SanctionsProgram: un.generateSanctionsProgram(),
			ProgramList:      un.generateProgramList(),
			Title:            un.generateTitle(),
			Remarks:          un.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       1.0,
			RiskLevel:        "high",
			Source:           "UN Security Council",
		}

		// Add optional fields for individuals
		if entry.EntityType == "individual" {
			if rand.Float64() > 0.3 {
				dob := time.Now().Add(-time.Duration(rand.Intn(36500)) * 24 * time.Hour)
				entry.DateOfBirth = &dob
			}
			if rand.Float64() > 0.5 {
				entry.PlaceOfBirth = un.generatePlaceOfBirth()
			}
			if rand.Float64() > 0.7 {
				entry.PassportNumber = un.generatePassportNumber()
			}
		}

		// Add address for entities
		if entry.EntityType == "entity" && rand.Float64() > 0.4 {
			entry.Address = un.generateAddress(entry.Country)
		}

		sanctionsList = append(sanctionsList, entry)
	}

	un.logger.Info("UN sanctions list retrieved (mock)",
		zap.Int("total_entries", len(sanctionsList)))

	return sanctionsList, nil
}

// GetSanctionsByCountry retrieves sanctions entries for a specific country
func (un *UNSanctionsClient) GetSanctionsByCountry(ctx context.Context, country string) ([]UNSanctionsList, error) {
	un.logger.Info("Retrieving UN sanctions by country (mock)",
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(600)+300) * time.Millisecond)

	// Generate mock country-specific sanctions
	var sanctionsList []UNSanctionsList
	numEntries := rand.Intn(50) + 10 // 10-60 entries per country

	for i := 0; i < numEntries; i++ {
		entry := UNSanctionsList{
			EntityID:         fmt.Sprintf("UN_%s_%d", country, rand.Intn(10000)),
			EntityName:       un.generateEntityName(),
			EntityType:       un.generateEntityType(),
			Country:          country,
			Nationality:      un.generateNationality(),
			SanctionsProgram: un.generateSanctionsProgram(),
			ProgramList:      un.generateProgramList(),
			Title:            un.generateTitle(),
			Remarks:          un.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       1.0,
			RiskLevel:        "high",
			Source:           "UN Security Council",
		}

		sanctionsList = append(sanctionsList, entry)
	}

	un.logger.Info("UN sanctions by country retrieved (mock)",
		zap.String("country", country),
		zap.Int("total_entries", len(sanctionsList)))

	return sanctionsList, nil
}

// IsHealthy checks if the UN sanctions service is healthy
func (un *UNSanctionsClient) IsHealthy(ctx context.Context) error {
	un.logger.Info("Checking UN sanctions service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}

// GenerateRiskFactors generates risk factors from UN sanctions data
func (un *UNSanctionsClient) GenerateRiskFactors(result *UNSanctionsSearchResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// UN sanctions risk factor
	sanctionsRisk := 0.1 // Base risk
	if result.TotalMatches > 0 {
		sanctionsRisk = 0.95 // Critical risk if UN sanctions matches found
	}

	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryCompliance,
		Subcategory: "un_sanctions",
		Name:        "un_sanctions_risk",
		Score:       sanctionsRisk,
		Weight:      0.5,
		Description: "Risk associated with UN Security Council sanctions list matches",
		Source:      "un_sanctions",
		Confidence:  0.98,
		Impact:      "UN sanctions violations can result in severe international penalties",
		Mitigation:  "Immediate compliance review and potential business relationship termination",
		LastUpdated: &now,
	})

	return riskFactors
}

// Helper methods for generating mock data

func (un *UNSanctionsClient) generateUNSanctionsMatches(entityName, country string) []UNSanctionsList {
	var matches []UNSanctionsList

	// 98% of entities are not on UN sanctions lists
	if rand.Float64() > 0.02 {
		return matches
	}

	// Generate 1-3 matches for high-risk entities
	numMatches := rand.Intn(3) + 1
	for i := 0; i < numMatches; i++ {
		match := UNSanctionsList{
			EntityID:         fmt.Sprintf("UN_%d", rand.Intn(100000)),
			EntityName:       un.generateSimilarName(entityName),
			EntityType:       un.generateEntityType(),
			Country:          country,
			Nationality:      un.generateNationality(),
			SanctionsProgram: un.generateSanctionsProgram(),
			ProgramList:      un.generateProgramList(),
			Title:            un.generateTitle(),
			Remarks:          un.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			MatchScore:       rand.Float64()*0.3 + 0.7, // 0.7-1.0
			RiskLevel:        "high",
			Source:           "UN Security Council",
		}

		// Add optional fields for individuals
		if match.EntityType == "individual" {
			if rand.Float64() > 0.3 {
				dob := time.Now().Add(-time.Duration(rand.Intn(36500)) * 24 * time.Hour)
				match.DateOfBirth = &dob
			}
			if rand.Float64() > 0.5 {
				match.PlaceOfBirth = un.generatePlaceOfBirth()
			}
		}

		matches = append(matches, match)
	}

	return matches
}

func (un *UNSanctionsClient) generateEntityName() string {
	names := []string{
		"John Smith", "Jane Doe", "Ahmed Hassan", "Maria Garcia",
		"Vladimir Petrov", "Chen Wei", "Mohammed Al-Rashid",
		"Acme Corporation", "Global Trading Ltd", "International Holdings",
		"Smith & Associates", "Johnson Enterprises", "Williams Group",
	}
	return names[rand.Intn(len(names))]
}

func (un *UNSanctionsClient) generateSimilarName(originalName string) string {
	variations := []string{
		originalName + " Ltd",
		originalName + " Inc",
		originalName + " Corp",
		"The " + originalName,
		originalName + " Group",
		originalName + " Holdings",
		originalName + " International",
	}
	return variations[rand.Intn(len(variations))]
}

func (un *UNSanctionsClient) generateEntityType() string {
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

func (un *UNSanctionsClient) generateCountry() string {
	countries := []string{
		"US", "UK", "CA", "AU", "DE", "FR", "IT", "ES", "NL", "SE",
		"RU", "CN", "JP", "KR", "IN", "BR", "MX", "AR", "ZA", "NG",
		"EG", "SA", "AE", "TR", "IR", "IQ", "AF", "PK", "BD", "ID",
	}
	return countries[rand.Intn(len(countries))]
}

func (un *UNSanctionsClient) generateNationality() string {
	nationalities := []string{
		"American", "British", "Canadian", "Australian", "German", "French",
		"Italian", "Spanish", "Dutch", "Swedish", "Russian", "Chinese",
		"Japanese", "Korean", "Indian", "Brazilian", "Mexican", "Argentine",
		"South African", "Nigerian", "Egyptian", "Saudi", "Emirati",
		"Turkish", "Iranian", "Iraqi", "Afghan", "Pakistani", "Bangladeshi",
		"Indonesian",
	}
	return nationalities[rand.Intn(len(nationalities))]
}

func (un *UNSanctionsClient) generateSanctionsProgram() string {
	programs := []string{
		"Al-Qaida Sanctions",
		"Taliban Sanctions",
		"ISIL (Da'esh) and Al-Qaida Sanctions",
		"Iran Sanctions",
		"North Korea Sanctions",
		"Libya Sanctions",
		"Sudan Sanctions",
		"Somalia Sanctions",
		"Central African Republic Sanctions",
		"Democratic Republic of the Congo Sanctions",
		"Mali Sanctions",
		"South Sudan Sanctions",
		"Yemen Sanctions",
		"Syria Sanctions",
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
	}
	return programs[rand.Intn(len(programs))]
}

func (un *UNSanctionsClient) generateProgramList() string {
	lists := []string{
		"Al-Qaida Sanctions List",
		"Taliban Sanctions List",
		"ISIL (Da'esh) and Al-Qaida Sanctions List",
		"Iran Sanctions List",
		"North Korea Sanctions List",
		"Libya Sanctions List",
		"Sudan Sanctions List",
		"Somalia Sanctions List",
		"Central African Republic Sanctions List",
		"Democratic Republic of the Congo Sanctions List",
	}
	return lists[rand.Intn(len(lists))]
}

func (un *UNSanctionsClient) generateTitle() string {
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
		"",
	}
	return titles[rand.Intn(len(titles))]
}

func (un *UNSanctionsClient) generateRemarks() string {
	remarks := []string{
		"Subject to UN Security Council sanctions",
		"Designated for sanctions violations",
		"Associated with sanctioned activities",
		"Blocked pursuant to UN Security Council resolution",
		"",
	}
	return remarks[rand.Intn(len(remarks))]
}

func (un *UNSanctionsClient) generatePlaceOfBirth() string {
	places := []string{
		"New York, USA", "London, UK", "Toronto, Canada", "Sydney, Australia",
		"Berlin, Germany", "Paris, France", "Rome, Italy", "Madrid, Spain",
		"Moscow, Russia", "Beijing, China", "Tokyo, Japan", "Seoul, South Korea",
		"Mumbai, India", "São Paulo, Brazil", "Mexico City, Mexico",
		"Buenos Aires, Argentina", "Cape Town, South Africa", "Lagos, Nigeria",
		"Cairo, Egypt", "Riyadh, Saudi Arabia", "Dubai, UAE", "Istanbul, Turkey",
		"Tehran, Iran", "Baghdad, Iraq", "Kabul, Afghanistan", "Islamabad, Pakistan",
		"Dhaka, Bangladesh", "Jakarta, Indonesia",
	}
	return places[rand.Intn(len(places))]
}

func (un *UNSanctionsClient) generatePassportNumber() string {
	return fmt.Sprintf("P%d%d%d%d%d%d%d%d%d",
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10),
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))
}

func (un *UNSanctionsClient) generateAddress(country string) string {
	addresses := map[string][]string{
		"US": {"123 Main St, New York, NY 10001", "456 Business Ave, San Francisco, CA 94105"},
		"UK": {"10 Downing Street, London SW1A 2AA", "25 Business Park, Manchester M1 1AA"},
		"CA": {"100 Bay Street, Toronto, ON M5H 2Y2", "2000 McGill College, Montreal, QC H3A 3H3"},
	}

	if countryAddresses, exists := addresses[country]; exists {
		return countryAddresses[rand.Intn(len(countryAddresses))]
	}
	return "123 International Blvd, Global City, GC 12345"
}

func (un *UNSanctionsClient) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}
