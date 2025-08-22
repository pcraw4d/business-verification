package risk_assessment

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FinancialAnalyzer provides financial health analysis capabilities
type FinancialAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
}

// FinancialAnalysisResult contains comprehensive financial analysis results
type FinancialAnalysisResult struct {
	BusinessName     string            `json:"business_name"`
	FinancialData    *FinancialData    `json:"financial_data,omitempty"`
	RevenueAnalysis  *RevenueAnalysis  `json:"revenue_analysis,omitempty"`
	GrowthAnalysis   *GrowthAnalysis   `json:"growth_analysis,omitempty"`
	StabilityMetrics *StabilityMetrics `json:"stability_metrics,omitempty"`
	OverallScore     float64           `json:"overall_score"`
	RiskFactors      []RiskFactor      `json:"risk_factors"`
	Recommendations  []string          `json:"recommendations"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// FinancialData contains basic financial information
type FinancialData struct {
	AnnualRevenue     float64 `json:"annual_revenue"`
	ProfitMargin      float64 `json:"profit_margin"`
	CashFlow          float64 `json:"cash_flow"`
	TotalAssets       float64 `json:"total_assets"`
	TotalLiabilities  float64 `json:"total_liabilities"`
	WorkingCapital    float64 `json:"working_capital"`
	DebtToEquityRatio float64 `json:"debt_to_equity_ratio"`
	CurrentRatio      float64 `json:"current_ratio"`
	QuickRatio        float64 `json:"quick_ratio"`
	ROI               float64 `json:"roi"`
	ROE               float64 `json:"roe"`
	ROA               float64 `json:"roa"`
}

// RevenueAnalysis contains revenue-related analysis
type RevenueAnalysis struct {
	RevenueTrend     string  `json:"revenue_trend"`
	RevenueGrowth    float64 `json:"revenue_growth"`
	RevenueStability float64 `json:"revenue_stability"`
	RevenueDiversity float64 `json:"revenue_diversity"`
	SeasonalPattern  string  `json:"seasonal_pattern"`
	RevenueScore     float64 `json:"revenue_score"`
	RevenueRisk      string  `json:"revenue_risk"`
}

// GrowthAnalysis contains growth-related metrics
type GrowthAnalysis struct {
	RevenueGrowthRate    float64 `json:"revenue_growth_rate"`
	ProfitGrowthRate     float64 `json:"profit_growth_rate"`
	AssetGrowthRate      float64 `json:"asset_growth_rate"`
	EmployeeGrowthRate   float64 `json:"employee_growth_rate"`
	MarketShareGrowth    float64 `json:"market_share_growth"`
	GrowthScore          float64 `json:"growth_score"`
	GrowthSustainability string  `json:"growth_sustainability"`
}

// StabilityMetrics contains financial stability indicators
type StabilityMetrics struct {
	LiquidityScore     float64 `json:"liquidity_score"`
	SolvencyScore      float64 `json:"solvency_score"`
	ProfitabilityScore float64 `json:"profitability_score"`
	EfficiencyScore    float64 `json:"efficiency_score"`
	StabilityScore     float64 `json:"stability_score"`
	OverallStability   string  `json:"overall_stability"`
	RiskLevel          string  `json:"risk_level"`
}

// NewFinancialAnalyzer creates a new financial analyzer
func NewFinancialAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *FinancialAnalyzer {
	return &FinancialAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeFinancial performs comprehensive financial analysis
func (fa *FinancialAnalyzer) AnalyzeFinancial(ctx context.Context, businessName string, industry string) (*FinancialAnalysisResult, error) {
	fa.logger.Info("Starting financial analysis",
		zap.String("business", businessName),
		zap.String("industry", industry))

	result := &FinancialAnalysisResult{
		BusinessName: businessName,
		LastUpdated:  time.Now(),
	}

	// Analyze financial data if enabled
	if fa.config.FinancialDataEnabled {
		financialData, err := fa.analyzeFinancialData(ctx, businessName, industry)
		if err != nil {
			fa.logger.Warn("Financial data analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    RiskCategoryFinancial,
				Factor:      "financial_data_analysis",
				Description: fmt.Sprintf("Financial data analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Unable to assess financial health",
			})
		} else {
			result.FinancialData = financialData
		}
	}

	// Analyze revenue patterns if financial data is available
	if result.FinancialData != nil {
		revenueAnalysis, err := fa.analyzeRevenuePatterns(ctx, businessName, result.FinancialData)
		if err != nil {
			fa.logger.Warn("Revenue analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
		} else {
			result.RevenueAnalysis = revenueAnalysis
		}
	}

	// Analyze growth patterns if financial data is available
	if result.FinancialData != nil {
		growthAnalysis, err := fa.analyzeGrowthPatterns(ctx, businessName, result.FinancialData)
		if err != nil {
			fa.logger.Warn("Growth analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
		} else {
			result.GrowthAnalysis = growthAnalysis
		}
	}

	// Analyze stability metrics if financial data is available
	if result.FinancialData != nil {
		stabilityMetrics, err := fa.analyzeStabilityMetrics(ctx, businessName, result.FinancialData)
		if err != nil {
			fa.logger.Warn("Stability analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
		} else {
			result.StabilityMetrics = stabilityMetrics
		}
	}

	// Calculate overall score
	result.OverallScore = fa.calculateOverallScore(result)

	// Generate recommendations
	result.Recommendations = fa.generateRecommendations(result)

	fa.logger.Info("Financial analysis completed",
		zap.String("business", businessName),
		zap.Float64("score", result.OverallScore))

	return result, nil
}

// analyzeFinancialData analyzes basic financial metrics
func (fa *FinancialAnalyzer) analyzeFinancialData(ctx context.Context, businessName string, industry string) (*FinancialData, error) {
	fa.logger.Debug("Analyzing financial data",
		zap.String("business", businessName))

	// In a real implementation, this would query financial databases or APIs
	// For now, we'll simulate financial data based on industry
	financialData := &FinancialData{}

	// Simulate financial data based on industry type
	switch strings.ToLower(industry) {
	case "technology", "software":
		financialData.AnnualRevenue = 2500000.0
		financialData.ProfitMargin = 0.25
		financialData.CashFlow = 625000.0
		financialData.TotalAssets = 1800000.0
		financialData.TotalLiabilities = 800000.0
	case "retail", "ecommerce":
		financialData.AnnualRevenue = 1500000.0
		financialData.ProfitMargin = 0.12
		financialData.CashFlow = 180000.0
		financialData.TotalAssets = 1200000.0
		financialData.TotalLiabilities = 600000.0
	case "financial", "banking":
		financialData.AnnualRevenue = 5000000.0
		financialData.ProfitMargin = 0.18
		financialData.CashFlow = 900000.0
		financialData.TotalAssets = 8000000.0
		financialData.TotalLiabilities = 4000000.0
	default:
		financialData.AnnualRevenue = 1000000.0
		financialData.ProfitMargin = 0.15
		financialData.CashFlow = 150000.0
		financialData.TotalAssets = 1000000.0
		financialData.TotalLiabilities = 500000.0
	}

	// Calculate derived metrics
	financialData.WorkingCapital = financialData.TotalAssets - financialData.TotalLiabilities
	if financialData.TotalLiabilities > 0 {
		financialData.DebtToEquityRatio = financialData.TotalLiabilities / (financialData.TotalAssets - financialData.TotalLiabilities)
	}
	financialData.CurrentRatio = financialData.TotalAssets / financialData.TotalLiabilities
	financialData.QuickRatio = (financialData.TotalAssets - financialData.TotalAssets*0.3) / financialData.TotalLiabilities // Assuming 30% inventory
	financialData.ROI = financialData.ProfitMargin * (financialData.AnnualRevenue / financialData.TotalAssets)
	financialData.ROE = financialData.ProfitMargin * (financialData.AnnualRevenue / (financialData.TotalAssets - financialData.TotalLiabilities))
	financialData.ROA = financialData.ProfitMargin * (financialData.AnnualRevenue / financialData.TotalAssets)

	return financialData, nil
}

// analyzeRevenuePatterns analyzes revenue trends and patterns
func (fa *FinancialAnalyzer) analyzeRevenuePatterns(ctx context.Context, businessName string, financialData *FinancialData) (*RevenueAnalysis, error) {
	fa.logger.Debug("Analyzing revenue patterns",
		zap.String("business", businessName))

	// In a real implementation, this would analyze historical revenue data
	// For now, we'll simulate revenue analysis
	revenueAnalysis := &RevenueAnalysis{}

	// Simulate revenue trend based on profit margin and cash flow
	if financialData.ProfitMargin > 0.2 {
		revenueAnalysis.RevenueTrend = "increasing"
		revenueAnalysis.RevenueGrowth = 0.15
	} else if financialData.ProfitMargin > 0.1 {
		revenueAnalysis.RevenueTrend = "stable"
		revenueAnalysis.RevenueGrowth = 0.05
	} else {
		revenueAnalysis.RevenueTrend = "declining"
		revenueAnalysis.RevenueGrowth = -0.05
	}

	// Calculate revenue stability based on cash flow consistency
	cashFlowRatio := financialData.CashFlow / financialData.AnnualRevenue
	if cashFlowRatio > 0.2 {
		revenueAnalysis.RevenueStability = 0.9
	} else if cashFlowRatio > 0.1 {
		revenueAnalysis.RevenueStability = 0.7
	} else {
		revenueAnalysis.RevenueStability = 0.5
	}

	// Simulate revenue diversity (single vs multiple revenue streams)
	revenueAnalysis.RevenueDiversity = 0.75 // Assume moderate diversity

	// Determine seasonal pattern
	revenueAnalysis.SeasonalPattern = "moderate" // Assume moderate seasonality

	// Calculate revenue score
	revenueAnalysis.RevenueScore = fa.calculateRevenueScore(revenueAnalysis)

	// Determine revenue risk
	revenueAnalysis.RevenueRisk = fa.determineRevenueRisk(revenueAnalysis)

	return revenueAnalysis, nil
}

// analyzeGrowthPatterns analyzes growth metrics and sustainability
func (fa *FinancialAnalyzer) analyzeGrowthPatterns(ctx context.Context, businessName string, financialData *FinancialData) (*GrowthAnalysis, error) {
	fa.logger.Debug("Analyzing growth patterns",
		zap.String("business", businessName))

	// In a real implementation, this would analyze historical growth data
	// For now, we'll simulate growth analysis
	growthAnalysis := &GrowthAnalysis{}

	// Simulate growth rates based on financial health
	growthAnalysis.RevenueGrowthRate = financialData.ProfitMargin * 0.8 // Correlate with profitability
	growthAnalysis.ProfitGrowthRate = financialData.ProfitMargin * 0.6
	growthAnalysis.AssetGrowthRate = financialData.ProfitMargin * 0.4
	growthAnalysis.EmployeeGrowthRate = financialData.ProfitMargin * 0.3
	growthAnalysis.MarketShareGrowth = financialData.ProfitMargin * 0.2

	// Calculate growth score
	growthAnalysis.GrowthScore = fa.calculateGrowthScore(growthAnalysis)

	// Determine growth sustainability
	growthAnalysis.GrowthSustainability = fa.determineGrowthSustainability(growthAnalysis)

	return growthAnalysis, nil
}

// analyzeStabilityMetrics analyzes financial stability indicators
func (fa *FinancialAnalyzer) analyzeStabilityMetrics(ctx context.Context, businessName string, financialData *FinancialData) (*StabilityMetrics, error) {
	fa.logger.Debug("Analyzing stability metrics",
		zap.String("business", businessName))

	// In a real implementation, this would calculate stability metrics
	// For now, we'll simulate stability analysis
	stabilityMetrics := &StabilityMetrics{}

	// Calculate liquidity score
	stabilityMetrics.LiquidityScore = fa.calculateLiquidityScore(financialData)

	// Calculate solvency score
	stabilityMetrics.SolvencyScore = fa.calculateSolvencyScore(financialData)

	// Calculate profitability score
	stabilityMetrics.ProfitabilityScore = fa.calculateProfitabilityScore(financialData)

	// Calculate efficiency score
	stabilityMetrics.EfficiencyScore = fa.calculateEfficiencyScore(financialData)

	// Calculate overall stability score
	stabilityMetrics.StabilityScore = fa.calculateStabilityScore(stabilityMetrics)

	// Determine overall stability
	stabilityMetrics.OverallStability = fa.determineOverallStability(stabilityMetrics.StabilityScore)

	// Determine risk level
	stabilityMetrics.RiskLevel = fa.determineRiskLevel(stabilityMetrics.StabilityScore)

	return stabilityMetrics, nil
}

// Scoring methods

func (fa *FinancialAnalyzer) calculateRevenueScore(analysis *RevenueAnalysis) float64 {
	score := 0.5 // Base score

	// Revenue growth contribution (30% weight)
	if analysis.RevenueGrowth > 0.1 {
		score += 0.3
	} else if analysis.RevenueGrowth > 0.05 {
		score += 0.2
	} else if analysis.RevenueGrowth > 0 {
		score += 0.1
	}

	// Revenue stability contribution (40% weight)
	score += analysis.RevenueStability * 0.4

	// Revenue diversity contribution (30% weight)
	score += analysis.RevenueDiversity * 0.3

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateGrowthScore(analysis *GrowthAnalysis) float64 {
	score := 0.5 // Base score

	// Revenue growth contribution (25% weight)
	if analysis.RevenueGrowthRate > 0.1 {
		score += 0.25
	} else if analysis.RevenueGrowthRate > 0.05 {
		score += 0.15
	} else if analysis.RevenueGrowthRate > 0 {
		score += 0.05
	}

	// Profit growth contribution (25% weight)
	if analysis.ProfitGrowthRate > 0.08 {
		score += 0.25
	} else if analysis.ProfitGrowthRate > 0.04 {
		score += 0.15
	} else if analysis.ProfitGrowthRate > 0 {
		score += 0.05
	}

	// Asset growth contribution (25% weight)
	if analysis.AssetGrowthRate > 0.05 {
		score += 0.25
	} else if analysis.AssetGrowthRate > 0.02 {
		score += 0.15
	} else if analysis.AssetGrowthRate > 0 {
		score += 0.05
	}

	// Market share growth contribution (25% weight)
	if analysis.MarketShareGrowth > 0.02 {
		score += 0.25
	} else if analysis.MarketShareGrowth > 0.01 {
		score += 0.15
	} else if analysis.MarketShareGrowth > 0 {
		score += 0.05
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateLiquidityScore(financialData *FinancialData) float64 {
	score := 0.5 // Base score

	// Current ratio contribution (50% weight)
	if financialData.CurrentRatio > 2.0 {
		score += 0.5
	} else if financialData.CurrentRatio > 1.5 {
		score += 0.4
	} else if financialData.CurrentRatio > 1.0 {
		score += 0.2
	}

	// Quick ratio contribution (50% weight)
	if financialData.QuickRatio > 1.5 {
		score += 0.5
	} else if financialData.QuickRatio > 1.0 {
		score += 0.4
	} else if financialData.QuickRatio > 0.5 {
		score += 0.2
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateSolvencyScore(financialData *FinancialData) float64 {
	score := 0.5 // Base score

	// Debt-to-equity ratio contribution (60% weight)
	if financialData.DebtToEquityRatio < 0.5 {
		score += 0.6
	} else if financialData.DebtToEquityRatio < 1.0 {
		score += 0.4
	} else if financialData.DebtToEquityRatio < 2.0 {
		score += 0.2
	}

	// Working capital contribution (40% weight)
	workingCapitalRatio := financialData.WorkingCapital / financialData.TotalAssets
	if workingCapitalRatio > 0.3 {
		score += 0.4
	} else if workingCapitalRatio > 0.2 {
		score += 0.3
	} else if workingCapitalRatio > 0.1 {
		score += 0.2
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateProfitabilityScore(financialData *FinancialData) float64 {
	score := 0.5 // Base score

	// Profit margin contribution (40% weight)
	if financialData.ProfitMargin > 0.2 {
		score += 0.4
	} else if financialData.ProfitMargin > 0.15 {
		score += 0.3
	} else if financialData.ProfitMargin > 0.1 {
		score += 0.2
	}

	// ROI contribution (30% weight)
	if financialData.ROI > 0.15 {
		score += 0.3
	} else if financialData.ROI > 0.1 {
		score += 0.2
	} else if financialData.ROI > 0.05 {
		score += 0.1
	}

	// ROE contribution (30% weight)
	if financialData.ROE > 0.2 {
		score += 0.3
	} else if financialData.ROE > 0.15 {
		score += 0.2
	} else if financialData.ROE > 0.1 {
		score += 0.1
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateEfficiencyScore(financialData *FinancialData) float64 {
	score := 0.5 // Base score

	// ROA contribution (50% weight)
	if financialData.ROA > 0.1 {
		score += 0.5
	} else if financialData.ROA > 0.05 {
		score += 0.3
	} else if financialData.ROA > 0.02 {
		score += 0.1
	}

	// Asset turnover contribution (50% weight)
	assetTurnover := financialData.AnnualRevenue / financialData.TotalAssets
	if assetTurnover > 2.0 {
		score += 0.5
	} else if assetTurnover > 1.5 {
		score += 0.3
	} else if assetTurnover > 1.0 {
		score += 0.1
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateStabilityScore(metrics *StabilityMetrics) float64 {
	// Weighted average of all stability metrics
	score := metrics.LiquidityScore*0.25 +
		metrics.SolvencyScore*0.25 +
		metrics.ProfitabilityScore*0.25 +
		metrics.EfficiencyScore*0.25

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateOverallScore(result *FinancialAnalysisResult) float64 {
	score := 0.5 // Base score

	// Financial data score (30% weight)
	if result.FinancialData != nil {
		financialScore := fa.calculateFinancialDataScore(result.FinancialData)
		score += financialScore * 0.30
	}

	// Revenue analysis score (25% weight)
	if result.RevenueAnalysis != nil {
		score += result.RevenueAnalysis.RevenueScore * 0.25
	}

	// Growth analysis score (20% weight)
	if result.GrowthAnalysis != nil {
		score += result.GrowthAnalysis.GrowthScore * 0.20
	}

	// Stability metrics score (25% weight)
	if result.StabilityMetrics != nil {
		score += result.StabilityMetrics.StabilityScore * 0.25
	}

	return fa.max(0.0, fa.min(1.0, score))
}

func (fa *FinancialAnalyzer) calculateFinancialDataScore(financialData *FinancialData) float64 {
	score := 0.5 // Base score

	// Profit margin contribution (30% weight)
	if financialData.ProfitMargin > 0.15 {
		score += 0.3
	} else if financialData.ProfitMargin > 0.1 {
		score += 0.2
	} else if financialData.ProfitMargin > 0.05 {
		score += 0.1
	}

	// Current ratio contribution (25% weight)
	if financialData.CurrentRatio > 1.5 {
		score += 0.25
	} else if financialData.CurrentRatio > 1.0 {
		score += 0.15
	} else if financialData.CurrentRatio > 0.5 {
		score += 0.05
	}

	// Debt-to-equity ratio contribution (25% weight)
	if financialData.DebtToEquityRatio < 1.0 {
		score += 0.25
	} else if financialData.DebtToEquityRatio < 2.0 {
		score += 0.15
	} else if financialData.DebtToEquityRatio < 3.0 {
		score += 0.05
	}

	// ROI contribution (20% weight)
	if financialData.ROI > 0.1 {
		score += 0.2
	} else if financialData.ROI > 0.05 {
		score += 0.1
	} else if financialData.ROI > 0.02 {
		score += 0.05
	}

	return fa.max(0.0, fa.min(1.0, score))
}

// Risk determination methods

func (fa *FinancialAnalyzer) determineRevenueRisk(analysis *RevenueAnalysis) string {
	if analysis.RevenueScore >= 0.8 {
		return "low"
	} else if analysis.RevenueScore >= 0.6 {
		return "medium"
	} else if analysis.RevenueScore >= 0.4 {
		return "high"
	}
	return "critical"
}

func (fa *FinancialAnalyzer) determineGrowthSustainability(analysis *GrowthAnalysis) string {
	if analysis.GrowthScore >= 0.8 {
		return "sustainable"
	} else if analysis.GrowthScore >= 0.6 {
		return "moderate"
	} else if analysis.GrowthScore >= 0.4 {
		return "unsustainable"
	}
	return "declining"
}

func (fa *FinancialAnalyzer) determineOverallStability(score float64) string {
	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.7 {
		return "good"
	} else if score >= 0.5 {
		return "fair"
	} else if score >= 0.3 {
		return "poor"
	}
	return "critical"
}

func (fa *FinancialAnalyzer) determineRiskLevel(score float64) string {
	if score >= 0.8 {
		return "low"
	} else if score >= 0.6 {
		return "medium"
	} else if score >= 0.4 {
		return "high"
	}
	return "critical"
}

// Recommendation generation

func (fa *FinancialAnalyzer) generateRecommendations(result *FinancialAnalysisResult) []string {
	var recommendations []string

	// Financial data recommendations
	if result.FinancialData != nil {
		if result.FinancialData.ProfitMargin < 0.1 {
			recommendations = append(recommendations, "Improve profit margins through cost optimization and pricing strategies")
		}
		if result.FinancialData.CurrentRatio < 1.0 {
			recommendations = append(recommendations, "Improve liquidity by increasing current assets or reducing current liabilities")
		}
		if result.FinancialData.DebtToEquityRatio > 2.0 {
			recommendations = append(recommendations, "Reduce debt levels to improve financial stability")
		}
		if result.FinancialData.ROI < 0.05 {
			recommendations = append(recommendations, "Improve return on investment through better asset utilization")
		}
	}

	// Revenue analysis recommendations
	if result.RevenueAnalysis != nil {
		if result.RevenueAnalysis.RevenueGrowth < 0 {
			recommendations = append(recommendations, "Address declining revenue through market expansion or product diversification")
		}
		if result.RevenueAnalysis.RevenueStability < 0.7 {
			recommendations = append(recommendations, "Improve revenue stability through better cash flow management")
		}
		if result.RevenueAnalysis.RevenueDiversity < 0.5 {
			recommendations = append(recommendations, "Diversify revenue streams to reduce dependency on single sources")
		}
	}

	// Growth analysis recommendations
	if result.GrowthAnalysis != nil {
		if result.GrowthAnalysis.GrowthSustainability == "unsustainable" {
			recommendations = append(recommendations, "Review growth strategy to ensure sustainable long-term growth")
		}
		if result.GrowthAnalysis.RevenueGrowthRate < 0.05 {
			recommendations = append(recommendations, "Implement strategies to accelerate revenue growth")
		}
	}

	// Stability metrics recommendations
	if result.StabilityMetrics != nil {
		if result.StabilityMetrics.LiquidityScore < 0.6 {
			recommendations = append(recommendations, "Improve liquidity position to meet short-term obligations")
		}
		if result.StabilityMetrics.SolvencyScore < 0.6 {
			recommendations = append(recommendations, "Strengthen solvency position through debt reduction or equity increase")
		}
		if result.StabilityMetrics.ProfitabilityScore < 0.6 {
			recommendations = append(recommendations, "Focus on improving profitability through operational efficiency")
		}
	}

	return recommendations
}

// Utility methods

func (fa *FinancialAnalyzer) max(a, b float64) float64 {
	return math.Max(a, b)
}

func (fa *FinancialAnalyzer) min(a, b float64) float64 {
	return math.Min(a, b)
}
