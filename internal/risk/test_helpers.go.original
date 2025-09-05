package risk

import (
	"context"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockRiskService provides a simplified risk service for testing
type MockRiskService struct {
	logger *observability.Logger
}

// NewMockRiskService creates a new mock risk service for testing
func NewMockRiskService(logger *observability.Logger) *MockRiskService {
	return &MockRiskService{
		logger: logger,
	}
}

// AssessRisk provides a mock implementation of risk assessment
func (m *MockRiskService) AssessRisk(ctx context.Context, request RiskAssessmentRequest) (*RiskAssessmentResponse, error) {
	// Create mock factor scores
	factorScores := []RiskScore{
		{
			FactorID:     "financial-1",
			FactorName:   "Cash Flow",
			Category:     RiskCategoryFinancial,
			Score:        80.0,
			Level:        RiskLevelMedium,
			Confidence:   0.85,
			Explanation:  "Cash flow analysis shows moderate risk",
			Evidence:     []string{"recent financial statements", "payment history"},
			CalculatedAt: time.Now(),
		},
		{
			FactorID:     "operational-1",
			FactorName:   "Operational Efficiency",
			Category:     RiskCategoryOperational,
			Score:        70.0,
			Level:        RiskLevelMedium,
			Confidence:   0.80,
			Explanation:  "Operational processes show some inefficiencies",
			Evidence:     []string{"process analysis", "performance metrics"},
			CalculatedAt: time.Now(),
		},
	}

	// Create mock category scores
	categoryScores := make(map[RiskCategory]RiskScore)
	for _, category := range request.Categories {
		categoryScores[category] = RiskScore{
			FactorID:     string(category) + "-category",
			FactorName:   string(category) + " Risk",
			Category:     category,
			Score:        75.0,
			Level:        RiskLevelMedium,
			Confidence:   0.82,
			Explanation:  "Moderate risk level for " + string(category),
			Evidence:     []string{"risk analysis", "industry benchmarks"},
			CalculatedAt: time.Now(),
		}
	}

	// Create mock recommendations
	recommendations := []RiskRecommendation{
		{
			ID:          "rec-1",
			RiskFactor:  "financial",
			Title:       "Monitor Cash Flow",
			Description: "Regular monitoring of cash flow to identify potential issues early",
			Priority:    RiskLevelMedium,
			Action:      "Implement weekly cash flow monitoring",
			Impact:      "Reduces financial risk by 15%",
			Timeline:    "2 weeks",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "rec-2",
			RiskFactor:  "operational",
			Title:       "Improve Operational Efficiency",
			Description: "Streamline operational processes to reduce inefficiencies",
			Priority:    RiskLevelMedium,
			Action:      "Conduct process optimization review",
			Impact:      "Improves operational efficiency by 20%",
			Timeline:    "1 month",
			CreatedAt:   time.Now(),
		},
	}

	// Create a mock assessment
	assessment := &RiskAssessment{
		ID:              "mock-assessment-123",
		BusinessID:      request.BusinessID,
		BusinessName:    request.BusinessName,
		OverallScore:    75.5,
		OverallLevel:    RiskLevelMedium,
		CategoryScores:  categoryScores,
		FactorScores:    factorScores,
		Recommendations: recommendations,
		Alerts:          []RiskAlert{},
		AlertLevel:      RiskLevelMedium,
		AssessedAt:      time.Now(),
		ValidUntil:      time.Now().AddDate(0, 1, 0),
		Metadata:        map[string]interface{}{"test": true},
	}

	// Create mock predictions if requested
	var predictions []RiskPrediction
	if request.IncludePredictions {
		predictions = []RiskPrediction{
			{
				ID:             "pred-1",
				BusinessID:     request.BusinessID,
				FactorID:       "financial-1",
				PredictedScore: 70.0,
				PredictedLevel: RiskLevelMedium,
				Confidence:     0.85,
				Horizon:        "6 months",
				PredictedAt:    time.Now(),
				Factors:        []string{"market conditions", "business growth"},
			},
		}
	}

	return &RiskAssessmentResponse{
		Assessment:  assessment,
		Trends:      []RiskTrend{},
		Predictions: predictions,
		Alerts:      []RiskAlert{},
		GeneratedAt: time.Now(),
	}, nil
}

// GetCategoryRegistry returns a mock category registry
func (m *MockRiskService) GetCategoryRegistry() *RiskCategoryRegistry {
	return NewRiskCategoryRegistry()
}

// GetThresholdManager returns a mock threshold manager
func (m *MockRiskService) GetThresholdManager() *ThresholdManager {
	return NewThresholdManager()
}

// GenerateRiskReport provides a mock implementation
func (m *MockRiskService) GenerateRiskReport(ctx context.Context, request ReportRequest) (*RiskReport, error) {
	return &RiskReport{
		ID:           "report-123",
		BusinessID:   request.BusinessID,
		BusinessName: "Mock Business",
		ReportType:   request.ReportType,
		Format:       request.Format,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 1, 0),
	}, nil
}

// ExportRiskData provides a mock implementation
func (m *MockRiskService) ExportRiskData(ctx context.Context, request ExportRequest) (*ExportResponse, error) {
	return &ExportResponse{
		ExportID:    "export-123",
		BusinessID:  request.BusinessID,
		ExportType:  request.ExportType,
		Format:      request.Format,
		Data:        "mock data",
		RecordCount: 100,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 1, 0),
	}, nil
}

// CreateExportJob provides a mock implementation
func (m *MockRiskService) CreateExportJob(ctx context.Context, request ExportRequest) (*ExportJob, error) {
	return &ExportJob{
		ID:         "job-123",
		BusinessID: request.BusinessID,
		ExportType: request.ExportType,
		Format:     request.Format,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now(),
	}, nil
}

// GetExportJob provides a mock implementation
func (m *MockRiskService) GetExportJob(ctx context.Context, jobID string) (*ExportJob, error) {
	return &ExportJob{
		ID:          jobID,
		BusinessID:  "mock-business",
		ExportType:  "assessments",
		Format:      "json",
		Status:      "completed",
		Progress:    100,
		CreatedAt:   time.Now(),
		CompletedAt: &time.Time{},
	}, nil
}

// Mock implementations for other methods that might be called in tests
func (m *MockRiskService) GetCompanyFinancials(ctx context.Context, businessID string) (*FinancialData, error) {
	return &FinancialData{
		BusinessID:  businessID,
		Provider:    "mock",
		LastUpdated: time.Now(),
		Revenue: &RevenueData{
			TotalRevenue:           1000000,
			GrossRevenue:           1000000,
			NetRevenue:             950000,
			RevenueGrowth:          5.0,
			RevenueStability:       85.0,
			RevenueDiversification: 70.0,
			Currency:               "USD",
			Period:                 "yearly",
		},
		Profitability: &ProfitabilityData{
			GrossProfitMargin:  25.0,
			NetProfitMargin:    15.0,
			OperatingMargin:    20.0,
			EBITDAMargin:       22.0,
			ReturnOnAssets:     12.0,
			ReturnOnEquity:     18.0,
			ReturnOnInvestment: 15.0,
		},
		Assets: &AssetsData{
			TotalAssets:      2000000,
			CurrentAssets:    800000,
			FixedAssets:      1200000,
			IntangibleAssets: 100000,
			AssetUtilization: 75.0,
		},
		Liabilities: &LiabilitiesData{
			TotalLiabilities:      500000,
			CurrentLiabilities:    300000,
			LongTermLiabilities:   200000,
			ContingentLiabilities: 0,
		},
	}, nil
}

func (m *MockRiskService) GetCreditScore(ctx context.Context, businessID string) (*CreditScore, error) {
	return &CreditScore{
		BusinessID:  businessID,
		Provider:    "mock",
		Score:       750,
		ScoreRange:  "good",
		LastUpdated: time.Now(),
		Trend:       "stable",
		RiskLevel:   RiskLevelMedium,
	}, nil
}

func (m *MockRiskService) GetSanctionsData(ctx context.Context, businessID string) (*SanctionsData, error) {
	return &SanctionsData{
		BusinessID:     businessID,
		Provider:       "mock",
		LastUpdated:    time.Now(),
		HasSanctions:   false,
		SanctionsList:  []SanctionsMatch{},
		RiskLevel:      RiskLevelLow,
		Confidence:     0.95,
		ScreeningLists: []string{"OFAC", "UN", "EU"},
	}, nil
}

func (m *MockRiskService) GetLicenseData(ctx context.Context, businessID string) (*LicenseData, error) {
	return &LicenseData{
		BusinessID:    businessID,
		Provider:      "mock",
		LastUpdated:   time.Now(),
		Licenses:      []BusinessLicense{},
		OverallStatus: "active",
		RiskLevel:     RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	return &ComplianceData{
		BusinessID:           businessID,
		Provider:             "mock",
		LastUpdated:          time.Now(),
		OverallScore:         85.0,
		ComplianceFrameworks: []ComplianceFramework{},
		RiskLevel:            RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetRegulatoryViolations(ctx context.Context, businessID string) (*RegulatoryViolations, error) {
	return &RegulatoryViolations{
		BusinessID:         businessID,
		Provider:           "mock",
		LastUpdated:        time.Now(),
		TotalViolations:    0,
		ActiveViolations:   0,
		ResolvedViolations: 0,
		Violations:         []RegulatoryViolation{},
		TotalFines:         0.0,
		RiskLevel:          RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetTaxComplianceData(ctx context.Context, businessID string) (*TaxComplianceData, error) {
	return &TaxComplianceData{
		BusinessID:        businessID,
		Provider:          "mock",
		LastUpdated:       time.Now(),
		TaxID:             "123456789",
		TaxIDStatus:       "valid",
		TaxLienCount:      0,
		TaxLienAmount:     0.0,
		TaxLiens:          []TaxLien{},
		ComplianceHistory: []TaxComplianceEvent{},
		RiskLevel:         RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetDataProtectionCompliance(ctx context.Context, businessID string) (*DataProtectionCompliance, error) {
	return &DataProtectionCompliance{
		BusinessID:          businessID,
		Provider:            "mock",
		LastUpdated:         time.Now(),
		OverallScore:        85.0,
		Frameworks:          []DataProtectionFramework{},
		DataBreaches:        []DataBreach{},
		PrivacyPolicyStatus: "active",
		DataHandlingScore:   80.0,
		RiskLevel:           RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error) {
	return &NewsResult{
		BusinessID:       businessID,
		Provider:         "mock",
		LastUpdated:      time.Now(),
		TotalArticles:    0,
		PositiveCount:    0,
		NegativeCount:    0,
		NeutralCount:     0,
		Articles:         []NewsArticle{},
		RiskLevel:        RiskLevelLow,
		OverallSentiment: 0.5,
	}, nil
}

func (m *MockRiskService) GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error) {
	return &SocialMediaResult{
		BusinessID:       businessID,
		Provider:         "mock",
		LastUpdated:      time.Now(),
		TotalMentions:    0,
		PositiveCount:    0,
		NegativeCount:    0,
		NeutralCount:     0,
		Mentions:         []SocialMediaMention{},
		RiskLevel:        RiskLevelLow,
		OverallSentiment: 0.5,
	}, nil
}

func (m *MockRiskService) GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error) {
	return &SentimentResult{
		BusinessID:    businessID,
		Provider:      "mock",
		LastUpdated:   time.Now(),
		OverallScore:  0.5,
		PositiveScore: 0.3,
		NegativeScore: 0.2,
		NeutralScore:  0.5,
		Confidence:    0.8,
		RiskLevel:     RiskLevelLow,
		Trend:         "stable",
	}, nil
}

func (m *MockRiskService) GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error) {
	return &ReputationScore{
		BusinessID:   businessID,
		Provider:     "mock",
		LastUpdated:  time.Now(),
		OverallScore: 75.0,
		NewsScore:    70.0,
		SocialScore:  80.0,
		ReviewScore:  75.0,
		RiskLevel:    RiskLevelLow,
		Trend:        "stable",
		Confidence:   0.8,
	}, nil
}

func (m *MockRiskService) GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error) {
	return &MediaAlerts{
		BusinessID:     businessID,
		Provider:       "mock",
		LastUpdated:    time.Now(),
		TotalAlerts:    0,
		HighPriority:   0,
		MediumPriority: 0,
		LowPriority:    0,
		Alerts:         []MediaAlert{},
		RiskLevel:      RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetEconomicIndicators(ctx context.Context, country string) (*EconomicIndicators, error) {
	return &EconomicIndicators{
		Country:     country,
		Provider:    "mock",
		LastUpdated: time.Now(),
		GDP: &GDPData{
			CurrentGDP:   2000000000000,
			GDPGrowth:    2.5,
			GDPPerCapita: 60000,
			GDPForecast:  2.8,
			LastUpdated:  time.Now(),
		},
		Inflation: &InflationData{
			CurrentInflation:  2.5,
			CoreInflation:     2.2,
			InflationTrend:    "stable",
			InflationForecast: 2.3,
			LastUpdated:       time.Now(),
		},
		Unemployment: &UnemploymentData{
			UnemploymentRate:   5.0,
			EmploymentGrowth:   1.5,
			LaborParticipation: 62.5,
			LastUpdated:        time.Now(),
		},
		RiskLevel: RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetMarketIndustryBenchmarks(ctx context.Context, industry string, region string) (*MarketIndustryBenchmarks, error) {
	return &MarketIndustryBenchmarks{
		Industry:    industry,
		Region:      region,
		Provider:    "mock",
		LastUpdated: time.Now(),
		RevenueMetrics: &RevenueMetrics{
			MedianRevenue:    1000000,
			AverageRevenue:   1200000,
			RevenueGrowth:    5.0,
			RevenueStability: 85.0,
			RevenueVariance:  15.0,
		},
		ProfitabilityMetrics: &ProfitabilityMetrics{
			MedianGrossMargin: 25.0,
			MedianNetMargin:   15.0,
			MedianROA:         12.0,
			MedianROE:         18.0,
			EBITDAMargin:      22.0,
		},
		RiskLevel: RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetMarketRiskFactors(ctx context.Context, sector string) (*MarketRiskFactors, error) {
	return &MarketRiskFactors{
		Sector:      sector,
		Provider:    "mock",
		LastUpdated: time.Now(),
		MarketVolatility: &MarketVolatility{
			VIXIndex:         20.0,
			SectorVolatility: 15.0,
			BetaCoefficient:  1.0,
			VolatilityTrend:  "stable",
		},
		RiskLevel: RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetCommodityPrices(ctx context.Context, commodities []string) (*CommodityPrices, error) {
	commodityPrices := make([]CommodityPrice, len(commodities))
	for i, commodity := range commodities {
		commodityPrices[i] = CommodityPrice{
			CommodityName:      commodity,
			CurrentPrice:       100.0,
			PriceChange:        2.0,
			PriceChangePercent: 2.0,
			Currency:           "USD",
			LastUpdated:        time.Now(),
		}
	}

	return &CommodityPrices{
		Provider:    "mock",
		LastUpdated: time.Now(),
		Commodities: commodityPrices,
		RiskLevel:   RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetCurrencyRates(ctx context.Context, baseCurrency string) (*CurrencyRates, error) {
	return &CurrencyRates{
		BaseCurrency: baseCurrency,
		Provider:     "mock",
		LastUpdated:  time.Now(),
		Rates:        []CurrencyRate{},
		RiskLevel:    RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetMarketTrends(ctx context.Context, market string) (*MarketTrends, error) {
	return &MarketTrends{
		Market:      market,
		Provider:    "mock",
		LastUpdated: time.Now(),
		Trends:      []MarketTrend{},
		RiskLevel:   RiskLevelLow,
	}, nil
}

func (m *MockRiskService) ValidateFinancialData(ctx context.Context, data *FinancialData) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "financial",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.95,
		QualityScore:      0.9,
		CompletenessScore: 0.95,
		ReliabilityScore:  0.9,
		ConsistencyScore:  0.95,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) ValidateRegulatoryData(ctx context.Context, data *RegulatoryViolations) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "regulatory",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.9,
		QualityScore:      0.85,
		CompletenessScore: 0.9,
		ReliabilityScore:  0.9,
		ConsistencyScore:  0.9,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) ValidateMediaData(ctx context.Context, data *NewsResult) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "media",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.85,
		QualityScore:      0.8,
		CompletenessScore: 0.85,
		ReliabilityScore:  0.8,
		ConsistencyScore:  0.85,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) ValidateMarketData(ctx context.Context, data *EconomicIndicators) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "market",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.92,
		QualityScore:      0.9,
		CompletenessScore: 0.92,
		ReliabilityScore:  0.9,
		ConsistencyScore:  0.92,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "risk_assessment",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.88,
		QualityScore:      0.85,
		CompletenessScore: 0.88,
		ReliabilityScore:  0.85,
		ConsistencyScore:  0.88,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) ValidateRiskFactor(ctx context.Context, factor *RiskFactorResult) (*ValidationResult, error) {
	return &ValidationResult{
		DataID:            "mock-validation",
		DataType:          "risk_factor",
		Provider:          "mock",
		ValidatedAt:       time.Now(),
		OverallScore:      0.87,
		QualityScore:      0.85,
		CompletenessScore: 0.87,
		ReliabilityScore:  0.85,
		ConsistencyScore:  0.87,
		IsValid:           true,
		Warnings:          []ValidationWarning{},
		Errors:            []ValidationError{},
		Recommendations:   []ValidationRecommendation{},
	}, nil
}

func (m *MockRiskService) MonitorThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	return []ThresholdAlert{}, nil
}

func (m *MockRiskService) GetThresholdConfig(ctx context.Context, category RiskCategory) (*ThresholdMonitoringConfig, error) {
	return &ThresholdMonitoringConfig{
		Category:             category,
		WarningThreshold:     70.0,
		CriticalThreshold:    90.0,
		ApproachingThreshold: 60.0,
		TrendingThreshold:    5.0,
		VolatilityThreshold:  10.0,
		AnomalyThreshold:     15.0,
		ImprovementThreshold: -5.0,
		Enabled:              true,
		AlertChannels:        []string{"email", "dashboard"},
		NotificationRules:    []NotificationRule{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}

func (m *MockRiskService) UpdateThresholdConfig(ctx context.Context, category RiskCategory, config *ThresholdMonitoringConfig) error {
	return nil
}

func (m *MockRiskService) GetMonitoringStatus(ctx context.Context) (*MonitoringStatus, error) {
	return &MonitoringStatus{
		ActiveMonitors:   5,
		TotalAlerts:      0,
		CriticalAlerts:   0,
		WarningAlerts:    0,
		LastAlertTime:    nil,
		MonitoringHealth: "healthy",
		Uptime:           time.Hour * 24,
		Metadata:         map[string]interface{}{},
	}, nil
}

func (m *MockRiskService) ProcessAutomatedAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	return []AutomatedAlert{}, nil
}

func (m *MockRiskService) GetAutomatedAlertRules(ctx context.Context) ([]*AutomatedAlertRule, error) {
	return []*AutomatedAlertRule{}, nil
}

func (m *MockRiskService) CreateAutomatedAlertRule(ctx context.Context, rule *AutomatedAlertRule) error {
	return nil
}

func (m *MockRiskService) UpdateAutomatedAlertRule(ctx context.Context, ruleID string, rule *AutomatedAlertRule) error {
	return nil
}

func (m *MockRiskService) DeleteAutomatedAlertRule(ctx context.Context, ruleID string) error {
	return nil
}

func (m *MockRiskService) GetAutomatedAlertHistory(ctx context.Context, businessID string) ([]AutomatedAlert, error) {
	return []AutomatedAlert{}, nil
}

func (m *MockRiskService) RegisterNotificationProvider(ctx context.Context, channel string, provider NotificationProvider) error {
	return nil
}

func (m *MockRiskService) GetPaymentHistory(ctx context.Context, businessID string) (*PaymentHistory, error) {
	return &PaymentHistory{
		BusinessID:        businessID,
		Provider:          "mock",
		TotalPayments:     100,
		OnTimePayments:    95,
		LatePayments:      3,
		DefaultedPayments: 2,
		PaymentRate:       95.0,
		AverageDaysLate:   5.0,
		LastPaymentDate:   time.Now().AddDate(0, 0, -5),
		PaymentTrend:      "improving",
		RiskLevel:         RiskLevelLow,
	}, nil
}

func (m *MockRiskService) GetIndustryBenchmarks(ctx context.Context, industry string) (*IndustryBenchmarks, error) {
	return &IndustryBenchmarks{
		Industry:    industry,
		Provider:    "mock",
		LastUpdated: time.Now(),
		RevenueBenchmarks: &RevenueBenchmarks{
			MedianRevenue:    1000000,
			AverageRevenue:   1200000,
			RevenueGrowth:    5.0,
			RevenueStability: 85.0,
		},
		ProfitabilityBenchmarks: &ProfitabilityBenchmarks{
			MedianGrossMargin: 25.0,
			MedianNetMargin:   15.0,
			MedianROA:         12.0,
			MedianROE:         18.0,
		},
		LiquidityBenchmarks: &LiquidityBenchmarks{
			MedianCurrentRatio: 2.0,
			MedianQuickRatio:   1.5,
			MedianCashRatio:    0.5,
		},
		SolvencyBenchmarks: &SolvencyBenchmarks{
			MedianDebtToEquity:     0.5,
			MedianDebtToAssets:     0.3,
			MedianInterestCoverage: 5.0,
		},
		Metadata: map[string]interface{}{},
	}, nil
}

func (m *MockRiskService) AnalyzeRiskTrends(ctx context.Context, businessID string, period string) (*TrendAnalysisResult, error) {
	return &TrendAnalysisResult{
		BusinessID:           businessID,
		AnalysisPeriod:       period,
		OverallTrend:         TrendDirectionStable,
		OverallTrendStrength: 0.5,
		CategoryTrends:       map[string]TrendData{},
		FactorTrends:         map[string]TrendData{},
		Seasonality: SeasonalityAnalysis{
			HasSeasonality: false,
			Pattern:        "",
			Strength:       0.0,
			PeakPeriods:    []time.Time{},
			TroughPeriods:  []time.Time{},
			SeasonalData:   map[string]float64{},
		},
		Volatility: VolatilityAnalysis{
			OverallVolatility:     0.1,
			VolatilityTrend:       TrendDirectionStable,
			HighVolatilityPeriods: []time.Time{},
			LowVolatilityPeriods:  []time.Time{},
			VolatilityByCategory:  map[string]float64{},
		},
		Predictions:     []TrendPrediction{},
		Anomalies:       []TrendAnomaly{},
		Recommendations: []TrendRecommendation{},
		GeneratedAt:     time.Now(),
	}, nil
}

func (m *MockRiskService) GetTrendPredictions(ctx context.Context, businessID string, horizon time.Duration) ([]TrendPrediction, error) {
	return []TrendPrediction{}, nil
}

func (m *MockRiskService) GetTrendAnomalies(ctx context.Context, businessID string) ([]TrendAnomaly, error) {
	return []TrendAnomaly{}, nil
}

func (m *MockRiskService) GenerateAdvancedReport(ctx context.Context, request AdvancedReportRequest) (*AdvancedRiskReport, error) {
	return &AdvancedRiskReport{
		ID:              "advanced-report-123",
		BusinessID:      request.BusinessID,
		BusinessName:    "Mock Business",
		ReportType:      request.ReportType,
		Format:          request.Format,
		GeneratedAt:     time.Now(),
		ValidUntil:      time.Now().AddDate(0, 1, 0),
		Summary:         &AdvancedReportSummary{},
		Analytics:       &ReportAnalytics{},
		Trends:          &AdvancedReportTrends{},
		Alerts:          []RiskAlert{},
		Recommendations: []RiskRecommendation{},
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Metadata:        map[string]interface{}{},
	}, nil
}
