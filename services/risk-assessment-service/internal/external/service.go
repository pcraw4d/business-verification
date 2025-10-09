package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ExternalDataService provides unified access to external data sources
type ExternalDataService struct {
	newsAPI        *NewsAPIClient
	openCorporates *OpenCorporatesClient
	government     *GovernmentClient
	logger         *zap.Logger
	config         *ExternalDataConfig
}

// ExternalDataConfig holds configuration for external data sources
type ExternalDataConfig struct {
	NewsAPIKey        string        `json:"news_api_key"`
	OpenCorporatesKey string        `json:"opencorporates_key"`
	GovernmentAPIKey  string        `json:"government_api_key"`
	Timeout           time.Duration `json:"timeout"`
	EnableNewsAPI     bool          `json:"enable_news_api"`
	EnableOpenCorporates bool       `json:"enable_opencorporates"`
	EnableGovernment  bool          `json:"enable_government"`
}

// ExternalDataResult represents the combined result from all external sources
type ExternalDataResult struct {
	BusinessName        string                      `json:"business_name"`
	Country             string                      `json:"country"`
	Industry            string                      `json:"industry"`
	AdverseMedia        *AdverseMediaResult         `json:"adverse_media,omitempty"`
	CompanyData         *CompanySearchResult        `json:"company_data,omitempty"`
	ComplianceCheck     *ComplianceCheckResult      `json:"compliance_check,omitempty"`
	OverallRiskScore    float64                     `json:"overall_risk_score"`
	RiskFactors         []models.RiskFactor         `json:"risk_factors"`
	DataQuality         string                      `json:"data_quality"`
	LastChecked         time.Time                   `json:"last_checked"`
	ProcessingTime      time.Duration               `json:"processing_time"`
}

// NewExternalDataService creates a new external data service
func NewExternalDataService(config *ExternalDataConfig, logger *zap.Logger) *ExternalDataService {
	var newsAPI *NewsAPIClient
	var openCorporates *OpenCorporatesClient
	var government *GovernmentClient

	// Initialize clients based on configuration
	if config.EnableNewsAPI && config.NewsAPIKey != "" {
		newsAPI = NewNewsAPIClient(config.NewsAPIKey, logger)
	}

	if config.EnableOpenCorporates && config.OpenCorporatesKey != "" {
		openCorporates = NewOpenCorporatesClient(config.OpenCorporatesKey, logger)
	}

	if config.EnableGovernment && config.GovernmentAPIKey != "" {
		government = NewGovernmentClient(config.GovernmentAPIKey, logger)
	}

	return &ExternalDataService{
		newsAPI:        newsAPI,
		openCorporates: openCorporates,
		government:     government,
		logger:         logger,
		config:         config,
	}
}

// GatherExternalData gathers data from all available external sources
func (eds *ExternalDataService) GatherExternalData(ctx context.Context, req *models.RiskAssessmentRequest) (*ExternalDataResult, error) {
	start := time.Now()
	
	eds.logger.Info("Gathering external data",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country),
		zap.String("industry", req.Industry))

	result := &ExternalDataResult{
		BusinessName:   req.BusinessName,
		Country:        req.Country,
		Industry:       req.Industry,
		LastChecked:    time.Now(),
		RiskFactors:    []models.RiskFactor{},
	}

	// Use goroutines to gather data from multiple sources concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Gather adverse media data
	if eds.newsAPI != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			adverseMedia, err := eds.newsAPI.SearchAdverseMedia(ctx, req.BusinessName)
			if err != nil {
				eds.logger.Warn("Failed to gather adverse media data", zap.Error(err))
				return
			}
			
			mu.Lock()
			result.AdverseMedia = adverseMedia
			mu.Unlock()
		}()
	}

	// Gather company data
	if eds.openCorporates != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			companyData, err := eds.openCorporates.SearchCompany(ctx, req.BusinessName, req.Country)
			if err != nil {
				eds.logger.Warn("Failed to gather company data", zap.Error(err))
				return
			}
			
			mu.Lock()
			result.CompanyData = companyData
			mu.Unlock()
		}()
	}

	// Gather compliance data
	if eds.government != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			complianceCheck, err := eds.government.CheckSanctions(ctx, req.BusinessName, req.Country)
			if err != nil {
				eds.logger.Warn("Failed to gather compliance data", zap.Error(err))
				return
			}
			
			mu.Lock()
			result.ComplianceCheck = complianceCheck
			mu.Unlock()
		}()
	}

	// Wait for all data gathering to complete
	wg.Wait()

	// Calculate overall risk score and risk factors
	eds.calculateOverallRisk(result)

	// Determine data quality
	result.DataQuality = eds.assessDataQuality(result)

	result.ProcessingTime = time.Since(start)

	eds.logger.Info("External data gathering completed",
		zap.String("business_name", req.BusinessName),
		zap.Float64("overall_risk_score", result.OverallRiskScore),
		zap.String("data_quality", result.DataQuality),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// calculateOverallRisk calculates the overall risk score and risk factors
func (eds *ExternalDataService) calculateOverallRisk(result *ExternalDataResult) {
	riskScores := []float64{}
	riskFactors := []models.RiskFactor{}

	// Adverse media risk
	if result.AdverseMedia != nil {
		riskScores = append(riskScores, result.AdverseMedia.RiskScore)
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryReputational,
			Name:        "Adverse Media",
			Score:       result.AdverseMedia.RiskScore,
			Weight:      0.3,
			Description: fmt.Sprintf("Found %d adverse media articles", result.AdverseMedia.TotalArticles),
			Source:      "NewsAPI",
			Confidence:  0.9,
		})
	}

	// Company data risk
	if result.CompanyData != nil {
		riskScores = append(riskScores, result.CompanyData.RiskScore)
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryOperational,
			Name:        "Company Registration",
			Score:       result.CompanyData.RiskScore,
			Weight:      0.4,
			Description: fmt.Sprintf("Company status: %s", result.CompanyData.ComplianceStatus),
			Source:      "OpenCorporates",
			Confidence:  0.95,
		})
	}

	// Compliance risk
	if result.ComplianceCheck != nil {
		riskScores = append(riskScores, result.ComplianceCheck.RiskScore)
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Name:        "Regulatory Compliance",
			Score:       result.ComplianceCheck.RiskScore,
			Weight:      0.3,
			Description: fmt.Sprintf("Compliance status: %s", result.ComplianceCheck.ComplianceStatus),
			Source:      "Government Databases",
			Confidence:  0.98,
		})
	}

	// Calculate weighted average risk score
	if len(riskScores) > 0 {
		totalWeight := 0.0
		weightedSum := 0.0

		for i, score := range riskScores {
			weight := riskFactors[i].Weight
			weightedSum += score * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			result.OverallRiskScore = weightedSum / totalWeight
		}
	}

	result.RiskFactors = riskFactors
}

// assessDataQuality assesses the quality of gathered data
func (eds *ExternalDataService) assessDataQuality(result *ExternalDataResult) string {
	sources := 0
	highQualitySources := 0

	if result.AdverseMedia != nil {
		sources++
		if result.AdverseMedia.TotalArticles > 0 {
			highQualitySources++
		}
	}

	if result.CompanyData != nil {
		sources++
		if result.CompanyData.TotalResults > 0 {
			highQualitySources++
		}
	}

	if result.ComplianceCheck != nil {
		sources++
		if result.ComplianceCheck.TotalRecords > 0 {
			highQualitySources++
		}
	}

	if sources == 0 {
		return "NO_DATA"
	}

	qualityRatio := float64(highQualitySources) / float64(sources)

	if qualityRatio >= 0.8 {
		return "HIGH"
	} else if qualityRatio >= 0.5 {
		return "MEDIUM"
	} else {
		return "LOW"
	}
}

// GetAdverseMedia gets adverse media data for a business
func (eds *ExternalDataService) GetAdverseMedia(ctx context.Context, businessName string) (*AdverseMediaResult, error) {
	if eds.newsAPI == nil {
		return nil, fmt.Errorf("NewsAPI client not configured")
	}

	return eds.newsAPI.SearchAdverseMedia(ctx, businessName)
}

// GetCompanyData gets company data from OpenCorporates
func (eds *ExternalDataService) GetCompanyData(ctx context.Context, businessName, country string) (*CompanySearchResult, error) {
	if eds.openCorporates == nil {
		return nil, fmt.Errorf("OpenCorporates client not configured")
	}

	return eds.openCorporates.SearchCompany(ctx, businessName, country)
}

// GetComplianceData gets compliance data from government databases
func (eds *ExternalDataService) GetComplianceData(ctx context.Context, businessName, country string) (*ComplianceCheckResult, error) {
	if eds.government == nil {
		return nil, fmt.Errorf("Government client not configured")
	}

	return eds.government.CheckSanctions(ctx, businessName, country)
}

// IsHealthy checks if all external data sources are healthy
func (eds *ExternalDataService) IsHealthy(ctx context.Context) error {
	var errors []error

	if eds.newsAPI != nil {
		if err := eds.newsAPI.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("NewsAPI: %w", err))
		}
	}

	if eds.openCorporates != nil {
		if err := eds.openCorporates.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("OpenCorporates: %w", err))
		}
	}

	if eds.government != nil {
		if err := eds.government.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("Government: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("external data sources unhealthy: %v", errors)
	}

	return nil
}

// GetAvailableSources returns the list of available external data sources
func (eds *ExternalDataService) GetAvailableSources() []string {
	sources := []string{}

	if eds.newsAPI != nil {
		sources = append(sources, "NewsAPI")
	}

	if eds.openCorporates != nil {
		sources = append(sources, "OpenCorporates")
	}

	if eds.government != nil {
		sources = append(sources, "Government Databases")
	}

	return sources
}

// Close closes all external data source connections
func (eds *ExternalDataService) Close() {
	if eds.newsAPI != nil {
		eds.newsAPI.Close()
	}

	if eds.openCorporates != nil {
		eds.openCorporates.Close()
	}

	if eds.government != nil {
		eds.government.Close()
	}
}
