package thomson_reuters

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ThomsonReutersMock provides mock implementation of Thomson Reuters API
type ThomsonReutersMock struct {
	logger     *zap.Logger
	config     *ThomsonReutersConfig
	worldCheck *WorldCheckMock
}

// ThomsonReutersConfig holds configuration for Thomson Reuters API
type ThomsonReutersConfig struct {
	APIKey           string        `json:"api_key"`
	BaseURL          string        `json:"base_url"`
	Timeout          time.Duration `json:"timeout"`
	RateLimit        int           `json:"rate_limit_per_minute"`
	Enabled          bool          `json:"enabled"`
	EnableWorldCheck bool          `json:"enable_worldcheck"`
}

// CompanyProfile represents a company profile from Thomson Reuters
type CompanyProfile struct {
	CompanyID           string                 `json:"company_id"`
	CompanyName         string                 `json:"company_name"`
	Industry            string                 `json:"industry"`
	Country             string                 `json:"country"`
	FoundedYear         int                    `json:"founded_year"`
	EmployeeCount       int                    `json:"employee_count"`
	Revenue             float64                `json:"revenue"`
	MarketCap           float64                `json:"market_cap"`
	BusinessDescription string                 `json:"business_description"`
	Website             string                 `json:"website"`
	Address             string                 `json:"address"`
	Phone               string                 `json:"phone"`
	Email               string                 `json:"email"`
	LastUpdated         time.Time              `json:"last_updated"`
	DataQuality         string                 `json:"data_quality"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// FinancialData represents financial information from Thomson Reuters
type FinancialData struct {
	CompanyID          string    `json:"company_id"`
	FiscalYear         int       `json:"fiscal_year"`
	Revenue            float64   `json:"revenue"`
	NetIncome          float64   `json:"net_income"`
	TotalAssets        float64   `json:"total_assets"`
	TotalLiabilities   float64   `json:"total_liabilities"`
	ShareholdersEquity float64   `json:"shareholders_equity"`
	CashFlow           float64   `json:"cash_flow"`
	EBITDA             float64   `json:"ebitda"`
	LastUpdated        time.Time `json:"last_updated"`
}

// FinancialRatios represents financial ratios from Thomson Reuters
type FinancialRatios struct {
	CompanyID         string    `json:"company_id"`
	CurrentRatio      float64   `json:"current_ratio"`
	QuickRatio        float64   `json:"quick_ratio"`
	DebtToEquityRatio float64   `json:"debt_to_equity_ratio"`
	ReturnOnEquity    float64   `json:"return_on_equity"`
	ReturnOnAssets    float64   `json:"return_on_assets"`
	GrossProfitMargin float64   `json:"gross_profit_margin"`
	NetProfitMargin   float64   `json:"net_profit_margin"`
	AssetTurnover     float64   `json:"asset_turnover"`
	InventoryTurnover float64   `json:"inventory_turnover"`
	LastUpdated       time.Time `json:"last_updated"`
}

// RiskMetrics represents risk metrics from Thomson Reuters
type RiskMetrics struct {
	CompanyID        string    `json:"company_id"`
	OverallRiskScore float64   `json:"overall_risk_score"`
	FinancialRisk    float64   `json:"financial_risk"`
	OperationalRisk  float64   `json:"operational_risk"`
	MarketRisk       float64   `json:"market_risk"`
	CreditRisk       float64   `json:"credit_risk"`
	LiquidityRisk    float64   `json:"liquidity_risk"`
	RegulatoryRisk   float64   `json:"regulatory_risk"`
	ESGRisk          float64   `json:"esg_risk"`
	RiskTrend        string    `json:"risk_trend"` // "improving", "stable", "deteriorating"
	LastUpdated      time.Time `json:"last_updated"`
}

// ESGScore represents ESG (Environmental, Social, Governance) scores
type ESGScore struct {
	CompanyID       string    `json:"company_id"`
	OverallESGScore float64   `json:"overall_esg_score"`
	Environmental   float64   `json:"environmental"`
	Social          float64   `json:"social"`
	Governance      float64   `json:"governance"`
	ESGLevel        string    `json:"esg_level"` // "excellent", "good", "average", "poor"
	LastUpdated     time.Time `json:"last_updated"`
}

// ExecutiveInfo represents executive information
type ExecutiveInfo struct {
	CompanyID   string      `json:"company_id"`
	Executives  []Executive `json:"executives"`
	LastUpdated time.Time   `json:"last_updated"`
}

// Executive represents an individual executive
type Executive struct {
	Name         string  `json:"name"`
	Title        string  `json:"title"`
	Tenure       int     `json:"tenure_years"`
	Experience   string  `json:"experience"`
	Education    string  `json:"education"`
	Compensation float64 `json:"compensation"`
}

// OwnershipStructure represents ownership information
type OwnershipStructure struct {
	CompanyID     string      `json:"company_id"`
	OwnershipData []Ownership `json:"ownership_data"`
	LastUpdated   time.Time   `json:"last_updated"`
}

// Ownership represents ownership information
type Ownership struct {
	OwnerName           string  `json:"owner_name"`
	OwnerType           string  `json:"owner_type"` // "individual", "institution", "government"
	OwnershipPercentage float64 `json:"ownership_percentage"`
	Shares              int64   `json:"shares"`
}

// ThomsonReutersResult represents the combined result from Thomson Reuters
type ThomsonReutersResult struct {
	CompanyProfile     *CompanyProfile        `json:"company_profile,omitempty"`
	FinancialData      *FinancialData         `json:"financial_data,omitempty"`
	FinancialRatios    *FinancialRatios       `json:"financial_ratios,omitempty"`
	RiskMetrics        *RiskMetrics           `json:"risk_metrics,omitempty"`
	ESGScore           *ESGScore              `json:"esg_score,omitempty"`
	ExecutiveInfo      *ExecutiveInfo         `json:"executive_info,omitempty"`
	OwnershipStructure *OwnershipStructure    `json:"ownership_structure,omitempty"`
	DataQuality        string                 `json:"data_quality"`
	LastChecked        time.Time              `json:"last_checked"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NewThomsonReutersMock creates a new Thomson Reuters mock client
func NewThomsonReutersMock(config *ThomsonReutersConfig, logger *zap.Logger) *ThomsonReutersMock {
	// Create World-Check client if enabled
	var worldCheck *WorldCheckMock
	if config.EnableWorldCheck {
		worldCheckConfig := &WorldCheckConfig{
			APIKey:          config.APIKey,
			BaseURL:         config.BaseURL,
			Timeout:         config.Timeout,
			RateLimit:       config.RateLimit,
			Enabled:         true,
			EnablePEP:       true,
			EnableSanctions: true,
			EnableAdverse:   true,
		}
		worldCheck = NewWorldCheckClient(worldCheckConfig, logger)
	}

	return &ThomsonReutersMock{
		logger:     logger,
		config:     config,
		worldCheck: worldCheck,
	}
}

// GetCompanyProfile retrieves company profile information
func (tr *ThomsonReutersMock) GetCompanyProfile(ctx context.Context, businessName, country string) (*CompanyProfile, error) {
	tr.logger.Info("Getting Thomson Reuters company profile (mock)",
		zap.String("business_name", businessName),
		zap.String("country", country))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(500)+100) * time.Millisecond)

	// Generate mock company profile
	profile := &CompanyProfile{
		CompanyID:           tr.generateCompanyID(businessName),
		CompanyName:         businessName,
		Industry:            tr.detectIndustry(businessName),
		Country:             country,
		FoundedYear:         rand.Intn(50) + 1970,
		EmployeeCount:       tr.generateEmployeeCount(),
		Revenue:             tr.generateRevenue(),
		MarketCap:           tr.generateMarketCap(),
		BusinessDescription: tr.generateBusinessDescription(businessName),
		Website:             fmt.Sprintf("https://www.%s.com", tr.sanitizeForURL(businessName)),
		Address:             tr.generateAddress(country),
		Phone:               tr.generatePhone(country),
		Email:               fmt.Sprintf("info@%s.com", tr.sanitizeForURL(businessName)),
		LastUpdated:         time.Now(),
		DataQuality:         tr.generateDataQuality(),
		Metadata: map[string]interface{}{
			"source":           "thomson_reuters_mock",
			"confidence_score": rand.Float64()*0.3 + 0.7, // 0.7-1.0
			"data_freshness":   "current",
		},
	}

	tr.logger.Info("Thomson Reuters company profile retrieved (mock)",
		zap.String("company_id", profile.CompanyID),
		zap.String("data_quality", profile.DataQuality))

	return profile, nil
}

// GetFinancialData retrieves financial data
func (tr *ThomsonReutersMock) GetFinancialData(ctx context.Context, companyID string) (*FinancialData, error) {
	tr.logger.Info("Getting Thomson Reuters financial data (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// Generate mock financial data
	revenue := tr.generateRevenue()
	financialData := &FinancialData{
		CompanyID:          companyID,
		FiscalYear:         time.Now().Year() - 1,
		Revenue:            revenue,
		NetIncome:          revenue * (rand.Float64()*0.2 + 0.05), // 5-25% margin
		TotalAssets:        revenue * (rand.Float64()*2 + 1),      // 1-3x revenue
		TotalLiabilities:   revenue * (rand.Float64()*1.5 + 0.5),  // 0.5-2x revenue
		ShareholdersEquity: revenue * (rand.Float64()*1 + 0.5),    // 0.5-1.5x revenue
		CashFlow:           revenue * (rand.Float64()*0.3 + 0.1),  // 10-40% of revenue
		EBITDA:             revenue * (rand.Float64()*0.4 + 0.1),  // 10-50% of revenue
		LastUpdated:        time.Now(),
	}

	tr.logger.Info("Thomson Reuters financial data retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Float64("revenue", financialData.Revenue))

	return financialData, nil
}

// GetFinancialRatios retrieves financial ratios
func (tr *ThomsonReutersMock) GetFinancialRatios(ctx context.Context, companyID string) (*FinancialRatios, error) {
	tr.logger.Info("Getting Thomson Reuters financial ratios (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)

	// Generate mock financial ratios
	ratios := &FinancialRatios{
		CompanyID:         companyID,
		CurrentRatio:      rand.Float64()*2 + 1,      // 1.0-3.0
		QuickRatio:        rand.Float64()*1.5 + 0.5,  // 0.5-2.0
		DebtToEquityRatio: rand.Float64()*2 + 0.1,    // 0.1-2.1
		ReturnOnEquity:    rand.Float64()*0.3 + 0.05, // 5-35%
		ReturnOnAssets:    rand.Float64()*0.2 + 0.02, // 2-22%
		GrossProfitMargin: rand.Float64()*0.4 + 0.2,  // 20-60%
		NetProfitMargin:   rand.Float64()*0.2 + 0.05, // 5-25%
		AssetTurnover:     rand.Float64()*2 + 0.5,    // 0.5-2.5
		InventoryTurnover: rand.Float64()*10 + 2,     // 2-12
		LastUpdated:       time.Now(),
	}

	tr.logger.Info("Thomson Reuters financial ratios retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Float64("current_ratio", ratios.CurrentRatio))

	return ratios, nil
}

// GetRiskMetrics retrieves risk metrics
func (tr *ThomsonReutersMock) GetRiskMetrics(ctx context.Context, companyID string) (*RiskMetrics, error) {
	tr.logger.Info("Getting Thomson Reuters risk metrics (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

	// Generate mock risk metrics
	financialRisk := rand.Float64()*0.8 + 0.1   // 0.1-0.9
	operationalRisk := rand.Float64()*0.8 + 0.1 // 0.1-0.9
	marketRisk := rand.Float64()*0.8 + 0.1      // 0.1-0.9
	creditRisk := rand.Float64()*0.8 + 0.1      // 0.1-0.9
	liquidityRisk := rand.Float64()*0.8 + 0.1   // 0.1-0.9
	regulatoryRisk := rand.Float64()*0.8 + 0.1  // 0.1-0.9
	esgRisk := rand.Float64()*0.8 + 0.1         // 0.1-0.9

	overallRisk := (financialRisk + operationalRisk + marketRisk + creditRisk + liquidityRisk + regulatoryRisk + esgRisk) / 7

	riskMetrics := &RiskMetrics{
		CompanyID:        companyID,
		OverallRiskScore: overallRisk,
		FinancialRisk:    financialRisk,
		OperationalRisk:  operationalRisk,
		MarketRisk:       marketRisk,
		CreditRisk:       creditRisk,
		LiquidityRisk:    liquidityRisk,
		RegulatoryRisk:   regulatoryRisk,
		ESGRisk:          esgRisk,
		RiskTrend:        tr.generateRiskTrend(),
		LastUpdated:      time.Now(),
	}

	tr.logger.Info("Thomson Reuters risk metrics retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Float64("overall_risk_score", riskMetrics.OverallRiskScore))

	return riskMetrics, nil
}

// GetESGScore retrieves ESG scores
func (tr *ThomsonReutersMock) GetESGScore(ctx context.Context, companyID string) (*ESGScore, error) {
	tr.logger.Info("Getting Thomson Reuters ESG score (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// Generate mock ESG scores
	environmental := rand.Float64() * 100 // 0-100
	social := rand.Float64() * 100        // 0-100
	governance := rand.Float64() * 100    // 0-100
	overallESG := (environmental + social + governance) / 3

	esgScore := &ESGScore{
		CompanyID:       companyID,
		OverallESGScore: overallESG,
		Environmental:   environmental,
		Social:          social,
		Governance:      governance,
		ESGLevel:        tr.calculateESGLevel(overallESG),
		LastUpdated:     time.Now(),
	}

	tr.logger.Info("Thomson Reuters ESG score retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Float64("overall_esg_score", esgScore.OverallESGScore))

	return esgScore, nil
}

// GetExecutiveInfo retrieves executive information
func (tr *ThomsonReutersMock) GetExecutiveInfo(ctx context.Context, companyID string) (*ExecutiveInfo, error) {
	tr.logger.Info("Getting Thomson Reuters executive info (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(250)+100) * time.Millisecond)

	// Generate mock executive information
	executives := []Executive{
		{
			Name:         tr.generateExecutiveName(),
			Title:        "Chief Executive Officer",
			Tenure:       rand.Intn(10) + 1,
			Experience:   tr.generateExperience(),
			Education:    tr.generateEducation(),
			Compensation: rand.Float64()*2000000 + 500000, // $500k-$2.5M
		},
		{
			Name:         tr.generateExecutiveName(),
			Title:        "Chief Financial Officer",
			Tenure:       rand.Intn(8) + 1,
			Experience:   tr.generateExperience(),
			Education:    tr.generateEducation(),
			Compensation: rand.Float64()*1500000 + 300000, // $300k-$1.8M
		},
		{
			Name:         tr.generateExecutiveName(),
			Title:        "Chief Technology Officer",
			Tenure:       rand.Intn(6) + 1,
			Experience:   tr.generateExperience(),
			Education:    tr.generateEducation(),
			Compensation: rand.Float64()*1200000 + 250000, // $250k-$1.45M
		},
	}

	executiveInfo := &ExecutiveInfo{
		CompanyID:   companyID,
		Executives:  executives,
		LastUpdated: time.Now(),
	}

	tr.logger.Info("Thomson Reuters executive info retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Int("executive_count", len(executives)))

	return executiveInfo, nil
}

// GetOwnershipStructure retrieves ownership structure
func (tr *ThomsonReutersMock) GetOwnershipStructure(ctx context.Context, companyID string) (*OwnershipStructure, error) {
	tr.logger.Info("Getting Thomson Reuters ownership structure (mock)",
		zap.String("company_id", companyID))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)

	// Generate mock ownership structure
	ownershipData := []Ownership{
		{
			OwnerName:           tr.generateOwnerName(),
			OwnerType:           "institution",
			OwnershipPercentage: rand.Float64()*30 + 10, // 10-40%
			Shares:              rand.Int63n(10000000) + 1000000,
		},
		{
			OwnerName:           tr.generateOwnerName(),
			OwnerType:           "individual",
			OwnershipPercentage: rand.Float64()*20 + 5, // 5-25%
			Shares:              rand.Int63n(5000000) + 500000,
		},
		{
			OwnerName:           tr.generateOwnerName(),
			OwnerType:           "institution",
			OwnershipPercentage: rand.Float64()*15 + 5, // 5-20%
			Shares:              rand.Int63n(3000000) + 300000,
		},
	}

	ownershipStructure := &OwnershipStructure{
		CompanyID:     companyID,
		OwnershipData: ownershipData,
		LastUpdated:   time.Now(),
	}

	tr.logger.Info("Thomson Reuters ownership structure retrieved (mock)",
		zap.String("company_id", companyID),
		zap.Int("owner_count", len(ownershipData)))

	return ownershipStructure, nil
}

// GetComprehensiveData retrieves all available data from Thomson Reuters
func (tr *ThomsonReutersMock) GetComprehensiveData(ctx context.Context, businessName, country string) (*ThomsonReutersResult, error) {
	startTime := time.Now()
	tr.logger.Info("Getting comprehensive Thomson Reuters data (mock)",
		zap.String("business_name", businessName),
		zap.String("country", country))

	// Get company profile first
	profile, err := tr.GetCompanyProfile(ctx, businessName, country)
	if err != nil {
		return nil, fmt.Errorf("failed to get company profile: %w", err)
	}

	// Get all other data in parallel
	type result struct {
		financialData      *FinancialData
		financialRatios    *FinancialRatios
		riskMetrics        *RiskMetrics
		esgScore           *ESGScore
		executiveInfo      *ExecutiveInfo
		ownershipStructure *OwnershipStructure
		err                error
	}

	results := make(chan result, 6)

	// Get financial data
	go func() {
		fd, err := tr.GetFinancialData(ctx, profile.CompanyID)
		results <- result{financialData: fd, err: err}
	}()

	// Get financial ratios
	go func() {
		fr, err := tr.GetFinancialRatios(ctx, profile.CompanyID)
		results <- result{financialRatios: fr, err: err}
	}()

	// Get risk metrics
	go func() {
		rm, err := tr.GetRiskMetrics(ctx, profile.CompanyID)
		results <- result{riskMetrics: rm, err: err}
	}()

	// Get ESG score
	go func() {
		esg, err := tr.GetESGScore(ctx, profile.CompanyID)
		results <- result{esgScore: esg, err: err}
	}()

	// Get executive info
	go func() {
		ei, err := tr.GetExecutiveInfo(ctx, profile.CompanyID)
		results <- result{executiveInfo: ei, err: err}
	}()

	// Get ownership structure
	go func() {
		os, err := tr.GetOwnershipStructure(ctx, profile.CompanyID)
		results <- result{ownershipStructure: os, err: err}
	}()

	// Collect results
	var comprehensiveResult result
	for i := 0; i < 6; i++ {
		r := <-results
		if r.err != nil {
			tr.logger.Warn("Failed to get some Thomson Reuters data",
				zap.Error(r.err))
		}
		if r.financialData != nil {
			comprehensiveResult.financialData = r.financialData
		}
		if r.financialRatios != nil {
			comprehensiveResult.financialRatios = r.financialRatios
		}
		if r.riskMetrics != nil {
			comprehensiveResult.riskMetrics = r.riskMetrics
		}
		if r.esgScore != nil {
			comprehensiveResult.esgScore = r.esgScore
		}
		if r.executiveInfo != nil {
			comprehensiveResult.executiveInfo = r.executiveInfo
		}
		if r.ownershipStructure != nil {
			comprehensiveResult.ownershipStructure = r.ownershipStructure
		}
	}

	// Create comprehensive result
	trResult := &ThomsonReutersResult{
		CompanyProfile:     profile,
		FinancialData:      comprehensiveResult.financialData,
		FinancialRatios:    comprehensiveResult.financialRatios,
		RiskMetrics:        comprehensiveResult.riskMetrics,
		ESGScore:           comprehensiveResult.esgScore,
		ExecutiveInfo:      comprehensiveResult.executiveInfo,
		OwnershipStructure: comprehensiveResult.ownershipStructure,
		DataQuality:        tr.generateDataQuality(),
		LastChecked:        time.Now(),
		ProcessingTime:     time.Since(startTime),
	}

	tr.logger.Info("Comprehensive Thomson Reuters data retrieved (mock)",
		zap.String("business_name", businessName),
		zap.Duration("processing_time", trResult.ProcessingTime),
		zap.String("data_quality", trResult.DataQuality))

	return trResult, nil
}

// GenerateRiskFactors generates risk factors from Thomson Reuters data
func (tr *ThomsonReutersMock) GenerateRiskFactors(result *ThomsonReutersResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Financial risk factors
	if result.FinancialData != nil {
		// Revenue growth risk
		revenueGrowthRisk := 0.3
		if result.FinancialData.Revenue < 1000000 {
			revenueGrowthRisk = 0.7
		} else if result.FinancialData.Revenue > 10000000 {
			revenueGrowthRisk = 0.2
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryFinancial,
			Subcategory: "revenue",
			Name:        "revenue_growth_risk",
			Score:       revenueGrowthRisk,
			Weight:      0.3,
			Description: "Risk associated with revenue growth and financial performance",
			Source:      "thomson_reuters",
			Confidence:  0.85,
			Impact:      "Revenue growth impacts business sustainability",
			Mitigation:  "Monitor revenue trends and implement growth strategies",
			LastUpdated: &now,
		})

		// Profitability risk
		profitabilityRisk := 0.4
		if result.FinancialData.NetIncome > 0 {
			profitabilityRisk = 0.2
		} else {
			profitabilityRisk = 0.8
		}

		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryFinancial,
			Subcategory: "profitability",
			Name:        "profitability_risk",
			Score:       profitabilityRisk,
			Weight:      0.25,
			Description: "Risk associated with profitability and financial health",
			Source:      "thomson_reuters",
			Confidence:  0.90,
			Impact:      "Profitability affects business viability",
			Mitigation:  "Improve operational efficiency and cost management",
			LastUpdated: &now,
		})
	}

	// Risk metrics factors
	if result.RiskMetrics != nil {
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryFinancial,
			Subcategory: "overall_risk",
			Name:        "thomson_reuters_risk_score",
			Score:       result.RiskMetrics.OverallRiskScore,
			Weight:      0.35,
			Description: "Overall risk score from Thomson Reuters analysis",
			Source:      "thomson_reuters",
			Confidence:  0.88,
			Impact:      "Comprehensive risk assessment from financial data provider",
			Mitigation:  "Address specific risk areas identified in analysis",
			LastUpdated: &now,
		})
	}

	// ESG risk factors
	if result.ESGScore != nil {
		esgRisk := 1.0 - (result.ESGScore.OverallESGScore / 100.0)
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryEnvironmental,
			Subcategory: "esg",
			Name:        "esg_risk",
			Score:       esgRisk,
			Weight:      0.2,
			Description: "Environmental, Social, and Governance risk assessment",
			Source:      "thomson_reuters",
			Confidence:  0.82,
			Impact:      "ESG factors affect reputation and regulatory compliance",
			Mitigation:  "Implement ESG best practices and monitoring",
			LastUpdated: &now,
		})
	}

	return riskFactors
}

// Helper methods for generating mock data

func (tr *ThomsonReutersMock) generateCompanyID(businessName string) string {
	return fmt.Sprintf("TR_%s_%d", tr.sanitizeForURL(businessName), time.Now().Unix())
}

func (tr *ThomsonReutersMock) detectIndustry(businessName string) string {
	name := strings.ToLower(businessName)
	switch {
	case strings.Contains(name, "bank") || strings.Contains(name, "financial"):
		return "financial_services"
	case strings.Contains(name, "tech") || strings.Contains(name, "software"):
		return "technology"
	case strings.Contains(name, "health") || strings.Contains(name, "medical"):
		return "healthcare"
	case strings.Contains(name, "retail") || strings.Contains(name, "store"):
		return "retail"
	case strings.Contains(name, "manufacturing") || strings.Contains(name, "production"):
		return "manufacturing"
	default:
		return "general"
	}
}

func (tr *ThomsonReutersMock) generateEmployeeCount() int {
	return rand.Intn(5000) + 10
}

func (tr *ThomsonReutersMock) generateRevenue() float64 {
	return rand.Float64()*50000000 + 100000 // $100k-$50M
}

func (tr *ThomsonReutersMock) generateMarketCap() float64 {
	return rand.Float64()*1000000000 + 1000000 // $1M-$1B
}

func (tr *ThomsonReutersMock) generateBusinessDescription(businessName string) string {
	descriptions := []string{
		fmt.Sprintf("%s is a leading company in its industry, providing innovative solutions to customers worldwide.", businessName),
		fmt.Sprintf("%s specializes in delivering high-quality products and services to meet customer needs.", businessName),
		fmt.Sprintf("%s is committed to excellence and innovation in all aspects of its business operations.", businessName),
		fmt.Sprintf("%s has established itself as a trusted partner in the industry with a strong track record.", businessName),
	}
	return descriptions[rand.Intn(len(descriptions))]
}

func (tr *ThomsonReutersMock) sanitizeForURL(name string) string {
	// Simple sanitization for URL generation
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", ""), "&", "and"))
}

func (tr *ThomsonReutersMock) generateAddress(country string) string {
	addresses := map[string][]string{
		"US": {"123 Main St, New York, NY 10001", "456 Business Ave, San Francisco, CA 94105", "789 Corporate Blvd, Chicago, IL 60601"},
		"UK": {"10 Downing Street, London SW1A 2AA", "25 Business Park, Manchester M1 1AA", "50 Commercial Road, Birmingham B1 1AA"},
		"CA": {"100 Bay Street, Toronto, ON M5H 2Y2", "2000 McGill College, Montreal, QC H3A 3H3", "300 Granville Street, Vancouver, BC V6C 1S4"},
	}

	if countryAddresses, exists := addresses[country]; exists {
		return countryAddresses[rand.Intn(len(countryAddresses))]
	}
	return "123 International Blvd, Global City, GC 12345"
}

func (tr *ThomsonReutersMock) generatePhone(country string) string {
	phoneFormats := map[string][]string{
		"US": {"+1-555-123-4567", "+1-212-555-0123", "+1-415-555-0456"},
		"UK": {"+44-20-7946-0958", "+44-161-555-0123", "+44-121-555-0456"},
		"CA": {"+1-416-555-0123", "+1-514-555-0456", "+1-604-555-0789"},
	}

	if countryPhones, exists := phoneFormats[country]; exists {
		return countryPhones[rand.Intn(len(countryPhones))]
	}
	return "+1-555-000-0000"
}

func (tr *ThomsonReutersMock) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}

func (tr *ThomsonReutersMock) generateRiskTrend() string {
	trends := []string{"improving", "stable", "deteriorating"}
	return trends[rand.Intn(len(trends))]
}

func (tr *ThomsonReutersMock) calculateESGLevel(score float64) string {
	switch {
	case score >= 80:
		return "excellent"
	case score >= 60:
		return "good"
	case score >= 40:
		return "average"
	default:
		return "poor"
	}
}

func (tr *ThomsonReutersMock) generateExecutiveName() string {
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Emily", "James", "Jessica"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func (tr *ThomsonReutersMock) generateExperience() string {
	experiences := []string{
		"20+ years in executive leadership",
		"15+ years in financial management",
		"10+ years in technology leadership",
		"12+ years in operations management",
		"18+ years in strategic planning",
	}
	return experiences[rand.Intn(len(experiences))]
}

func (tr *ThomsonReutersMock) generateEducation() string {
	educations := []string{
		"MBA from Harvard Business School",
		"Master's in Finance from Wharton",
		"Bachelor's in Engineering from MIT",
		"PhD in Economics from Stanford",
		"Master's in Business Administration",
	}
	return educations[rand.Intn(len(educations))]
}

func (tr *ThomsonReutersMock) generateOwnerName() string {
	ownerNames := []string{
		"BlackRock Inc.",
		"Vanguard Group",
		"State Street Corporation",
		"Fidelity Investments",
		"John Smith",
		"Jane Doe",
		"Michael Johnson",
		"Sarah Williams",
	}
	return ownerNames[rand.Intn(len(ownerNames))]
}

// World-Check Integration Methods

// GetWorldCheckScreening performs World-Check screening if enabled
func (tr *ThomsonReutersMock) GetWorldCheckScreening(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	if tr.worldCheck == nil {
		return nil, fmt.Errorf("World-Check client not enabled")
	}

	tr.logger.Info("Performing World-Check screening via Thomson Reuters (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	return tr.worldCheck.GetComprehensiveScreening(ctx, entityName, country)
}

// GetWorldCheckPEP performs PEP screening if enabled
func (tr *ThomsonReutersMock) GetWorldCheckPEP(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	if tr.worldCheck == nil {
		return nil, fmt.Errorf("World-Check client not enabled")
	}

	tr.logger.Info("Performing World-Check PEP screening via Thomson Reuters (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	return tr.worldCheck.ScreenPEP(ctx, entityName, country)
}

// GetWorldCheckSanctions performs sanctions screening if enabled
func (tr *ThomsonReutersMock) GetWorldCheckSanctions(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	if tr.worldCheck == nil {
		return nil, fmt.Errorf("World-Check client not enabled")
	}

	tr.logger.Info("Performing World-Check sanctions screening via Thomson Reuters (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	return tr.worldCheck.ScreenSanctions(ctx, entityName, country)
}

// GetWorldCheckAdverseMedia performs adverse media screening if enabled
func (tr *ThomsonReutersMock) GetWorldCheckAdverseMedia(ctx context.Context, entityName, country string) (*WorldCheckScreeningResult, error) {
	if tr.worldCheck == nil {
		return nil, fmt.Errorf("World-Check client not enabled")
	}

	tr.logger.Info("Performing World-Check adverse media screening via Thomson Reuters (mock)",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	return tr.worldCheck.ScreenAdverseMedia(ctx, entityName, country)
}

// GetEnhancedComprehensiveData retrieves all available data including World-Check screening
func (tr *ThomsonReutersMock) GetEnhancedComprehensiveData(ctx context.Context, businessName, country string) (*ThomsonReutersResult, error) {
	startTime := time.Now()
	tr.logger.Info("Getting enhanced comprehensive Thomson Reuters data with World-Check (mock)",
		zap.String("business_name", businessName),
		zap.String("country", country))

	// Get standard comprehensive data
	result, err := tr.GetComprehensiveData(ctx, businessName, country)
	if err != nil {
		return nil, fmt.Errorf("failed to get comprehensive data: %w", err)
	}

	// Add World-Check screening if enabled
	if tr.worldCheck != nil {
		worldCheckResult, err := tr.worldCheck.GetComprehensiveScreening(ctx, businessName, country)
		if err != nil {
			tr.logger.Warn("Failed to get World-Check screening data",
				zap.Error(err))
		} else {
			// Add World-Check data to metadata
			if result.Metadata == nil {
				result.Metadata = make(map[string]interface{})
			}
			result.Metadata["worldcheck_screening"] = worldCheckResult
			result.Metadata["worldcheck_risk_level"] = worldCheckResult.OverallRiskLevel
			result.Metadata["worldcheck_risk_score"] = worldCheckResult.OverallRiskScore
			result.Metadata["worldcheck_matches"] = worldCheckResult.TotalMatches
		}
	}

	result.ProcessingTime = time.Since(startTime)

	tr.logger.Info("Enhanced comprehensive Thomson Reuters data retrieved (mock)",
		zap.String("business_name", businessName),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.String("data_quality", result.DataQuality))

	return result, nil
}

// GenerateEnhancedRiskFactors generates risk factors including World-Check data
func (tr *ThomsonReutersMock) GenerateEnhancedRiskFactors(result *ThomsonReutersResult) []models.RiskFactor {
	// Get standard risk factors
	riskFactors := tr.GenerateRiskFactors(result)

	// Add World-Check risk factors if available
	if result.Metadata != nil {
		if worldCheckData, exists := result.Metadata["worldcheck_screening"]; exists {
			if worldCheckResult, ok := worldCheckData.(*WorldCheckScreeningResult); ok {
				worldCheckFactors := tr.worldCheck.GenerateRiskFactors(worldCheckResult)
				riskFactors = append(riskFactors, worldCheckFactors...)
			}
		}
	}

	return riskFactors
}

// IsHealthy checks if the Thomson Reuters service is healthy
func (tr *ThomsonReutersMock) IsHealthy(ctx context.Context) error {
	tr.logger.Info("Checking Thomson Reuters service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}
