package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/external/ofac"
	"kyb-platform/services/risk-assessment-service/internal/external/thomson_reuters"
	"kyb-platform/services/risk-assessment-service/internal/external/worldcheck"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ExternalAPIManager coordinates all external API calls and provides a unified interface
type ExternalAPIManager struct {
	thomsonReuters *thomson_reuters.ThomsonReutersMock
	ofac           *ofac.OFACMock
	worldCheck     *worldcheck.WorldCheckMock
	logger         *zap.Logger
	config         *ExternalAPIManagerConfig
}

// ExternalAPIManagerConfig holds configuration for the external API manager
type ExternalAPIManagerConfig struct {
	ThomsonReuters *ThomsonReutersConfig `json:"thomson_reuters"`
	OFAC           *OFACConfig           `json:"ofac"`
	WorldCheck     *WorldCheckConfig     `json:"worldcheck"`
	Timeout        time.Duration         `json:"timeout"`
	MaxRetries     int                   `json:"max_retries"`
	EnableCache    bool                  `json:"enable_cache"`
	CacheTTL       time.Duration         `json:"cache_ttl"`
}

// ThomsonReutersConfig holds configuration for Thomson Reuters API
type ThomsonReutersConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// OFACConfig holds configuration for OFAC API
type OFACConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// WorldCheckConfig holds configuration for World-Check API
type WorldCheckConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
}

// PremiumExternalDataResult represents the combined result from all premium external sources
type PremiumExternalDataResult struct {
	BusinessName     string                                `json:"business_name"`
	Country          string                                `json:"country"`
	Industry         string                                `json:"industry"`
	ThomsonReuters   *thomson_reuters.ThomsonReutersResult `json:"thomson_reuters,omitempty"`
	OFAC             *ofac.OFACResult                      `json:"ofac,omitempty"`
	WorldCheck       *worldcheck.WorldCheckResult          `json:"worldcheck,omitempty"`
	OverallRiskScore float64                               `json:"overall_risk_score"`
	RiskFactors      []models.RiskFactor                   `json:"risk_factors"`
	DataQuality      string                                `json:"data_quality"`
	LastChecked      time.Time                             `json:"last_checked"`
	ProcessingTime   time.Duration                         `json:"processing_time"`
	APIStatus        map[string]string                     `json:"api_status"`
}

// NewExternalAPIManager creates a new external API manager
func NewExternalAPIManager(config *ExternalAPIManagerConfig, logger *zap.Logger) *ExternalAPIManager {
	manager := &ExternalAPIManager{
		logger: logger,
		config: config,
	}

	// Initialize Thomson Reuters client
	if config.ThomsonReuters != nil && config.ThomsonReuters.Enabled {
		trConfig := &thomson_reuters.ThomsonReutersConfig{
			APIKey:    config.ThomsonReuters.APIKey,
			BaseURL:   config.ThomsonReuters.BaseURL,
			Timeout:   config.ThomsonReuters.Timeout,
			RateLimit: config.ThomsonReuters.RateLimit,
			Enabled:   config.ThomsonReuters.Enabled,
		}
		manager.thomsonReuters = thomson_reuters.NewThomsonReutersMock(trConfig, logger)
		logger.Info("Thomson Reuters client initialized")
	}

	// Initialize OFAC client
	if config.OFAC != nil && config.OFAC.Enabled {
		ofacConfig := &ofac.OFACConfig{
			APIKey:    config.OFAC.APIKey,
			BaseURL:   config.OFAC.BaseURL,
			Timeout:   config.OFAC.Timeout,
			RateLimit: config.OFAC.RateLimit,
			Enabled:   config.OFAC.Enabled,
		}
		manager.ofac = ofac.NewOFACMock(ofacConfig, logger)
		logger.Info("OFAC client initialized")
	}

	// Initialize World-Check client
	if config.WorldCheck != nil && config.WorldCheck.Enabled {
		wcConfig := &worldcheck.WorldCheckConfig{
			APIKey:    config.WorldCheck.APIKey,
			BaseURL:   config.WorldCheck.BaseURL,
			Timeout:   config.WorldCheck.Timeout,
			RateLimit: config.WorldCheck.RateLimit,
			Enabled:   config.WorldCheck.Enabled,
		}
		manager.worldCheck = worldcheck.NewWorldCheckMock(wcConfig, logger)
		logger.Info("World-Check client initialized")
	}

	logger.Info("External API manager initialized",
		zap.Bool("thomson_reuters_enabled", manager.thomsonReuters != nil),
		zap.Bool("ofac_enabled", manager.ofac != nil),
		zap.Bool("worldcheck_enabled", manager.worldCheck != nil))

	return manager
}

// GetComprehensiveData retrieves data from all enabled premium external APIs
func (eam *ExternalAPIManager) GetComprehensiveData(ctx context.Context, business *models.RiskAssessmentRequest) (*PremiumExternalDataResult, error) {
	startTime := time.Now()
	eam.logger.Info("Getting comprehensive premium external data",
		zap.String("business_name", business.BusinessName),
		zap.String("country", business.Country))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, eam.config.Timeout)
	defer cancel()

	// Collect results from all APIs in parallel
	type apiResult struct {
		source string
		data   interface{}
		err    error
	}

	results := make(chan apiResult, 3)
	var wg sync.WaitGroup

	// Get Thomson Reuters data
	if eam.thomsonReuters != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := eam.thomsonReuters.GetComprehensiveData(ctx, business.BusinessName, business.Country)
			results <- apiResult{source: "thomson_reuters", data: data, err: err}
		}()
	}

	// Get OFAC data
	if eam.ofac != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := eam.ofac.GetComprehensiveData(ctx, business.BusinessName, business.Country)
			results <- apiResult{source: "ofac", data: data, err: err}
		}()
	}

	// Get World-Check data
	if eam.worldCheck != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := eam.worldCheck.GetComprehensiveData(ctx, business.BusinessName, business.Country)
			results <- apiResult{source: "worldcheck", data: data, err: err}
		}()
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var thomsonReutersResult *thomson_reuters.ThomsonReutersResult
	var ofacResult *ofac.OFACResult
	var worldCheckResult *worldcheck.WorldCheckResult
	apiStatus := make(map[string]string)

	for result := range results {
		if result.err != nil {
			eam.logger.Warn("Failed to get data from external API",
				zap.String("source", result.source),
				zap.Error(result.err))
			apiStatus[result.source] = "error"
		} else {
			apiStatus[result.source] = "success"
			switch result.source {
			case "thomson_reuters":
				if tr, ok := result.data.(*thomson_reuters.ThomsonReutersResult); ok {
					thomsonReutersResult = tr
				}
			case "ofac":
				if of, ok := result.data.(*ofac.OFACResult); ok {
					ofacResult = of
				}
			case "worldcheck":
				if wc, ok := result.data.(*worldcheck.WorldCheckResult); ok {
					worldCheckResult = wc
				}
			}
		}
	}

	// Generate risk factors from all sources
	riskFactors := eam.generateCombinedRiskFactors(thomsonReutersResult, ofacResult, worldCheckResult)

	// Calculate overall risk score
	overallRiskScore := eam.calculateOverallRiskScore(riskFactors)

	// Determine data quality
	dataQuality := eam.determineDataQuality(thomsonReutersResult, ofacResult, worldCheckResult)

	// Create comprehensive result
	result := &PremiumExternalDataResult{
		BusinessName:     business.BusinessName,
		Country:          business.Country,
		Industry:         business.Industry,
		ThomsonReuters:   thomsonReutersResult,
		OFAC:             ofacResult,
		WorldCheck:       worldCheckResult,
		OverallRiskScore: overallRiskScore,
		RiskFactors:      riskFactors,
		DataQuality:      dataQuality,
		LastChecked:      time.Now(),
		ProcessingTime:   time.Since(startTime),
		APIStatus:        apiStatus,
	}

	eam.logger.Info("Comprehensive premium external data retrieved",
		zap.String("business_name", business.BusinessName),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.String("data_quality", result.DataQuality),
		zap.Float64("overall_risk_score", result.OverallRiskScore))

	return result, nil
}

// GetThomsonReutersData retrieves data from Thomson Reuters only
func (eam *ExternalAPIManager) GetThomsonReutersData(ctx context.Context, business *models.RiskAssessmentRequest) (*thomson_reuters.ThomsonReutersResult, error) {
	if eam.thomsonReuters == nil {
		return nil, fmt.Errorf("Thomson Reuters client not enabled")
	}

	eam.logger.Info("Getting Thomson Reuters data",
		zap.String("business_name", business.BusinessName))

	return eam.thomsonReuters.GetComprehensiveData(ctx, business.BusinessName, business.Country)
}

// GetOFACData retrieves data from OFAC only
func (eam *ExternalAPIManager) GetOFACData(ctx context.Context, business *models.RiskAssessmentRequest) (*ofac.OFACResult, error) {
	if eam.ofac == nil {
		return nil, fmt.Errorf("OFAC client not enabled")
	}

	eam.logger.Info("Getting OFAC data",
		zap.String("business_name", business.BusinessName))

	return eam.ofac.GetComprehensiveData(ctx, business.BusinessName, business.Country)
}

// GetWorldCheckData retrieves data from World-Check only
func (eam *ExternalAPIManager) GetWorldCheckData(ctx context.Context, business *models.RiskAssessmentRequest) (*worldcheck.WorldCheckResult, error) {
	if eam.worldCheck == nil {
		return nil, fmt.Errorf("World-Check client not enabled")
	}

	eam.logger.Info("Getting World-Check data",
		zap.String("business_name", business.BusinessName))

	return eam.worldCheck.GetComprehensiveData(ctx, business.BusinessName, business.Country)
}

// GetAPIStatus returns the status of all external APIs
func (eam *ExternalAPIManager) GetAPIStatus() map[string]string {
	status := make(map[string]string)

	if eam.thomsonReuters != nil {
		status["thomson_reuters"] = "enabled"
	} else {
		status["thomson_reuters"] = "disabled"
	}

	if eam.ofac != nil {
		status["ofac"] = "enabled"
	} else {
		status["ofac"] = "disabled"
	}

	if eam.worldCheck != nil {
		status["worldcheck"] = "enabled"
	} else {
		status["worldcheck"] = "disabled"
	}

	return status
}

// GetSupportedAPIs returns a list of supported external APIs
func (eam *ExternalAPIManager) GetSupportedAPIs() []string {
	apis := []string{}

	if eam.thomsonReuters != nil {
		apis = append(apis, "thomson_reuters")
	}

	if eam.ofac != nil {
		apis = append(apis, "ofac")
	}

	if eam.worldCheck != nil {
		apis = append(apis, "worldcheck")
	}

	return apis
}

// Helper methods

func (eam *ExternalAPIManager) generateCombinedRiskFactors(
	thomsonReutersResult *thomson_reuters.ThomsonReutersResult,
	ofacResult *ofac.OFACResult,
	worldCheckResult *worldcheck.WorldCheckResult,
) []models.RiskFactor {
	var riskFactors []models.RiskFactor

	// Add Thomson Reuters risk factors
	if thomsonReutersResult != nil {
		trFactors := eam.thomsonReuters.GenerateRiskFactors(thomsonReutersResult)
		riskFactors = append(riskFactors, trFactors...)
	}

	// Add OFAC risk factors
	if ofacResult != nil {
		ofacFactors := eam.ofac.GenerateRiskFactors(ofacResult)
		riskFactors = append(riskFactors, ofacFactors...)
	}

	// Add World-Check risk factors
	if worldCheckResult != nil {
		wcFactors := eam.worldCheck.GenerateRiskFactors(worldCheckResult)
		riskFactors = append(riskFactors, wcFactors...)
	}

	// Remove duplicates and merge similar factors
	riskFactors = eam.mergeSimilarRiskFactors(riskFactors)

	return riskFactors
}

func (eam *ExternalAPIManager) mergeSimilarRiskFactors(factors []models.RiskFactor) []models.RiskFactor {
	// Simple deduplication based on name and category
	seen := make(map[string]bool)
	var merged []models.RiskFactor

	for _, factor := range factors {
		key := fmt.Sprintf("%s_%s", factor.Category, factor.Name)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, factor)
		}
	}

	return merged
}

func (eam *ExternalAPIManager) calculateOverallRiskScore(factors []models.RiskFactor) float64 {
	if len(factors) == 0 {
		return 0.5 // Default risk score
	}

	totalWeightedScore := 0.0
	totalWeight := 0.0

	for _, factor := range factors {
		totalWeightedScore += factor.Score * factor.Weight
		totalWeight += factor.Weight
	}

	if totalWeight == 0 {
		return 0.5
	}

	overallScore := totalWeightedScore / totalWeight

	// Ensure score is within bounds
	if overallScore > 1.0 {
		overallScore = 1.0
	}
	if overallScore < 0.0 {
		overallScore = 0.0
	}

	return overallScore
}

func (eam *ExternalAPIManager) determineDataQuality(
	thomsonReutersResult *thomson_reuters.ThomsonReutersResult,
	ofacResult *ofac.OFACResult,
	worldCheckResult *worldcheck.WorldCheckResult,
) string {
	qualityScores := []string{}

	if thomsonReutersResult != nil {
		qualityScores = append(qualityScores, thomsonReutersResult.DataQuality)
	}

	if ofacResult != nil {
		qualityScores = append(qualityScores, ofacResult.DataQuality)
	}

	if worldCheckResult != nil {
		qualityScores = append(qualityScores, worldCheckResult.DataQuality)
	}

	if len(qualityScores) == 0 {
		return "unknown"
	}

	// Determine overall quality based on individual scores
	excellentCount := 0
	goodCount := 0
	averageCount := 0

	for _, quality := range qualityScores {
		switch quality {
		case "excellent":
			excellentCount++
		case "good":
			goodCount++
		case "average":
			averageCount++
		}
	}

	if excellentCount > 0 {
		return "excellent"
	} else if goodCount > 0 {
		return "good"
	} else {
		return "average"
	}
}

// Health check methods

// HealthCheck performs a health check on all enabled external APIs
func (eam *ExternalAPIManager) HealthCheck(ctx context.Context) map[string]bool {
	healthStatus := make(map[string]bool)

	// Check Thomson Reuters
	if eam.thomsonReuters != nil {
		healthStatus["thomson_reuters"] = eam.checkThomsonReutersHealth(ctx)
	}

	// Check OFAC
	if eam.ofac != nil {
		healthStatus["ofac"] = eam.checkOFACHealth(ctx)
	}

	// Check World-Check
	if eam.worldCheck != nil {
		healthStatus["worldcheck"] = eam.checkWorldCheckHealth(ctx)
	}

	return healthStatus
}

func (eam *ExternalAPIManager) checkThomsonReutersHealth(ctx context.Context) bool {
	// Simple health check - try to get a basic profile
	_, err := eam.thomsonReuters.GetCompanyProfile(ctx, "Health Check", "US")
	return err == nil
}

func (eam *ExternalAPIManager) checkOFACHealth(ctx context.Context) bool {
	// Simple health check - try to search sanctions
	_, err := eam.ofac.SearchSanctions(ctx, "Health Check", "US")
	return err == nil
}

func (eam *ExternalAPIManager) checkWorldCheckHealth(ctx context.Context) bool {
	// Simple health check - try to get a profile
	_, err := eam.worldCheck.SearchProfile(ctx, "Health Check", "US")
	return err == nil
}
