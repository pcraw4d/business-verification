package thomson_reuters

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// WorldCheckMock provides World-Check specific screening capabilities
type WorldCheckMock struct {
	logger *zap.Logger
	config *WorldCheckConfig
}

// WorldCheckConfig holds configuration for World-Check API
type WorldCheckConfig struct {
	APIKey          string        `json:"api_key"`
	BaseURL         string        `json:"base_url"`
	Timeout         time.Duration `json:"timeout"`
	RateLimit       int           `json:"rate_limit_per_minute"`
	Enabled         bool          `json:"enabled"`
	EnablePEP       bool          `json:"enable_pep"`
	EnableSanctions bool          `json:"enable_sanctions"`
	EnableAdverse   bool          `json:"enable_adverse"`
}

// WorldCheckEntity represents a World-Check entity
type WorldCheckEntity struct {
	EntityID           string                 `json:"entity_id"`
	EntityName         string                 `json:"entity_name"`
	EntityType         string                 `json:"entity_type"` // "individual", "entity", "vessel", "aircraft"
	Country            string                 `json:"country"`
	DateOfBirth        *time.Time             `json:"date_of_birth,omitempty"`
	PlaceOfBirth       string                 `json:"place_of_birth,omitempty"`
	Nationality        string                 `json:"nationality,omitempty"`
	Address            string                 `json:"address,omitempty"`
	Phone              string                 `json:"phone,omitempty"`
	Email              string                 `json:"email,omitempty"`
	Website            string                 `json:"website,omitempty"`
	BusinessType       string                 `json:"business_type,omitempty"`
	RegistrationNumber string                 `json:"registration_number,omitempty"`
	TaxID              string                 `json:"tax_id,omitempty"`
	RiskLevel          string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	RiskScore          float64                `json:"risk_score"`
	Category           string                 `json:"category"` // "sanctions", "pep", "adverse_media", "regulatory"
	Subcategory        string                 `json:"subcategory"`
	Source             string                 `json:"source"`
	LastUpdated        time.Time              `json:"last_updated"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// WorldCheckMatch represents a match found in World-Check
type WorldCheckMatch struct {
	MatchID        string     `json:"match_id"`
	EntityID       string     `json:"entity_id"`
	EntityName     string     `json:"entity_name"`
	MatchType      string     `json:"match_type"` // "exact", "partial", "fuzzy", "phonetic"
	MatchScore     float64    `json:"match_score"`
	Confidence     float64    `json:"confidence"`
	RiskLevel      string     `json:"risk_level"`
	Category       string     `json:"category"`
	Subcategory    string     `json:"subcategory"`
	Source         string     `json:"source"`
	MatchDetails   string     `json:"match_details"`
	AdditionalInfo string     `json:"additional_info"`
	LastUpdated    time.Time  `json:"last_updated"`
	IsActive       bool       `json:"is_active"`
	ExpiryDate     *time.Time `json:"expiry_date,omitempty"`
}

// WorldCheckScreeningResult represents the result of a World-Check screening
type WorldCheckScreeningResult struct {
	RequestID         string            `json:"request_id"`
	QueryName         string            `json:"query_name"`
	QueryCountry      string            `json:"query_country"`
	TotalMatches      int               `json:"total_matches"`
	HighRiskMatches   int               `json:"high_risk_matches"`
	MediumRiskMatches int               `json:"medium_risk_matches"`
	LowRiskMatches    int               `json:"low_risk_matches"`
	Matches           []WorldCheckMatch `json:"matches"`
	OverallRiskScore  float64           `json:"overall_risk_score"`
	OverallRiskLevel  string            `json:"overall_risk_level"`
	ScreeningTime     time.Duration     `json:"screening_time"`
	LastChecked       time.Time         `json:"last_checked"`
	DataQuality       string            `json:"data_quality"`
	Sources           []string          `json:"sources"`
}

// WorldCheckBatchResult represents the result of a batch screening
type WorldCheckBatchResult struct {
	BatchID           string                      `json:"batch_id"`
	TotalRequests     int                         `json:"total_requests"`
	ProcessedRequests int                         `json:"processed_requests"`
	FailedRequests    int                         `json:"failed_requests"`
	Results           []WorldCheckScreeningResult `json:"results"`
	ProcessingTime    time.Duration               `json:"processing_time"`
	CompletedAt       time.Time                   `json:"completed_at"`
}

// WorldCheckWatchlist represents a watchlist entry
type WorldCheckWatchlist struct {
	WatchlistID       string     `json:"watchlist_id"`
	EntityID          string     `json:"entity_id"`
	EntityName        string     `json:"entity_name"`
	WatchlistType     string     `json:"watchlist_type"` // "custom", "regulatory", "internal"
	RiskLevel         string     `json:"risk_level"`
	Reason            string     `json:"reason"`
	CreatedBy         string     `json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
	LastUpdated       time.Time  `json:"last_updated"`
	IsActive          bool       `json:"is_active"`
	ExpiryDate        *time.Time `json:"expiry_date,omitempty"`
	NotificationEmail string     `json:"notification_email,omitempty"`
}

// NewWorldCheckClient creates a new World-Check client
func NewWorldCheckClient(config *WorldCheckConfig, logger *zap.Logger) *WorldCheckMock {
	return &WorldCheckMock{
		logger: logger,
		config: config,
	}
}

// ScreenEntity performs comprehensive screening of an entity
func (wc *WorldCheckMock) ScreenEntity(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	startTime := time.Now()
	wc.logger.Info("Screening entity in World-Check (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(800)+200) * time.Millisecond)

	// Generate mock screening results
	matches := wc.generateWorldCheckMatches(entityName, country)

	// Calculate risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range matches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.9
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.6
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.3
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(matches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(matches))
		if overallRiskScore >= 0.8 {
			overallRiskLevel = "critical"
		} else if overallRiskScore >= 0.6 {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.4 {
			overallRiskLevel = "medium"
		}
	}

	result := &WorldCheckScreeningResult{
		RequestID:         wc.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(matches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           matches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Since(startTime),
		LastChecked:       time.Now(),
		DataQuality:       wc.generateDataQuality(),
		Sources:           wc.getAvailableSources(),
	}

	wc.logger.Info("World-Check screening completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", result.TotalMatches),
		zap.String("overall_risk_level", result.OverallRiskLevel),
		zap.Duration("screening_time", result.ScreeningTime))

	return result, nil
}

// ScreenBatch performs batch screening of multiple entities
func (wc *WorldCheckMock) ScreenBatch(ctx context.Context, entities []string, country string) (*WorldCheckBatchResult, error) {
	startTime := time.Now()
	wc.logger.Info("Performing batch screening in World-Check (mock)",
		zap.Int("entity_count", len(entities)),
		zap.String("country", country))

	// Simulate batch processing delay
	time.Sleep(time.Duration(rand.Intn(2000)+500) * time.Millisecond)

	var results []WorldCheckScreeningResult
	processedCount := 0
	failedCount := 0

	for _, entity := range entities {
		result, err := wc.ScreenEntity(ctx, entity, country)
		if err != nil {
			wc.logger.Warn("Failed to screen entity in batch",
				zap.String("entity", entity),
				zap.Error(err))
			failedCount++
			continue
		}
		results = append(results, *result)
		processedCount++
	}

	batchResult := &WorldCheckBatchResult{
		BatchID:           wc.generateBatchID(),
		TotalRequests:     len(entities),
		ProcessedRequests: processedCount,
		FailedRequests:    failedCount,
		Results:           results,
		ProcessingTime:    time.Since(startTime),
		CompletedAt:       time.Now(),
	}

	wc.logger.Info("World-Check batch screening completed (mock)",
		zap.String("batch_id", batchResult.BatchID),
		zap.Int("processed", batchResult.ProcessedRequests),
		zap.Int("failed", batchResult.FailedRequests),
		zap.Duration("processing_time", batchResult.ProcessingTime))

	return batchResult, nil
}

// ScreenPEP performs Politically Exposed Person screening
func (wc *WorldCheckMock) ScreenPEP(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	wc.logger.Info("Screening for PEP status in World-Check (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate PEP-specific screening
	time.Sleep(time.Duration(rand.Intn(600)+200) * time.Millisecond)

	// Generate PEP-specific matches
	matches := wc.generatePEPMatches(entityName, country)

	// Calculate PEP-specific risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range matches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.9
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.6
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.3
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(matches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(matches))
		if overallRiskScore >= 0.7 {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.4 {
			overallRiskLevel = "medium"
		}
	}

	result := &WorldCheckScreeningResult{
		RequestID:         wc.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(matches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           matches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Duration(rand.Intn(500)+200) * time.Millisecond,
		LastChecked:       time.Now(),
		DataQuality:       wc.generateDataQuality(),
		Sources:           []string{"PEP Database", "Government Records", "Public Records"},
	}

	wc.logger.Info("World-Check PEP screening completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("pep_matches", result.TotalMatches),
		zap.String("risk_level", result.OverallRiskLevel))

	return result, nil
}

// ScreenSanctions performs sanctions list screening
func (wc *WorldCheckMock) ScreenSanctions(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	wc.logger.Info("Screening for sanctions in World-Check (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate sanctions-specific screening
	time.Sleep(time.Duration(rand.Intn(700)+300) * time.Millisecond)

	// Generate sanctions-specific matches
	matches := wc.generateSanctionsMatches(entityName, country)

	// Calculate sanctions-specific risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range matches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.95 // Sanctions are high risk
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.7
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.4
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(matches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(matches))
		if overallRiskScore >= 0.8 {
			overallRiskLevel = "critical"
		} else if overallRiskScore >= 0.6 {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.3 {
			overallRiskLevel = "medium"
		}
	}

	result := &WorldCheckScreeningResult{
		RequestID:         wc.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(matches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           matches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Duration(rand.Intn(600)+300) * time.Millisecond,
		LastChecked:       time.Now(),
		DataQuality:       wc.generateDataQuality(),
		Sources:           []string{"OFAC", "UN Sanctions", "EU Sanctions", "UK Sanctions"},
	}

	wc.logger.Info("World-Check sanctions screening completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("sanctions_matches", result.TotalMatches),
		zap.String("risk_level", result.OverallRiskLevel))

	return result, nil
}

// ScreenAdverseMedia performs adverse media screening
func (wc *WorldCheckMock) ScreenAdverseMedia(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	wc.logger.Info("Screening for adverse media in World-Check (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Simulate adverse media screening
	time.Sleep(time.Duration(rand.Intn(1000)+400) * time.Millisecond)

	// Generate adverse media matches
	matches := wc.generateAdverseMediaMatches(entityName, country)

	// Calculate adverse media risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range matches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.8
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.5
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.2
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(matches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(matches))
		if overallRiskScore >= 0.7 {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.4 {
			overallRiskLevel = "medium"
		}
	}

	result := &WorldCheckScreeningResult{
		RequestID:         wc.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(matches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           matches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Duration(rand.Intn(800)+400) * time.Millisecond,
		LastChecked:       time.Now(),
		DataQuality:       wc.generateDataQuality(),
		Sources:           []string{"News Database", "Media Monitoring", "Public Records", "Regulatory Filings"},
	}

	wc.logger.Info("World-Check adverse media screening completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("adverse_media_matches", result.TotalMatches),
		zap.String("risk_level", result.OverallRiskLevel))

	return result, nil
}

// GetComprehensiveScreening performs all types of screening
func (wc *WorldCheckMock) GetComprehensiveScreening(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	startTime := time.Now()
	wc.logger.Info("Performing comprehensive World-Check screening (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Perform all screening types in parallel
	type screeningResult struct {
		pepResult       *WorldCheckScreeningResult
		sanctionsResult *WorldCheckScreeningResult
		adverseResult   *WorldCheckScreeningResult
		err             error
	}

	results := make(chan screeningResult, 1)

	go func() {
		var result screeningResult

		// PEP screening
		if wc.config.EnablePEP {
			pep, err := wc.ScreenPEP(ctx, entityName, country)
			result.pepResult = pep
			if err != nil {
				result.err = fmt.Errorf("PEP screening failed: %w", err)
			}
		}

		// Sanctions screening
		if wc.config.EnableSanctions {
			sanctions, err := wc.ScreenSanctions(ctx, entityName, country)
			result.sanctionsResult = sanctions
			if err != nil {
				result.err = fmt.Errorf("sanctions screening failed: %w", err)
			}
		}

		// Adverse media screening
		if wc.config.EnableAdverse {
			adverse, err := wc.ScreenAdverseMedia(ctx, entityName, country)
			result.adverseResult = adverse
			if err != nil {
				result.err = fmt.Errorf("adverse media screening failed: %w", err)
			}
		}

		results <- result
	}()

	// Wait for results
	result := <-results
	if result.err != nil {
		return nil, result.err
	}

	// Combine all matches
	var allMatches []WorldCheckMatch
	var allSources []string

	if result.pepResult != nil {
		allMatches = append(allMatches, result.pepResult.Matches...)
		allSources = append(allSources, result.pepResult.Sources...)
	}

	if result.sanctionsResult != nil {
		allMatches = append(allMatches, result.sanctionsResult.Matches...)
		allSources = append(allSources, result.sanctionsResult.Sources...)
	}

	if result.adverseResult != nil {
		allMatches = append(allMatches, result.adverseResult.Matches...)
		allSources = append(allSources, result.adverseResult.Sources...)
	}

	// Calculate combined risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range allMatches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.9
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.6
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.3
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(allMatches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(allMatches))
		if overallRiskScore >= 0.8 {
			overallRiskLevel = "critical"
		} else if overallRiskScore >= 0.6 {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.4 {
			overallRiskLevel = "medium"
		}
	}

	// Remove duplicate sources
	uniqueSources := make(map[string]bool)
	for _, source := range allSources {
		uniqueSources[source] = true
	}
	var finalSources []string
	for source := range uniqueSources {
		finalSources = append(finalSources, source)
	}

	comprehensiveResult := &WorldCheckScreeningResult{
		RequestID:         wc.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(allMatches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           allMatches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Since(startTime),
		LastChecked:       time.Now(),
		DataQuality:       wc.generateDataQuality(),
		Sources:           finalSources,
	}

	wc.logger.Info("Comprehensive World-Check screening completed (mock)",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", comprehensiveResult.TotalMatches),
		zap.String("overall_risk_level", comprehensiveResult.OverallRiskLevel),
		zap.Duration("screening_time", comprehensiveResult.ScreeningTime))

	return comprehensiveResult, nil
}

// GenerateRiskFactors generates risk factors from World-Check data
func (wc *WorldCheckMock) GenerateRiskFactors(result *WorldCheckScreeningResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Overall World-Check risk factor
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryCompliance,
		Subcategory: "worldcheck",
		Name:        "worldcheck_risk",
		Score:       result.OverallRiskScore,
		Weight:      0.4,
		Description: "Overall risk score from World-Check comprehensive screening",
		Source:      "worldcheck",
		Confidence:  0.95,
		Impact:      "World-Check matches indicate potential compliance and reputational risks",
		Mitigation:  "Conduct enhanced due diligence and monitor for updates",
		LastUpdated: &now,
	})

	// PEP risk factor
	pepMatches := 0
	for _, match := range result.Matches {
		if match.Category == "pep" {
			pepMatches++
		}
	}

	if pepMatches > 0 {
		pepRisk := 0.8 // High risk for PEP matches
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "pep",
			Name:        "pep_risk",
			Score:       pepRisk,
			Weight:      0.3,
			Description: "Politically Exposed Person risk from World-Check screening",
			Source:      "worldcheck",
			Confidence:  0.90,
			Impact:      "PEP status requires enhanced due diligence and monitoring",
			Mitigation:  "Implement PEP-specific due diligence procedures",
			LastUpdated: &now,
		})
	}

	// Sanctions risk factor
	sanctionsMatches := 0
	for _, match := range result.Matches {
		if match.Category == "sanctions" {
			sanctionsMatches++
		}
	}

	if sanctionsMatches > 0 {
		sanctionsRisk := 0.95 // Critical risk for sanctions matches
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "sanctions",
			Name:        "sanctions_risk",
			Score:       sanctionsRisk,
			Weight:      0.5,
			Description: "Sanctions list matches from World-Check screening",
			Source:      "worldcheck",
			Confidence:  0.98,
			Impact:      "Sanctions matches require immediate action and compliance review",
			Mitigation:  "Immediate compliance review and potential business relationship termination",
			LastUpdated: &now,
		})
	}

	// Adverse media risk factor
	adverseMatches := 0
	for _, match := range result.Matches {
		if match.Category == "adverse_media" {
			adverseMatches++
		}
	}

	if adverseMatches > 0 {
		adverseRisk := 0.6 // Medium-high risk for adverse media
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryReputational,
			Subcategory: "adverse_media",
			Name:        "adverse_media_risk",
			Score:       adverseRisk,
			Weight:      0.25,
			Description: "Adverse media matches from World-Check screening",
			Source:      "worldcheck",
			Confidence:  0.85,
			Impact:      "Adverse media can impact business reputation and relationships",
			Mitigation:  "Monitor media coverage and assess reputational impact",
			LastUpdated: &now,
		})
	}

	return riskFactors
}

// Helper methods for generating mock data

func (wc *WorldCheckMock) generateRequestID() string {
	return fmt.Sprintf("WC_%d", time.Now().UnixNano())
}

func (wc *WorldCheckMock) generateBatchID() string {
	return fmt.Sprintf("WC_BATCH_%d", time.Now().UnixNano())
}

func (wc *WorldCheckMock) generateWorldCheckMatches(entityName, country string) []WorldCheckMatch {
	var matches []WorldCheckMatch

	// 90% of entities have no matches
	if rand.Float64() > 0.1 {
		return matches
	}

	// Generate 1-5 matches for entities with matches
	numMatches := rand.Intn(5) + 1
	for i := 0; i < numMatches; i++ {
		match := WorldCheckMatch{
			MatchID:        fmt.Sprintf("WC_MATCH_%d_%d", time.Now().UnixNano(), i),
			EntityID:       fmt.Sprintf("WC_ENTITY_%d", rand.Intn(100000)),
			EntityName:     wc.generateSimilarName(entityName),
			MatchType:      wc.generateMatchType(),
			MatchScore:     rand.Float64()*0.4 + 0.6, // 0.6-1.0
			Confidence:     rand.Float64()*0.3 + 0.7, // 0.7-1.0
			RiskLevel:      wc.generateRiskLevel(),
			Category:       wc.generateCategory(),
			Subcategory:    wc.generateSubcategory(),
			Source:         wc.generateSource(),
			MatchDetails:   wc.generateMatchDetails(),
			AdditionalInfo: wc.generateAdditionalInfo(),
			LastUpdated:    time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			IsActive:       true,
		}

		// Add expiry date for some matches
		if rand.Float64() > 0.7 {
			expiry := time.Now().Add(time.Duration(rand.Intn(365)+30) * 24 * time.Hour)
			match.ExpiryDate = &expiry
		}

		matches = append(matches, match)
	}

	return matches
}

func (wc *WorldCheckMock) generatePEPMatches(entityName, country string) []WorldCheckMatch {
	var matches []WorldCheckMatch

	// 95% of entities are not PEPs
	if rand.Float64() > 0.05 {
		return matches
	}

	// Generate 1-3 PEP matches
	numMatches := rand.Intn(3) + 1
	for i := 0; i < numMatches; i++ {
		match := WorldCheckMatch{
			MatchID:        fmt.Sprintf("WC_PEP_%d_%d", time.Now().UnixNano(), i),
			EntityID:       fmt.Sprintf("WC_PEP_ENTITY_%d", rand.Intn(10000)),
			EntityName:     wc.generateSimilarName(entityName),
			MatchType:      "partial",
			MatchScore:     rand.Float64()*0.3 + 0.7, // 0.7-1.0
			Confidence:     rand.Float64()*0.2 + 0.8, // 0.8-1.0
			RiskLevel:      "high",
			Category:       "pep",
			Subcategory:    wc.generatePEPSubcategory(),
			Source:         "PEP Database",
			MatchDetails:   wc.generatePEPMatchDetails(),
			AdditionalInfo: wc.generatePEPAdditionalInfo(),
			LastUpdated:    time.Now().Add(-time.Duration(rand.Intn(60)) * 24 * time.Hour),
			IsActive:       true,
		}

		matches = append(matches, match)
	}

	return matches
}

func (wc *WorldCheckMock) generateSanctionsMatches(entityName, country string) []WorldCheckMatch {
	var matches []WorldCheckMatch

	// 98% of entities are not on sanctions lists
	if rand.Float64() > 0.02 {
		return matches
	}

	// Generate 1-2 sanctions matches
	numMatches := rand.Intn(2) + 1
	for i := 0; i < numMatches; i++ {
		match := WorldCheckMatch{
			MatchID:        fmt.Sprintf("WC_SANCTIONS_%d_%d", time.Now().UnixNano(), i),
			EntityID:       fmt.Sprintf("WC_SANCTIONS_ENTITY_%d", rand.Intn(5000)),
			EntityName:     wc.generateSimilarName(entityName),
			MatchType:      "exact",
			MatchScore:     rand.Float64()*0.2 + 0.8, // 0.8-1.0
			Confidence:     rand.Float64()*0.1 + 0.9, // 0.9-1.0
			RiskLevel:      "critical",
			Category:       "sanctions",
			Subcategory:    wc.generateSanctionsSubcategory(),
			Source:         wc.generateSanctionsSource(),
			MatchDetails:   wc.generateSanctionsMatchDetails(),
			AdditionalInfo: wc.generateSanctionsAdditionalInfo(),
			LastUpdated:    time.Now().Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
			IsActive:       true,
		}

		matches = append(matches, match)
	}

	return matches
}

func (wc *WorldCheckMock) generateAdverseMediaMatches(entityName, country string) []WorldCheckMatch {
	var matches []WorldCheckMatch

	// 85% of entities have no adverse media
	if rand.Float64() > 0.15 {
		return matches
	}

	// Generate 1-4 adverse media matches
	numMatches := rand.Intn(4) + 1
	for i := 0; i < numMatches; i++ {
		match := WorldCheckMatch{
			MatchID:        fmt.Sprintf("WC_ADVERSE_%d_%d", time.Now().UnixNano(), i),
			EntityID:       fmt.Sprintf("WC_ADVERSE_ENTITY_%d", rand.Intn(20000)),
			EntityName:     wc.generateSimilarName(entityName),
			MatchType:      wc.generateMatchType(),
			MatchScore:     rand.Float64()*0.5 + 0.5, // 0.5-1.0
			Confidence:     rand.Float64()*0.4 + 0.6, // 0.6-1.0
			RiskLevel:      wc.generateAdverseMediaRiskLevel(),
			Category:       "adverse_media",
			Subcategory:    wc.generateAdverseMediaSubcategory(),
			Source:         wc.generateAdverseMediaSource(),
			MatchDetails:   wc.generateAdverseMediaMatchDetails(),
			AdditionalInfo: wc.generateAdverseMediaAdditionalInfo(),
			LastUpdated:    time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour),
			IsActive:       true,
		}

		matches = append(matches, match)
	}

	return matches
}

func (wc *WorldCheckMock) generateSimilarName(originalName string) string {
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

func (wc *WorldCheckMock) generateMatchType() string {
	types := []string{"exact", "partial", "fuzzy", "phonetic"}
	weights := []float64{0.3, 0.4, 0.2, 0.1} // 30% exact, 40% partial, 20% fuzzy, 10% phonetic

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return types[i]
		}
	}
	return "partial"
}

func (wc *WorldCheckMock) generateRiskLevel() string {
	levels := []string{"low", "medium", "high", "critical"}
	weights := []float64{0.4, 0.3, 0.2, 0.1} // 40% low, 30% medium, 20% high, 10% critical

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return levels[i]
		}
	}
	return "medium"
}

func (wc *WorldCheckMock) generateCategory() string {
	categories := []string{"sanctions", "pep", "adverse_media", "regulatory"}
	return categories[rand.Intn(len(categories))]
}

func (wc *WorldCheckMock) generateSubcategory() string {
	subcategories := []string{
		"government_official", "family_member", "close_associate",
		"terrorism", "narcotics", "proliferation", "cyber_crime",
		"financial_crime", "corruption", "money_laundering",
		"regulatory_violation", "enforcement_action",
	}
	return subcategories[rand.Intn(len(subcategories))]
}

func (wc *WorldCheckMock) generatePEPSubcategory() string {
	subcategories := []string{
		"head_of_state", "government_minister", "senior_judicial_official",
		"senior_military_official", "senior_executive_of_state_owned_enterprise",
		"family_member", "close_associate",
	}
	return subcategories[rand.Intn(len(subcategories))]
}

func (wc *WorldCheckMock) generateSanctionsSubcategory() string {
	subcategories := []string{
		"ofac_sdn", "un_security_council", "eu_sanctions", "uk_sanctions",
		"terrorism", "narcotics", "proliferation", "cyber_crime",
	}
	return subcategories[rand.Intn(len(subcategories))]
}

func (wc *WorldCheckMock) generateAdverseMediaSubcategory() string {
	subcategories := []string{
		"financial_crime", "corruption", "money_laundering", "fraud",
		"regulatory_violation", "enforcement_action", "reputational_risk",
		"litigation", "investigation",
	}
	return subcategories[rand.Intn(len(subcategories))]
}

func (wc *WorldCheckMock) generateSource() string {
	sources := []string{
		"OFAC", "UN Security Council", "EU Sanctions", "UK Sanctions",
		"PEP Database", "Government Records", "Public Records",
		"News Database", "Media Monitoring", "Regulatory Filings",
	}
	return sources[rand.Intn(len(sources))]
}

func (wc *WorldCheckMock) generateSanctionsSource() string {
	sources := []string{"OFAC", "UN Security Council", "EU Sanctions", "UK Sanctions"}
	return sources[rand.Intn(len(sources))]
}

func (wc *WorldCheckMock) generateAdverseMediaSource() string {
	sources := []string{"News Database", "Media Monitoring", "Public Records", "Regulatory Filings"}
	return sources[rand.Intn(len(sources))]
}

func (wc *WorldCheckMock) generateMatchDetails() string {
	details := []string{
		"Entity matches name and country of residence",
		"Partial name match with similar business activities",
		"Entity associated with sanctioned individual",
		"Entity operates in high-risk jurisdiction",
		"Entity has regulatory enforcement history",
	}
	return details[rand.Intn(len(details))]
}

func (wc *WorldCheckMock) generatePEPMatchDetails() string {
	details := []string{
		"Individual holds senior government position",
		"Family member of senior government official",
		"Close associate of politically exposed person",
		"Former government official with ongoing influence",
	}
	return details[rand.Intn(len(details))]
}

func (wc *WorldCheckMock) generateSanctionsMatchDetails() string {
	details := []string{
		"Entity listed on OFAC SDN list",
		"Entity subject to UN Security Council sanctions",
		"Entity designated under EU sanctions regime",
		"Entity blocked pursuant to Executive Order",
	}
	return details[rand.Intn(len(details))]
}

func (wc *WorldCheckMock) generateAdverseMediaMatchDetails() string {
	details := []string{
		"Entity mentioned in financial crime investigation",
		"Entity subject to regulatory enforcement action",
		"Entity involved in money laundering case",
		"Entity associated with corruption allegations",
		"Entity mentioned in fraud investigation",
	}
	return details[rand.Intn(len(details))]
}

func (wc *WorldCheckMock) generateAdditionalInfo() string {
	info := []string{
		"Additional verification recommended",
		"Monitor for updates to sanctions status",
		"Enhanced due diligence required",
		"Regular screening recommended",
		"",
	}
	return info[rand.Intn(len(info))]
}

func (wc *WorldCheckMock) generatePEPAdditionalInfo() string {
	info := []string{
		"Enhanced due diligence required for PEP relationships",
		"Senior management approval required for business relationship",
		"Ongoing monitoring of PEP status recommended",
		"",
	}
	return info[rand.Intn(len(info))]
}

func (wc *WorldCheckMock) generateSanctionsAdditionalInfo() string {
	info := []string{
		"IMMEDIATE ACTION REQUIRED - Do not proceed with business relationship",
		"Legal review required before any business activities",
		"Compliance team notification required",
		"",
	}
	return info[rand.Intn(len(info))]
}

func (wc *WorldCheckMock) generateAdverseMediaAdditionalInfo() string {
	info := []string{
		"Review adverse media for business impact",
		"Assess reputational risk before proceeding",
		"Monitor for additional adverse media coverage",
		"Consider enhanced due diligence",
		"",
	}
	return info[rand.Intn(len(info))]
}

func (wc *WorldCheckMock) generateAdverseMediaRiskLevel() string {
	levels := []string{"low", "medium", "high"}
	weights := []float64{0.3, 0.5, 0.2} // 30% low, 50% medium, 20% high

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return levels[i]
		}
	}
	return "medium"
}

func (wc *WorldCheckMock) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}

func (wc *WorldCheckMock) getAvailableSources() []string {
	return []string{
		"OFAC", "UN Security Council", "EU Sanctions", "UK Sanctions",
		"PEP Database", "Government Records", "Public Records",
		"News Database", "Media Monitoring", "Regulatory Filings",
	}
}

// AddToWatchlist adds an entity to a watchlist
func (wc *WorldCheckMock) AddToWatchlist(ctx context.Context, entityName, country, watchlistType, reason string) (*WorldCheckWatchlist, error) {
	wc.logger.Info("Adding entity to watchlist (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country),
		zap.String("watchlist_type", watchlistType))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	watchlist := &WorldCheckWatchlist{
		WatchlistID:       fmt.Sprintf("WC_WATCH_%d", time.Now().UnixNano()),
		EntityID:          fmt.Sprintf("WC_ENTITY_%d", rand.Intn(100000)),
		EntityName:        entityName,
		WatchlistType:     watchlistType,
		RiskLevel:         wc.generateRiskLevel(),
		Reason:            reason,
		CreatedBy:         "system",
		CreatedAt:         time.Now(),
		LastUpdated:       time.Now(),
		IsActive:          true,
		NotificationEmail: "alerts@company.com",
	}

	wc.logger.Info("Entity added to watchlist (mock)",
		zap.String("watchlist_id", watchlist.WatchlistID),
		zap.String("entity_name", entityName))

	return watchlist, nil
}

// RemoveFromWatchlist removes an entity from a watchlist
func (wc *WorldCheckMock) RemoveFromWatchlist(ctx context.Context, watchlistID string) error {
	wc.logger.Info("Removing entity from watchlist (mock)",
		zap.String("watchlist_id", watchlistID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)

	wc.logger.Info("Entity removed from watchlist (mock)",
		zap.String("watchlist_id", watchlistID))

	return nil
}

// GetWatchlist retrieves watchlist entries
func (wc *WorldCheckMock) GetWatchlist(ctx context.Context, watchlistType string) ([]WorldCheckWatchlist, error) {
	wc.logger.Info("Getting watchlist (mock)",
		zap.String("watchlist_type", watchlistType))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

	// Generate mock watchlist entries
	var watchlist []WorldCheckWatchlist
	numEntries := rand.Intn(5) + 1

	for i := 0; i < numEntries; i++ {
		entry := WorldCheckWatchlist{
			WatchlistID:       fmt.Sprintf("WC_WATCH_%d_%d", time.Now().UnixNano(), i),
			EntityID:          fmt.Sprintf("WC_ENTITY_%d", rand.Intn(100000)),
			EntityName:        fmt.Sprintf("Entity %d", i+1),
			WatchlistType:     watchlistType,
			RiskLevel:         wc.generateRiskLevel(),
			Reason:            "Automated risk assessment",
			CreatedBy:         "system",
			CreatedAt:         time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			LastUpdated:       time.Now().Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
			IsActive:          true,
			NotificationEmail: "alerts@company.com",
		}

		// Add expiry date for some entries
		if rand.Float64() > 0.7 {
			expiry := time.Now().Add(time.Duration(rand.Intn(365)+30) * 24 * time.Hour)
			entry.ExpiryDate = &expiry
		}

		watchlist = append(watchlist, entry)
	}

	wc.logger.Info("Watchlist retrieved (mock)",
		zap.String("watchlist_type", watchlistType),
		zap.Int("entry_count", len(watchlist)))

	return watchlist, nil
}

// IsHealthy checks if the World-Check service is healthy
func (wc *WorldCheckMock) IsHealthy(ctx context.Context) error {
	wc.logger.Info("Checking World-Check service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}
