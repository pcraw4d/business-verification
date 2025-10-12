package worldcheck

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// WorldCheckMock provides mock implementation of World-Check API
type WorldCheckMock struct {
	logger *zap.Logger
	config *WorldCheckConfig
}

// WorldCheckConfig holds configuration for World-Check API
type WorldCheckConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// WorldCheckProfile represents a World-Check profile
type WorldCheckProfile struct {
	ProfileID     string     `json:"profile_id"`
	EntityName    string     `json:"entity_name"`
	EntityType    string     `json:"entity_type"` // "individual", "entity", "vessel", "aircraft"
	Country       string     `json:"country"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	PlaceOfBirth  string     `json:"place_of_birth,omitempty"`
	Nationality   string     `json:"nationality,omitempty"`
	Address       string     `json:"address,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Email         string     `json:"email,omitempty"`
	Website       string     `json:"website,omitempty"`
	RiskLevel     string     `json:"risk_level"` // "high", "medium", "low"
	Category      string     `json:"category"`
	SubCategory   string     `json:"sub_category"`
	Source        string     `json:"source"`
	LastUpdated   time.Time  `json:"last_updated"`
	ProfileStatus string     `json:"profile_status"` // "active", "inactive", "deceased"
	MatchScore    float64    `json:"match_score"`
	Confidence    float64    `json:"confidence"`
}

// AdverseMedia represents adverse media information
type AdverseMedia struct {
	MediaID       string    `json:"media_id"`
	Title         string    `json:"title"`
	Source        string    `json:"source"`
	URL           string    `json:"url"`
	PublishedDate time.Time `json:"published_date"`
	Content       string    `json:"content"`
	Sentiment     string    `json:"sentiment"` // "negative", "neutral", "positive"
	Relevance     float64   `json:"relevance"`
	RiskLevel     string    `json:"risk_level"`
}

// PEPStatus represents Politically Exposed Person status
type PEPStatus struct {
	IsPEP           bool       `json:"is_pep"`
	PEPLevel        string     `json:"pep_level"` // "domestic", "foreign", "international"
	Position        string     `json:"position,omitempty"`
	Jurisdiction    string     `json:"jurisdiction,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	FamilyMembers   []string   `json:"family_members,omitempty"`
	CloseAssociates []string   `json:"close_associates,omitempty"`
}

// SanctionsInfo represents sanctions information
type SanctionsInfo struct {
	IsSanctioned     bool       `json:"is_sanctioned"`
	SanctionsList    string     `json:"sanctions_list,omitempty"`
	SanctionsProgram string     `json:"sanctions_program,omitempty"`
	EffectiveDate    *time.Time `json:"effective_date,omitempty"`
	ExpiryDate       *time.Time `json:"expiry_date,omitempty"`
	Reason           string     `json:"reason,omitempty"`
	Authority        string     `json:"authority,omitempty"`
}

// RiskAssessment represents risk assessment from World-Check
type RiskAssessment struct {
	OverallRiskScore float64  `json:"overall_risk_score"`
	FinancialRisk    float64  `json:"financial_risk"`
	ReputationalRisk float64  `json:"reputational_risk"`
	RegulatoryRisk   float64  `json:"regulatory_risk"`
	OperationalRisk  float64  `json:"operational_risk"`
	RiskFactors      []string `json:"risk_factors"`
	Recommendations  []string `json:"recommendations"`
}

// WorldCheckResult represents the combined result from World-Check
type WorldCheckResult struct {
	Profile        *WorldCheckProfile `json:"profile,omitempty"`
	AdverseMedia   []AdverseMedia     `json:"adverse_media,omitempty"`
	PEPStatus      *PEPStatus         `json:"pep_status,omitempty"`
	SanctionsInfo  *SanctionsInfo     `json:"sanctions_info,omitempty"`
	RiskAssessment *RiskAssessment    `json:"risk_assessment,omitempty"`
	DataQuality    string             `json:"data_quality"`
	LastChecked    time.Time          `json:"last_checked"`
	ProcessingTime time.Duration      `json:"processing_time"`
}

// NewWorldCheckMock creates a new World-Check mock client
func NewWorldCheckMock(config *WorldCheckConfig, logger *zap.Logger) *WorldCheckMock {
	return &WorldCheckMock{
		logger: logger,
		config: config,
	}
}

// SearchProfile searches for entities in World-Check database
func (wc *WorldCheckMock) SearchProfile(ctx context.Context, entityName, country string) (*WorldCheckProfile, error) {
	wc.logger.Info("Searching World-Check profile (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

	// Generate mock profile
	profile := &WorldCheckProfile{
		ProfileID:     wc.generateProfileID(entityName),
		EntityName:    entityName,
		EntityType:    wc.generateEntityType(),
		Country:       country,
		DateOfBirth:   wc.generateDateOfBirth(),
		PlaceOfBirth:  wc.generatePlaceOfBirth(country),
		Nationality:   wc.generateNationality(country),
		Address:       wc.generateAddress(country),
		Phone:         wc.generatePhone(country),
		Email:         wc.generateEmail(entityName),
		Website:       wc.generateWebsite(entityName),
		RiskLevel:     wc.generateRiskLevel(),
		Category:      wc.generateCategory(),
		SubCategory:   wc.generateSubCategory(),
		Source:        "World-Check Database",
		LastUpdated:   time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		ProfileStatus: wc.generateProfileStatus(),
		MatchScore:    wc.generateMatchScore(),
		Confidence:    wc.generateConfidence(),
	}

	wc.logger.Info("World-Check profile retrieved (mock)",
		zap.String("profile_id", profile.ProfileID),
		zap.String("risk_level", profile.RiskLevel))

	return profile, nil
}

// GetAdverseMedia retrieves adverse media information
func (wc *WorldCheckMock) GetAdverseMedia(ctx context.Context, entityName string) ([]AdverseMedia, error) {
	wc.logger.Info("Getting World-Check adverse media (mock)",
		zap.String("entity_name", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// Generate mock adverse media
	adverseMedia := wc.generateAdverseMedia(entityName)

	wc.logger.Info("World-Check adverse media retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Int("media_count", len(adverseMedia)))

	return adverseMedia, nil
}

// GetPEPStatus retrieves PEP (Politically Exposed Person) status
func (wc *WorldCheckMock) GetPEPStatus(ctx context.Context, entityName string) (*PEPStatus, error) {
	wc.logger.Info("Getting World-Check PEP status (mock)",
		zap.String("entity_name", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)

	// Generate mock PEP status
	pepStatus := &PEPStatus{
		IsPEP:           wc.generateIsPEP(),
		PEPLevel:        wc.generatePEPLevel(),
		Position:        wc.generatePosition(),
		Jurisdiction:    wc.generateJurisdiction(),
		StartDate:       wc.generateStartDate(),
		EndDate:         wc.generateEndDate(),
		FamilyMembers:   wc.generateFamilyMembers(),
		CloseAssociates: wc.generateCloseAssociates(),
	}

	wc.logger.Info("World-Check PEP status retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Bool("is_pep", pepStatus.IsPEP))

	return pepStatus, nil
}

// GetSanctionsInfo retrieves sanctions information
func (wc *WorldCheckMock) GetSanctionsInfo(ctx context.Context, entityName string) (*SanctionsInfo, error) {
	wc.logger.Info("Getting World-Check sanctions info (mock)",
		zap.String("entity_name", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(250)+100) * time.Millisecond)

	// Generate mock sanctions info
	sanctionsInfo := &SanctionsInfo{
		IsSanctioned:     wc.generateIsSanctioned(),
		SanctionsList:    wc.generateSanctionsList(),
		SanctionsProgram: wc.generateSanctionsProgram(),
		EffectiveDate:    wc.generateEffectiveDate(),
		ExpiryDate:       wc.generateExpiryDate(),
		Reason:           wc.generateReason(),
		Authority:        wc.generateAuthority(),
	}

	wc.logger.Info("World-Check sanctions info retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Bool("is_sanctioned", sanctionsInfo.IsSanctioned))

	return sanctionsInfo, nil
}

// GetRiskAssessment retrieves risk assessment
func (wc *WorldCheckMock) GetRiskAssessment(ctx context.Context, entityName string) (*RiskAssessment, error) {
	wc.logger.Info("Getting World-Check risk assessment (mock)",
		zap.String("entity_name", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(350)+100) * time.Millisecond)

	// Generate mock risk assessment
	riskAssessment := &RiskAssessment{
		OverallRiskScore: wc.generateOverallRiskScore(),
		FinancialRisk:    wc.generateFinancialRisk(),
		ReputationalRisk: wc.generateReputationalRisk(),
		RegulatoryRisk:   wc.generateRegulatoryRisk(),
		OperationalRisk:  wc.generateOperationalRisk(),
		RiskFactors:      wc.generateRiskFactors(),
		Recommendations:  wc.generateRecommendations(),
	}

	wc.logger.Info("World-Check risk assessment retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Float64("overall_risk_score", riskAssessment.OverallRiskScore))

	return riskAssessment, nil
}

// GetComprehensiveData retrieves all available data from World-Check
func (wc *WorldCheckMock) GetComprehensiveData(ctx context.Context, entityName, country string) (*WorldCheckResult, error) {
	startTime := time.Now()
	wc.logger.Info("Getting comprehensive World-Check data (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Get all data in parallel
	type result struct {
		profile        *WorldCheckProfile
		adverseMedia   []AdverseMedia
		pepStatus      *PEPStatus
		sanctionsInfo  *SanctionsInfo
		riskAssessment *RiskAssessment
		err            error
	}

	results := make(chan result, 5)

	// Get profile
	go func() {
		p, err := wc.SearchProfile(ctx, entityName, country)
		results <- result{profile: p, err: err}
	}()

	// Get adverse media
	go func() {
		am, err := wc.GetAdverseMedia(ctx, entityName)
		results <- result{adverseMedia: am, err: err}
	}()

	// Get PEP status
	go func() {
		ps, err := wc.GetPEPStatus(ctx, entityName)
		results <- result{pepStatus: ps, err: err}
	}()

	// Get sanctions info
	go func() {
		si, err := wc.GetSanctionsInfo(ctx, entityName)
		results <- result{sanctionsInfo: si, err: err}
	}()

	// Get risk assessment
	go func() {
		ra, err := wc.GetRiskAssessment(ctx, entityName)
		results <- result{riskAssessment: ra, err: err}
	}()

	// Collect results
	var comprehensiveResult result
	for i := 0; i < 5; i++ {
		r := <-results
		if r.err != nil {
			wc.logger.Warn("Failed to get some World-Check data",
				zap.Error(r.err))
		}
		if r.profile != nil {
			comprehensiveResult.profile = r.profile
		}
		if r.adverseMedia != nil {
			comprehensiveResult.adverseMedia = r.adverseMedia
		}
		if r.pepStatus != nil {
			comprehensiveResult.pepStatus = r.pepStatus
		}
		if r.sanctionsInfo != nil {
			comprehensiveResult.sanctionsInfo = r.sanctionsInfo
		}
		if r.riskAssessment != nil {
			comprehensiveResult.riskAssessment = r.riskAssessment
		}
	}

	// Create comprehensive result
	wcResult := &WorldCheckResult{
		Profile:        comprehensiveResult.profile,
		AdverseMedia:   comprehensiveResult.adverseMedia,
		PEPStatus:      comprehensiveResult.pepStatus,
		SanctionsInfo:  comprehensiveResult.sanctionsInfo,
		RiskAssessment: comprehensiveResult.riskAssessment,
		DataQuality:    wc.generateDataQuality(),
		LastChecked:    time.Now(),
		ProcessingTime: time.Since(startTime),
	}

	wc.logger.Info("Comprehensive World-Check data retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Duration("processing_time", wcResult.ProcessingTime),
		zap.String("data_quality", wcResult.DataQuality))

	return wcResult, nil
}

// GenerateRiskFactors generates risk factors from World-Check data
func (wc *WorldCheckMock) GenerateRiskFactors(result *WorldCheckResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Profile risk factors
	if result.Profile != nil {
		profileRisk := 0.2 // Base risk
		if result.Profile.RiskLevel == "high" {
			profileRisk = 0.8
		} else if result.Profile.RiskLevel == "medium" {
			profileRisk = 0.5
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryReputational,
			Subcategory: "profile",
			Name:        "worldcheck_profile_risk",
			Score:       profileRisk,
			Weight:      0.3,
			Description: "Risk associated with World-Check profile",
			Source:      "worldcheck",
			Confidence:  0.90,
			Impact:      "Profile risk affects business reputation",
			Mitigation:  "Monitor profile changes and implement due diligence",
			LastUpdated: &now,
		})
	}

	// Adverse media risk factors
	if len(result.AdverseMedia) > 0 {
		adverseMediaRisk := 0.3
		if len(result.AdverseMedia) > 5 {
			adverseMediaRisk = 0.8
		} else if len(result.AdverseMedia) > 2 {
			adverseMediaRisk = 0.6
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryReputational,
			Subcategory: "adverse_media",
			Name:        "adverse_media_risk",
			Score:       adverseMediaRisk,
			Weight:      0.25,
			Description: "Risk associated with adverse media coverage",
			Source:      "worldcheck",
			Confidence:  0.85,
			Impact:      "Adverse media can damage business reputation",
			Mitigation:  "Monitor media coverage and address issues promptly",
			LastUpdated: &now,
		})
	}

	// PEP risk factors
	if result.PEPStatus != nil && result.PEPStatus.IsPEP {
		pepRisk := 0.6 // Base PEP risk
		if result.PEPStatus.PEPLevel == "international" {
			pepRisk = 0.8
		} else if result.PEPStatus.PEPLevel == "foreign" {
			pepRisk = 0.7
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "pep",
			Name:        "pep_risk",
			Score:       pepRisk,
			Weight:      0.35,
			Description: "Risk associated with Politically Exposed Person status",
			Source:      "worldcheck",
			Confidence:  0.95,
			Impact:      "PEP status requires enhanced due diligence",
			Mitigation:  "Implement PEP-specific compliance procedures",
			LastUpdated: &now,
		})
	}

	// Sanctions risk factors
	if result.SanctionsInfo != nil && result.SanctionsInfo.IsSanctioned {
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "sanctions",
			Name:        "sanctions_risk",
			Score:       0.9,
			Weight:      0.4,
			Description: "Risk associated with sanctions status",
			Source:      "worldcheck",
			Confidence:  0.98,
			Impact:      "Sanctions violations can result in severe penalties",
			Mitigation:  "Immediate compliance review and reporting",
			LastUpdated: &now,
		})
	}

	// Risk assessment factors
	if result.RiskAssessment != nil {
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryFinancial,
			Subcategory: "overall_risk",
			Name:        "worldcheck_risk_assessment",
			Score:       result.RiskAssessment.OverallRiskScore,
			Weight:      0.3,
			Description: "Overall risk assessment from World-Check",
			Source:      "worldcheck",
			Confidence:  0.88,
			Impact:      "Comprehensive risk assessment from specialized database",
			Mitigation:  "Address specific risk areas identified in assessment",
			LastUpdated: &now,
		})
	}

	return riskFactors
}

// Helper methods for generating mock data

func (wc *WorldCheckMock) generateProfileID(entityName string) string {
	return fmt.Sprintf("WC_%s_%d", wc.sanitizeForID(entityName), time.Now().Unix())
}

func (wc *WorldCheckMock) generateEntityType() string {
	types := []string{"individual", "entity", "vessel", "aircraft"}
	return types[rand.Intn(len(types))]
}

func (wc *WorldCheckMock) generateDateOfBirth() *time.Time {
	if rand.Float64() > 0.3 { // 70% chance of having DOB
		dob := time.Now().Add(-time.Duration(rand.Intn(50)+18) * 365 * 24 * time.Hour)
		return &dob
	}
	return nil
}

func (wc *WorldCheckMock) generatePlaceOfBirth(country string) string {
	places := map[string][]string{
		"US": {"New York, NY", "Los Angeles, CA", "Chicago, IL", "Houston, TX"},
		"UK": {"London, England", "Manchester, England", "Birmingham, England", "Glasgow, Scotland"},
		"CA": {"Toronto, ON", "Montreal, QC", "Vancouver, BC", "Calgary, AB"},
	}

	if countryPlaces, exists := places[country]; exists {
		return countryPlaces[rand.Intn(len(countryPlaces))]
	}
	return "Unknown"
}

func (wc *WorldCheckMock) generateNationality(country string) string {
	nationalities := map[string][]string{
		"US": {"American", "US Citizen"},
		"UK": {"British", "UK Citizen"},
		"CA": {"Canadian", "Canadian Citizen"},
	}

	if countryNationalities, exists := nationalities[country]; exists {
		return countryNationalities[rand.Intn(len(countryNationalities))]
	}
	return "Unknown"
}

func (wc *WorldCheckMock) generateAddress(country string) string {
	addresses := map[string][]string{
		"US": {"123 Main St, New York, NY 10001", "456 Oak Ave, Los Angeles, CA 90210"},
		"UK": {"10 Downing Street, London SW1A 2AA", "25 Business Park, Manchester M1 1AA"},
		"CA": {"100 Bay Street, Toronto, ON M5H 2Y2", "2000 McGill College, Montreal, QC H3A 3H3"},
	}

	if countryAddresses, exists := addresses[country]; exists {
		return countryAddresses[rand.Intn(len(countryAddresses))]
	}
	return "Unknown Address"
}

func (wc *WorldCheckMock) generatePhone(country string) string {
	phoneFormats := map[string][]string{
		"US": {"+1-555-123-4567", "+1-212-555-0123"},
		"UK": {"+44-20-7946-0958", "+44-161-555-0123"},
		"CA": {"+1-416-555-0123", "+1-514-555-0456"},
	}

	if countryPhones, exists := phoneFormats[country]; exists {
		return countryPhones[rand.Intn(len(countryPhones))]
	}
	return "+1-555-000-0000"
}

func (wc *WorldCheckMock) generateEmail(entityName string) string {
	return fmt.Sprintf("contact@%s.com", wc.sanitizeForURL(entityName))
}

func (wc *WorldCheckMock) generateWebsite(entityName string) string {
	return fmt.Sprintf("https://www.%s.com", wc.sanitizeForURL(entityName))
}

func (wc *WorldCheckMock) generateRiskLevel() string {
	levels := []string{"low", "medium", "high"}
	weights := []float64{0.6, 0.3, 0.1} // 60% low, 30% medium, 10% high

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return levels[i]
		}
	}
	return "low"
}

func (wc *WorldCheckMock) generateCategory() string {
	categories := []string{
		"Financial Crime",
		"Terrorism",
		"Sanctions",
		"PEP",
		"Adverse Media",
		"Regulatory",
	}
	return categories[rand.Intn(len(categories))]
}

func (wc *WorldCheckMock) generateSubCategory() string {
	subCategories := []string{
		"Money Laundering",
		"Fraud",
		"Corruption",
		"Drug Trafficking",
		"Human Trafficking",
		"Arms Dealing",
	}
	return subCategories[rand.Intn(len(subCategories))]
}

func (wc *WorldCheckMock) generateProfileStatus() string {
	statuses := []string{"active", "inactive", "deceased"}
	weights := []float64{0.8, 0.15, 0.05} // 80% active, 15% inactive, 5% deceased

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return statuses[i]
		}
	}
	return "active"
}

func (wc *WorldCheckMock) generateMatchScore() float64 {
	return rand.Float64()*0.3 + 0.7 // 0.7-1.0
}

func (wc *WorldCheckMock) generateConfidence() float64 {
	return rand.Float64()*0.2 + 0.8 // 0.8-1.0
}

func (wc *WorldCheckMock) generateAdverseMedia(entityName string) []AdverseMedia {
	// Most entities have 0-2 adverse media items
	numItems := rand.Intn(3)
	if numItems == 0 {
		return []AdverseMedia{}
	}

	media := make([]AdverseMedia, numItems)
	for i := 0; i < numItems; i++ {
		media[i] = AdverseMedia{
			MediaID:       fmt.Sprintf("MEDIA_%d", rand.Intn(100000)),
			Title:         wc.generateMediaTitle(entityName),
			Source:        wc.generateMediaSource(),
			URL:           wc.generateMediaURL(),
			PublishedDate: time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
			Content:       wc.generateMediaContent(entityName),
			Sentiment:     wc.generateSentiment(),
			Relevance:     rand.Float64()*0.4 + 0.6, // 0.6-1.0
			RiskLevel:     wc.generateRiskLevel(),
		}
	}

	return media
}

func (wc *WorldCheckMock) generateMediaTitle(entityName string) string {
	titles := []string{
		fmt.Sprintf("%s under investigation", entityName),
		fmt.Sprintf("Regulatory action against %s", entityName),
		fmt.Sprintf("%s faces compliance issues", entityName),
		fmt.Sprintf("Legal proceedings involving %s", entityName),
	}
	return titles[rand.Intn(len(titles))]
}

func (wc *WorldCheckMock) generateMediaSource() string {
	sources := []string{
		"Reuters",
		"Bloomberg",
		"Financial Times",
		"Wall Street Journal",
		"BBC News",
		"CNN",
	}
	return sources[rand.Intn(len(sources))]
}

func (wc *WorldCheckMock) generateMediaURL() string {
	return fmt.Sprintf("https://example.com/news/%d", rand.Intn(100000))
}

func (wc *WorldCheckMock) generateMediaContent(entityName string) string {
	return fmt.Sprintf("Recent developments involving %s have raised concerns among regulators and industry observers.", entityName)
}

func (wc *WorldCheckMock) generateSentiment() string {
	sentiments := []string{"negative", "neutral", "positive"}
	weights := []float64{0.7, 0.2, 0.1} // 70% negative, 20% neutral, 10% positive

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return sentiments[i]
		}
	}
	return "negative"
}

func (wc *WorldCheckMock) generateIsPEP() bool {
	return rand.Float64() < 0.1 // 10% chance of being PEP
}

func (wc *WorldCheckMock) generatePEPLevel() string {
	levels := []string{"domestic", "foreign", "international"}
	return levels[rand.Intn(len(levels))]
}

func (wc *WorldCheckMock) generatePosition() string {
	positions := []string{
		"Minister",
		"Member of Parliament",
		"Judge",
		"Senior Government Official",
		"Military Officer",
		"Central Bank Official",
	}
	return positions[rand.Intn(len(positions))]
}

func (wc *WorldCheckMock) generateJurisdiction() string {
	jurisdictions := []string{"United States", "United Kingdom", "Canada", "Germany", "France"}
	return jurisdictions[rand.Intn(len(jurisdictions))]
}

func (wc *WorldCheckMock) generateStartDate() *time.Time {
	startDate := time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour)
	return &startDate
}

func (wc *WorldCheckMock) generateEndDate() *time.Time {
	if rand.Float64() > 0.5 { // 50% chance of having end date
		endDate := time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour)
		return &endDate
	}
	return nil
}

func (wc *WorldCheckMock) generateFamilyMembers() []string {
	if rand.Float64() > 0.7 { // 30% chance of having family members
		return []string{"Spouse", "Child", "Parent"}
	}
	return []string{}
}

func (wc *WorldCheckMock) generateCloseAssociates() []string {
	if rand.Float64() > 0.8 { // 20% chance of having close associates
		return []string{"Business Partner", "Close Friend"}
	}
	return []string{}
}

func (wc *WorldCheckMock) generateIsSanctioned() bool {
	return rand.Float64() < 0.02 // 2% chance of being sanctioned
}

func (wc *WorldCheckMock) generateSanctionsList() string {
	lists := []string{
		"OFAC SDN List",
		"EU Sanctions List",
		"UN Security Council List",
		"UK Sanctions List",
	}
	return lists[rand.Intn(len(lists))]
}

func (wc *WorldCheckMock) generateSanctionsProgram() string {
	programs := []string{
		"Terrorism Sanctions",
		"Narcotics Trafficking Sanctions",
		"Proliferation Sanctions",
		"Human Rights Sanctions",
	}
	return programs[rand.Intn(len(programs))]
}

func (wc *WorldCheckMock) generateEffectiveDate() *time.Time {
	effectiveDate := time.Now().Add(-time.Duration(rand.Intn(1825)) * 24 * time.Hour)
	return &effectiveDate
}

func (wc *WorldCheckMock) generateExpiryDate() *time.Time {
	if rand.Float64() > 0.3 { // 70% chance of having expiry date
		expiryDate := time.Now().Add(time.Duration(rand.Intn(1825)) * 24 * time.Hour)
		return &expiryDate
	}
	return nil
}

func (wc *WorldCheckMock) generateReason() string {
	reasons := []string{
		"Terrorism activities",
		"Narcotics trafficking",
		"Proliferation of weapons",
		"Human rights violations",
		"Corruption",
	}
	return reasons[rand.Intn(len(reasons))]
}

func (wc *WorldCheckMock) generateAuthority() string {
	authorities := []string{
		"US Treasury Department",
		"European Union",
		"United Nations",
		"UK Government",
	}
	return authorities[rand.Intn(len(authorities))]
}

func (wc *WorldCheckMock) generateOverallRiskScore() float64 {
	return rand.Float64()*0.8 + 0.1 // 0.1-0.9
}

func (wc *WorldCheckMock) generateFinancialRisk() float64 {
	return rand.Float64()*0.8 + 0.1 // 0.1-0.9
}

func (wc *WorldCheckMock) generateReputationalRisk() float64 {
	return rand.Float64()*0.8 + 0.1 // 0.1-0.9
}

func (wc *WorldCheckMock) generateRegulatoryRisk() float64 {
	return rand.Float64()*0.8 + 0.1 // 0.1-0.9
}

func (wc *WorldCheckMock) generateOperationalRisk() float64 {
	return rand.Float64()*0.8 + 0.1 // 0.1-0.9
}

func (wc *WorldCheckMock) generateRiskFactors() []string {
	factors := []string{
		"Adverse media coverage",
		"Regulatory investigations",
		"Financial irregularities",
		"Compliance violations",
		"Reputational issues",
	}

	numFactors := rand.Intn(3) + 1
	selectedFactors := make([]string, numFactors)
	for i := 0; i < numFactors; i++ {
		selectedFactors[i] = factors[rand.Intn(len(factors))]
	}

	return selectedFactors
}

func (wc *WorldCheckMock) generateRecommendations() []string {
	recommendations := []string{
		"Enhanced due diligence required",
		"Regular monitoring recommended",
		"Additional verification needed",
		"Compliance review suggested",
		"Risk assessment update required",
	}

	numRecommendations := rand.Intn(3) + 1
	selectedRecommendations := make([]string, numRecommendations)
	for i := 0; i < numRecommendations; i++ {
		selectedRecommendations[i] = recommendations[rand.Intn(len(recommendations))]
	}

	return selectedRecommendations
}

func (wc *WorldCheckMock) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}

func (wc *WorldCheckMock) sanitizeForID(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", "_"), "&", "and"))
}

func (wc *WorldCheckMock) sanitizeForURL(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", ""), "&", "and"))
}
