package ofac

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// OFACMock provides mock implementation of OFAC (Office of Foreign Assets Control) API
type OFACMock struct {
	logger *zap.Logger
	config *OFACConfig
	// Real-time update simulation
	lastUpdate     time.Time
	updateInterval time.Duration
}

// OFACConfig holds configuration for OFAC API
type OFACConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// SanctionsList represents a sanctions list entry
type SanctionsList struct {
	EntityID         string    `json:"entity_id"`
	EntityName       string    `json:"entity_name"`
	EntityType       string    `json:"entity_type"` // "individual", "entity", "vessel", "aircraft"
	Country          string    `json:"country"`
	SanctionsProgram string    `json:"sanctions_program"`
	ProgramList      string    `json:"program_list"`
	Title            string    `json:"title"`
	Remarks          string    `json:"remarks"`
	EffectiveDate    time.Time `json:"effective_date"`
	LastUpdated      time.Time `json:"last_updated"`
	MatchScore       float64   `json:"match_score"`
	RiskLevel        string    `json:"risk_level"` // "high", "medium", "low"
}

// SanctionsSearchResult represents the result of a sanctions search
type SanctionsSearchResult struct {
	Query        string          `json:"query"`
	TotalMatches int             `json:"total_matches"`
	Matches      []SanctionsList `json:"matches"`
	SearchTime   time.Duration   `json:"search_time"`
	LastChecked  time.Time       `json:"last_checked"`
	DataQuality  string          `json:"data_quality"`
}

// ComplianceStatus represents compliance status for an entity
type ComplianceStatus struct {
	EntityName       string    `json:"entity_name"`
	IsCompliant      bool      `json:"is_compliant"`
	RiskLevel        string    `json:"risk_level"`
	SanctionsMatches int       `json:"sanctions_matches"`
	ComplianceScore  float64   `json:"compliance_score"`
	LastScreened     time.Time `json:"last_screened"`
	NextScreening    time.Time `json:"next_screening"`
	ComplianceNotes  string    `json:"compliance_notes"`
}

// EntityVerification represents entity verification result
type EntityVerification struct {
	EntityName        string    `json:"entity_name"`
	IsVerified        bool      `json:"is_verified"`
	VerificationScore float64   `json:"verification_score"`
	MatchType         string    `json:"match_type"` // "exact", "partial", "fuzzy", "none"
	Confidence        float64   `json:"confidence"`
	VerificationDate  time.Time `json:"verification_date"`
	Notes             string    `json:"notes"`
}

// OFACResult represents the combined result from OFAC
type OFACResult struct {
	SanctionsSearch    *SanctionsSearchResult `json:"sanctions_search,omitempty"`
	ComplianceStatus   *ComplianceStatus      `json:"compliance_status,omitempty"`
	EntityVerification *EntityVerification    `json:"entity_verification,omitempty"`
	DataQuality        string                 `json:"data_quality"`
	LastChecked        time.Time              `json:"last_checked"`
	ProcessingTime     time.Duration          `json:"processing_time"`
}

// NewOFACMock creates a new OFAC mock client
func NewOFACMock(config *OFACConfig, logger *zap.Logger) *OFACMock {
	return &OFACMock{
		logger:         logger,
		config:         config,
		lastUpdate:     time.Now().Add(-24 * time.Hour), // Simulate last update 24 hours ago
		updateInterval: 24 * time.Hour,                  // Simulate daily updates
	}
}

// SearchSanctions searches for entities in sanctions lists
func (ofac *OFACMock) SearchSanctions(ctx context.Context, entityName, country string) (*SanctionsSearchResult, error) {
	startTime := time.Now()
	ofac.logger.Info("Searching OFAC sanctions lists (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// Generate mock sanctions search results
	matches := ofac.generateSanctionsMatches(entityName, country)

	result := &SanctionsSearchResult{
		Query:        entityName,
		TotalMatches: len(matches),
		Matches:      matches,
		SearchTime:   time.Since(startTime),
		LastChecked:  time.Now(),
		DataQuality:  ofac.generateDataQuality(),
	}

	ofac.logger.Info("OFAC sanctions search completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", result.TotalMatches),
		zap.Duration("search_time", result.SearchTime))

	return result, nil
}

// VerifyEntity verifies an entity against sanctions lists
func (ofac *OFACMock) VerifyEntity(ctx context.Context, entityName, country string) (*EntityVerification, error) {
	ofac.logger.Info("Verifying entity against OFAC lists (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)

	// Generate mock verification result
	verification := &EntityVerification{
		EntityName:        entityName,
		IsVerified:        ofac.generateVerificationStatus(entityName),
		VerificationScore: ofac.generateVerificationScore(),
		MatchType:         ofac.generateMatchType(),
		Confidence:        ofac.generateConfidence(),
		VerificationDate:  time.Now(),
		Notes:             ofac.generateVerificationNotes(entityName),
	}

	ofac.logger.Info("OFAC entity verification completed (mock)",
		zap.String("entity_name", entityName),
		zap.Bool("is_verified", verification.IsVerified),
		zap.Float64("verification_score", verification.VerificationScore))

	return verification, nil
}

// GetComplianceStatus gets compliance status for an entity
func (ofac *OFACMock) GetComplianceStatus(ctx context.Context, entityName, country string) (*ComplianceStatus, error) {
	ofac.logger.Info("Getting OFAC compliance status (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(250)+100) * time.Millisecond)

	// Generate mock compliance status
	complianceStatus := &ComplianceStatus{
		EntityName:       entityName,
		IsCompliant:      ofac.generateComplianceStatus(entityName),
		RiskLevel:        ofac.generateRiskLevel(),
		SanctionsMatches: ofac.generateSanctionsMatchesCount(),
		ComplianceScore:  ofac.generateComplianceScore(),
		LastScreened:     time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		NextScreening:    time.Now().Add(time.Duration(rand.Intn(30)+1) * 24 * time.Hour),
		ComplianceNotes:  ofac.generateComplianceNotes(entityName),
	}

	ofac.logger.Info("OFAC compliance status retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Bool("is_compliant", complianceStatus.IsCompliant),
		zap.String("risk_level", complianceStatus.RiskLevel))

	return complianceStatus, nil
}

// GetComprehensiveData retrieves all available data from OFAC
func (ofac *OFACMock) GetComprehensiveData(ctx context.Context, entityName, country string) (*OFACResult, error) {
	startTime := time.Now()
	ofac.logger.Info("Getting comprehensive OFAC data (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Get all data in parallel
	type result struct {
		sanctionsSearch    *SanctionsSearchResult
		complianceStatus   *ComplianceStatus
		entityVerification *EntityVerification
		err                error
	}

	results := make(chan result, 3)

	// Get sanctions search
	go func() {
		ss, err := ofac.SearchSanctions(ctx, entityName, country)
		results <- result{sanctionsSearch: ss, err: err}
	}()

	// Get compliance status
	go func() {
		cs, err := ofac.GetComplianceStatus(ctx, entityName, country)
		results <- result{complianceStatus: cs, err: err}
	}()

	// Get entity verification
	go func() {
		ev, err := ofac.VerifyEntity(ctx, entityName, country)
		results <- result{entityVerification: ev, err: err}
	}()

	// Collect results
	var comprehensiveResult result
	for i := 0; i < 3; i++ {
		r := <-results
		if r.err != nil {
			ofac.logger.Warn("Failed to get some OFAC data",
				zap.Error(r.err))
		}
		if r.sanctionsSearch != nil {
			comprehensiveResult.sanctionsSearch = r.sanctionsSearch
		}
		if r.complianceStatus != nil {
			comprehensiveResult.complianceStatus = r.complianceStatus
		}
		if r.entityVerification != nil {
			comprehensiveResult.entityVerification = r.entityVerification
		}
	}

	// Create comprehensive result
	ofacResult := &OFACResult{
		SanctionsSearch:    comprehensiveResult.sanctionsSearch,
		ComplianceStatus:   comprehensiveResult.complianceStatus,
		EntityVerification: comprehensiveResult.entityVerification,
		DataQuality:        ofac.generateDataQuality(),
		LastChecked:        time.Now(),
		ProcessingTime:     time.Since(startTime),
	}

	ofac.logger.Info("Comprehensive OFAC data retrieved (mock)",
		zap.String("entity_name", entityName),
		zap.Duration("processing_time", ofacResult.ProcessingTime),
		zap.String("data_quality", ofacResult.DataQuality))

	return ofacResult, nil
}

// GenerateRiskFactors generates risk factors from OFAC data
func (ofac *OFACMock) GenerateRiskFactors(result *OFACResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Sanctions risk factors
	if result.SanctionsSearch != nil {
		sanctionsRisk := 0.1 // Base risk
		if result.SanctionsSearch.TotalMatches > 0 {
			sanctionsRisk = 0.9 // High risk if matches found
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "sanctions",
			Name:        "sanctions_risk",
			Score:       sanctionsRisk,
			Weight:      0.4,
			Description: "Risk associated with sanctions list matches",
			Source:      "ofac",
			Confidence:  0.95,
			Impact:      "Sanctions violations can result in severe penalties",
			Mitigation:  "Regular sanctions screening and compliance monitoring",
			LastUpdated: &now,
		})
	}

	// Compliance risk factors
	if result.ComplianceStatus != nil {
		complianceRisk := 0.2 // Base risk
		if !result.ComplianceStatus.IsCompliant {
			complianceRisk = 0.8
		} else if result.ComplianceStatus.RiskLevel == "high" {
			complianceRisk = 0.6
		} else if result.ComplianceStatus.RiskLevel == "medium" {
			complianceRisk = 0.4
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "regulatory",
			Name:        "compliance_risk",
			Score:       complianceRisk,
			Weight:      0.3,
			Description: "Risk associated with regulatory compliance",
			Source:      "ofac",
			Confidence:  0.90,
			Impact:      "Compliance violations can result in regulatory action",
			Mitigation:  "Implement robust compliance monitoring and reporting",
			LastUpdated: &now,
		})
	}

	// Entity verification risk factors
	if result.EntityVerification != nil {
		verificationRisk := 0.3 // Base risk
		if !result.EntityVerification.IsVerified {
			verificationRisk = 0.7
		} else if result.EntityVerification.Confidence < 0.8 {
			verificationRisk = 0.5
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryOperational,
			Subcategory: "verification",
			Name:        "entity_verification_risk",
			Score:       verificationRisk,
			Weight:      0.3,
			Description: "Risk associated with entity verification",
			Source:      "ofac",
			Confidence:  0.85,
			Impact:      "Verification issues can impact business relationships",
			Mitigation:  "Implement additional verification procedures",
			LastUpdated: &now,
		})
	}

	return riskFactors
}

// Helper methods for generating mock data

func (ofac *OFACMock) generateSanctionsMatches(entityName, country string) []SanctionsList {
	// Most entities will have no matches (95% of the time)
	if rand.Float64() > 0.05 {
		return []SanctionsList{}
	}

	// Generate 1-3 matches for high-risk entities
	numMatches := rand.Intn(3) + 1
	matches := make([]SanctionsList, numMatches)

	for i := 0; i < numMatches; i++ {
		matches[i] = SanctionsList{
			EntityID:         fmt.Sprintf("OFAC_%d", rand.Intn(100000)),
			EntityName:       ofac.generateSimilarName(entityName),
			EntityType:       ofac.generateEntityType(),
			Country:          country,
			SanctionsProgram: ofac.generateSanctionsProgram(),
			ProgramList:      ofac.generateProgramList(),
			Title:            ofac.generateTitle(),
			Remarks:          ofac.generateRemarks(),
			EffectiveDate:    time.Now().Add(-time.Duration(rand.Intn(3650)) * 24 * time.Hour),
			LastUpdated:      time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
			MatchScore:       rand.Float64()*0.3 + 0.7, // 0.7-1.0
			RiskLevel:        "high",
		}
	}

	return matches
}

func (ofac *OFACMock) generateSimilarName(originalName string) string {
	// Generate a similar but not identical name
	variations := []string{
		originalName + " Ltd",
		originalName + " Inc",
		originalName + " Corp",
		"The " + originalName,
		originalName + " Group",
	}
	return variations[rand.Intn(len(variations))]
}

func (ofac *OFACMock) generateEntityType() string {
	types := []string{"individual", "entity", "vessel", "aircraft"}
	return types[rand.Intn(len(types))]
}

func (ofac *OFACMock) generateSanctionsProgram() string {
	programs := []string{
		"SDN (Specially Designated Nationals)",
		"OFAC Sanctions",
		"Terrorism Sanctions",
		"Narcotics Trafficking Sanctions",
		"Proliferation Sanctions",
	}
	return programs[rand.Intn(len(programs))]
}

func (ofac *OFACMock) generateProgramList() string {
	lists := []string{
		"SDN List",
		"OFAC List",
		"Terrorism List",
		"Narcotics List",
		"Proliferation List",
	}
	return lists[rand.Intn(len(lists))]
}

func (ofac *OFACMock) generateTitle() string {
	titles := []string{
		"Chief Executive Officer",
		"President",
		"Director",
		"Manager",
		"Owner",
		"",
	}
	return titles[rand.Intn(len(titles))]
}

func (ofac *OFACMock) generateRemarks() string {
	remarks := []string{
		"Subject to OFAC sanctions",
		"Designated for sanctions violations",
		"Associated with sanctioned activities",
		"Blocked pursuant to Executive Order",
		"",
	}
	return remarks[rand.Intn(len(remarks))]
}

func (ofac *OFACMock) generateVerificationStatus(entityName string) bool {
	// 95% of entities are verified
	return rand.Float64() > 0.05
}

func (ofac *OFACMock) generateVerificationScore() float64 {
	return rand.Float64()*0.3 + 0.7 // 0.7-1.0
}

func (ofac *OFACMock) generateMatchType() string {
	types := []string{"exact", "partial", "fuzzy", "none"}
	return types[rand.Intn(len(types))]
}

func (ofac *OFACMock) generateConfidence() float64 {
	return rand.Float64()*0.2 + 0.8 // 0.8-1.0
}

func (ofac *OFACMock) generateVerificationNotes(entityName string) string {
	notes := []string{
		"Entity verified against OFAC lists",
		"No matches found in sanctions databases",
		"Entity cleared for business operations",
		"Additional verification recommended",
		"",
	}
	return notes[rand.Intn(len(notes))]
}

func (ofac *OFACMock) generateComplianceStatus(entityName string) bool {
	// 90% of entities are compliant
	return rand.Float64() > 0.1
}

func (ofac *OFACMock) generateRiskLevel() string {
	levels := []string{"low", "medium", "high"}
	weights := []float64{0.7, 0.2, 0.1} // 70% low, 20% medium, 10% high

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

func (ofac *OFACMock) generateSanctionsMatchesCount() int {
	// Most entities have 0 matches
	if rand.Float64() > 0.05 {
		return 0
	}
	return rand.Intn(3) + 1
}

func (ofac *OFACMock) generateComplianceScore() float64 {
	return rand.Float64()*0.3 + 0.7 // 0.7-1.0
}

func (ofac *OFACMock) generateComplianceNotes(entityName string) string {
	notes := []string{
		"Entity is compliant with OFAC regulations",
		"Regular screening recommended",
		"Monitor for any changes in sanctions status",
		"Entity requires additional due diligence",
		"",
	}
	return notes[rand.Intn(len(notes))]
}

func (ofac *OFACMock) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}

// Real-time update simulation methods

// CheckForUpdates simulates checking for sanctions list updates
func (ofac *OFACMock) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	ofac.logger.Info("Checking for OFAC sanctions list updates (mock)")

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)

	// Check if enough time has passed for an update
	timeSinceLastUpdate := time.Since(ofac.lastUpdate)
	hasUpdate := timeSinceLastUpdate >= ofac.updateInterval

	updateInfo := &UpdateInfo{
		HasUpdate:       hasUpdate,
		LastUpdate:      ofac.lastUpdate,
		NextUpdate:      ofac.lastUpdate.Add(ofac.updateInterval),
		UpdateAvailable: hasUpdate,
		UpdateSize:      rand.Intn(100) + 10, // 10-110 new entries
		UpdateType:      "incremental",
		CheckTime:       time.Now(),
	}

	if hasUpdate {
		// Simulate update
		ofac.lastUpdate = time.Now()
		updateInfo.LastUpdate = ofac.lastUpdate
		updateInfo.NextUpdate = ofac.lastUpdate.Add(ofac.updateInterval)

		ofac.logger.Info("OFAC sanctions list update available (mock)",
			zap.Int("update_size", updateInfo.UpdateSize),
			zap.Time("last_update", updateInfo.LastUpdate))
	} else {
		ofac.logger.Info("No OFAC sanctions list updates available (mock)",
			zap.Time("next_check", updateInfo.NextUpdate))
	}

	return updateInfo, nil
}

// GetUpdateHistory simulates retrieving update history
func (ofac *OFACMock) GetUpdateHistory(ctx context.Context, days int) ([]UpdateInfo, error) {
	ofac.logger.Info("Retrieving OFAC update history (mock)",
		zap.Int("days", days))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+150) * time.Millisecond)

	var history []UpdateInfo
	now := time.Now()

	// Generate mock update history
	for i := 0; i < days; i++ {
		updateDate := now.Add(-time.Duration(i) * 24 * time.Hour)

		// Simulate updates every 1-3 days
		if rand.Float64() > 0.3 { // 70% chance of update
			updateInfo := UpdateInfo{
				HasUpdate:       true,
				LastUpdate:      updateDate,
				NextUpdate:      updateDate.Add(ofac.updateInterval),
				UpdateAvailable: true,
				UpdateSize:      rand.Intn(50) + 5, // 5-55 new entries
				UpdateType:      "incremental",
				CheckTime:       updateDate,
			}
			history = append(history, updateInfo)
		}
	}

	ofac.logger.Info("OFAC update history retrieved (mock)",
		zap.Int("history_entries", len(history)))

	return history, nil
}

// GetSanctionsListVersion simulates getting the current sanctions list version
func (ofac *OFACMock) GetSanctionsListVersion(ctx context.Context) (*SanctionsListVersion, error) {
	ofac.logger.Info("Getting OFAC sanctions list version (mock)")

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)

	version := &SanctionsListVersion{
		Version:         fmt.Sprintf("v%d.%d.%d", rand.Intn(10)+1, rand.Intn(10), rand.Intn(100)),
		LastUpdated:     ofac.lastUpdate,
		TotalEntries:    rand.Intn(10000) + 5000, // 5000-15000 entries
		NewEntries:      rand.Intn(100) + 10,     // 10-110 new entries
		ModifiedEntries: rand.Intn(50) + 5,       // 5-55 modified entries
		RemovedEntries:  rand.Intn(20) + 1,       // 1-21 removed entries
		DataQuality:     ofac.generateDataQuality(),
		Source:          "OFAC",
	}

	ofac.logger.Info("OFAC sanctions list version retrieved (mock)",
		zap.String("version", version.Version),
		zap.Int("total_entries", version.TotalEntries),
		zap.Time("last_updated", version.LastUpdated))

	return version, nil
}

// IsHealthy checks if the OFAC service is healthy
func (ofac *OFACMock) IsHealthy(ctx context.Context) error {
	ofac.logger.Info("Checking OFAC service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}

// Additional types for real-time updates

// UpdateInfo represents information about sanctions list updates
type UpdateInfo struct {
	HasUpdate       bool      `json:"has_update"`
	LastUpdate      time.Time `json:"last_update"`
	NextUpdate      time.Time `json:"next_update"`
	UpdateAvailable bool      `json:"update_available"`
	UpdateSize      int       `json:"update_size"`
	UpdateType      string    `json:"update_type"` // "incremental", "full"
	CheckTime       time.Time `json:"check_time"`
}

// SanctionsListVersion represents the version information of a sanctions list
type SanctionsListVersion struct {
	Version         string    `json:"version"`
	LastUpdated     time.Time `json:"last_updated"`
	TotalEntries    int       `json:"total_entries"`
	NewEntries      int       `json:"new_entries"`
	ModifiedEntries int       `json:"modified_entries"`
	RemovedEntries  int       `json:"removed_entries"`
	DataQuality     string    `json:"data_quality"`
	Source          string    `json:"source"`
}
