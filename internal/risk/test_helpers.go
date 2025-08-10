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

	// Create mock category scores
	mockCategoryScores := make(map[RiskCategory]RiskScore)
	for _, category := range request.Categories {
		mockCategoryScores[category] = RiskScore{
			FactorID:     "factor-" + string(category),
			FactorName:   string(category) + " risk",
			Category:     category,
			Score:        75.0,
			Level:        "medium",
			Confidence:   0.85,
			Explanation:  "Mock risk score for " + string(category),
			Evidence:     []string{"factor1", "factor2"},
			CalculatedAt: time.Now(),
		}
	}

	// Create mock factor scores
	mockFactorScores := []RiskScore{
		{
			FactorID:     "financial-001",
			FactorName:   "Financial Risk",
			Category:     "financial",
			Score:        80.0,
			Level:        "medium",
			Confidence:   0.90,
			Explanation:  "Mock financial risk score",
			Evidence:     []string{"cash_flow", "debt_ratio"},
			CalculatedAt: time.Now(),
		},
		{
			FactorID:     "operational-001",
			FactorName:   "Operational Risk",
			Category:     "operational",
			Score:        70.0,
			Level:        "medium",
			Confidence:   0.85,
			Explanation:  "Mock operational risk score",
			Evidence:     []string{"efficiency", "processes"},
			CalculatedAt: time.Now(),
		},
	}

	// Create the assessment
	assessment := &RiskAssessment{
		ID:             "assessment-123",
		BusinessID:     request.BusinessID,
		BusinessName:   request.BusinessName,
		OverallScore:   75.0,
		OverallLevel:   "medium",
		CategoryScores: mockCategoryScores,
		FactorScores:   mockFactorScores,
		Recommendations: []RiskRecommendation{
			{
				ID:          "rec-001",
				RiskFactor:  "financial",
				Title:       "Improve Financial Management",
				Description: "Consider implementing better financial controls",
				Priority:    "medium",
				Action:      "Review financial processes",
				Impact:      "Reduce financial risk",
				Timeline:    "3 months",
				CreatedAt:   time.Now(),
			},
		},
		Alerts:     []RiskAlert{},
		AlertLevel: "medium",
		AssessedAt: time.Now(),
		ValidUntil: time.Now().AddDate(0, 1, 0),
		Metadata:   request.Metadata,
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
		ID:          "report-123",
		BusinessID:  request.BusinessID,
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
		ID:        "job-123",
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
		ID:        jobID,
		BusinessID: "mock-business",
		ExportType: "assessments",
		Format:     "json",
		Status:     "completed",
		Progress:   100,
		CreatedAt:  time.Now(),
		CompletedAt: &time.Time{},
	}, nil
}

// Mock implementations for other methods that might be called in tests
func (m *MockRiskService) GetCompanyFinancials(ctx context.Context, businessID string) (*FinancialData, error) {
	return &FinancialData{
		BusinessID:  businessID,
		Revenue:     1000000,
		Profit:      100000,
		Assets:      2000000,
		Liabilities: 500000,
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockRiskService) GetCreditScore(ctx context.Context, businessID string) (*CreditScore, error) {
	return &CreditScore{
		BusinessID: businessID,
		Score:      750,
		Provider:   "mock",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetSanctionsData(ctx context.Context, businessID string) (*SanctionsData, error) {
	return &SanctionsData{
		BusinessID:   businessID,
		IsSanctioned: false,
		Matches:      []string{},
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockRiskService) GetLicenseData(ctx context.Context, businessID string) (*LicenseData, error) {
	return &LicenseData{
		BusinessID: businessID,
		Licenses:   []string{"business_license"},
		Status:     "active",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	return &ComplianceData{
		BusinessID: businessID,
		Status:     "compliant",
		Score:      85.0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetRegulatoryViolations(ctx context.Context, businessID string) (*RegulatoryViolations, error) {
	return &RegulatoryViolations{
		BusinessID: businessID,
		Violations: []string{},
		Count:      0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetTaxComplianceData(ctx context.Context, businessID string) (*TaxComplianceData, error) {
	return &TaxComplianceData{
		BusinessID: businessID,
		Status:     "compliant",
		LastFiling: time.Now().AddDate(0, -1, 0),
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetDataProtectionCompliance(ctx context.Context, businessID string) (*DataProtectionCompliance, error) {
	return &DataProtectionCompliance{
		BusinessID: businessID,
		Status:     "compliant",
		Framework:  "GDPR",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error) {
	return &NewsResult{
		BusinessID: businessID,
		Articles:   []string{},
		Count:      0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error) {
	return &SocialMediaResult{
		BusinessID: businessID,
		Mentions:   []string{},
		Count:      0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error) {
	return &SentimentResult{
		BusinessID: businessID,
		Sentiment:  "neutral",
		Score:      0.5,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error) {
	return &ReputationScore{
		BusinessID: businessID,
		Score:      75.0,
		Provider:   "mock",
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error) {
	return &MediaAlerts{
		BusinessID: businessID,
		Alerts:     []string{},
		Count:      0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetEconomicIndicators(ctx context.Context, country string) (*EconomicIndicators, error) {
	return &EconomicIndicators{
		Country:      country,
		GDP:          2000000000000,
		Inflation:    2.5,
		Unemployment: 5.0,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockRiskService) GetMarketIndustryBenchmarks(ctx context.Context, industry string, region string) (*MarketIndustryBenchmarks, error) {
	return &MarketIndustryBenchmarks{
		Industry:   industry,
		Region:     region,
		Benchmarks: map[string]float64{"revenue_growth": 5.0},
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetMarketRiskFactors(ctx context.Context, sector string) (*MarketRiskFactors, error) {
	return &MarketRiskFactors{
		Sector:    sector,
		Factors:   []string{"market_volatility"},
		RiskLevel: "medium",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) GetCommodityPrices(ctx context.Context, commodities []string) (*CommodityPrices, error) {
	return &CommodityPrices{
		Commodities: commodities,
		Prices:      map[string]float64{"oil": 75.0},
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockRiskService) GetCurrencyRates(ctx context.Context, baseCurrency string) (*CurrencyRates, error) {
	return &CurrencyRates{
		BaseCurrency: baseCurrency,
		Rates:        map[string]float64{"USD": 1.0},
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockRiskService) GetMarketTrends(ctx context.Context, market string) (*MarketTrends, error) {
	return &MarketTrends{
		Market:    market,
		Trend:     "stable",
		Direction: "neutral",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateFinancialData(ctx context.Context, data *FinancialData) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     95.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateRegulatoryData(ctx context.Context, data *RegulatoryViolations) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     90.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateMediaData(ctx context.Context, data *NewsResult) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     85.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateMarketData(ctx context.Context, data *EconomicIndicators) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     92.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     88.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) ValidateRiskFactor(ctx context.Context, factor *RiskFactorResult) (*ValidationResult, error) {
	return &ValidationResult{
		IsValid:   true,
		Score:     87.0,
		Issues:    []string{},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskService) MonitorThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	return []ThresholdAlert{}, nil
}

func (m *MockRiskService) GetThresholdConfig(ctx context.Context, category RiskCategory) (*ThresholdMonitoringConfig, error) {
	return &ThresholdMonitoringConfig{
		Category:     category,
		MinThreshold: 0.0,
		MaxThreshold: 100.0,
		Enabled:      true,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockRiskService) UpdateThresholdConfig(ctx context.Context, category RiskCategory, config *ThresholdMonitoringConfig) error {
	return nil
}

func (m *MockRiskService) GetMonitoringStatus(ctx context.Context) (*MonitoringStatus, error) {
	return &MonitoringStatus{
		Active:    true,
		Alerts:    0,
		LastCheck: time.Now(),
		CreatedAt: time.Now(),
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
		BusinessID: businessID,
		Payments:   []Payment{},
		Count:      0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) GetIndustryBenchmarks(ctx context.Context, industry string) (*IndustryBenchmarks, error) {
	return &IndustryBenchmarks{
		Industry:   industry,
		Benchmarks: map[string]float64{"revenue_growth": 5.0},
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRiskService) AnalyzeRiskTrends(ctx context.Context, businessID string, period string) (*TrendAnalysisResult, error) {
	return &TrendAnalysisResult{
		BusinessID: businessID,
		Period:     period,
		Trend:      "stable",
		Direction:  "neutral",
		CreatedAt:  time.Now(),
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
		ID:         "advanced-report-123",
		BusinessID: request.BusinessID,
		Content:    "Mock advanced report content",
		CreatedAt:  time.Now(),
	}, nil
}
