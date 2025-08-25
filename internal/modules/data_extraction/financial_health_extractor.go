package data_extraction

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FinancialHealthExtractor extracts financial health indicators from business data
type FinancialHealthExtractor struct {
	// Configuration
	config *FinancialHealthConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Extraction components
	fundingDetector    *FundingDetector
	revenueExtractor   *RevenueExtractor
	stabilityDetector  *StabilityDetector
	creditRiskDetector *CreditRiskDetector
	financialAnalyzer  *FinancialAnalyzer

	// Thread safety
	extractionMux sync.RWMutex
}

// FinancialHealthConfig configuration for financial health extraction
type FinancialHealthConfig struct {
	// Funding detection settings
	FundingDetectionEnabled bool
	FundingPatterns         []string
	FundingKeywords         []string
	FundingAmountPatterns   []string

	// Revenue extraction settings
	RevenueExtractionEnabled bool
	RevenuePatterns          []string
	RevenueKeywords          []string
	RevenueAmountPatterns    []string

	// Stability detection settings
	StabilityDetectionEnabled bool
	StabilityIndicators       []string
	StabilityKeywords         []string
	StabilityThresholds       map[string]float64

	// Credit risk detection settings
	CreditRiskDetectionEnabled bool
	CreditRiskPatterns         []string
	CreditRiskKeywords         []string
	CreditRiskThresholds       map[string]float64

	// Analysis settings
	AnalysisEnabled     bool
	ConfidenceThreshold float64
	MaxExtractionTime   time.Duration
}

// FinancialHealthData represents extracted financial health data
type FinancialHealthData struct {
	// Funding information
	FundingInfo *FundingInfo

	// Revenue information
	RevenueInfo *RevenueInfo

	// Stability indicators
	StabilityInfo *StabilityInfo

	// Credit risk indicators
	CreditRiskInfo *CreditRiskInfo

	// Analysis results
	Analysis *FinancialAnalysis

	// Metadata
	ExtractionTime time.Time
	Confidence     float64
	Sources        []string
}

// FundingInfo represents funding-related information
type FundingInfo struct {
	HasFunding      bool
	FundingAmount   float64
	FundingCurrency string
	FundingType     string
	FundingDate     time.Time
	Investors       []string
	FundingRound    string
	Confidence      float64
	Sources         []string
}

// RevenueInfo represents revenue-related information
type RevenueInfo struct {
	RevenueRange    string
	RevenueAmount   float64
	RevenueCurrency string
	RevenuePeriod   string
	RevenueGrowth   float64
	RevenueSources  []string
	Confidence      float64
	Sources         []string
}

// StabilityInfo represents financial stability indicators
type StabilityInfo struct {
	StabilityScore   float64
	StabilityLevel   string
	StabilityFactors []string
	RiskFactors      []string
	Confidence       float64
	Sources          []string
}

// CreditRiskInfo represents credit risk indicators
type CreditRiskInfo struct {
	RiskScore       float64
	RiskLevel       string
	RiskFactors     []string
	CreditHistory   string
	PaymentBehavior string
	Confidence      float64
	Sources         []string
}

// FinancialAnalysis represents comprehensive financial analysis
type FinancialAnalysis struct {
	OverallHealth   string
	HealthScore     float64
	KeyStrengths    []string
	KeyRisks        []string
	Recommendations []string
	Confidence      float64
}

// FundingDetector detects funding information
type FundingDetector struct {
	enabled        bool
	patterns       []*regexp.Regexp
	keywords       []string
	amountPatterns []*regexp.Regexp
	detectionMux   sync.RWMutex
}

// RevenueExtractor extracts revenue information
type RevenueExtractor struct {
	enabled        bool
	patterns       []*regexp.Regexp
	keywords       []string
	amountPatterns []*regexp.Regexp
	extractionMux  sync.RWMutex
}

// StabilityDetector detects financial stability indicators
type StabilityDetector struct {
	enabled      bool
	indicators   []string
	keywords     []string
	thresholds   map[string]float64
	detectionMux sync.RWMutex
}

// CreditRiskDetector detects credit risk indicators
type CreditRiskDetector struct {
	enabled      bool
	patterns     []*regexp.Regexp
	keywords     []string
	thresholds   map[string]float64
	detectionMux sync.RWMutex
}

// FinancialAnalyzer performs comprehensive financial analysis
type FinancialAnalyzer struct {
	enabled             bool
	confidenceThreshold float64
	analysisMux         sync.RWMutex
}

// NewFinancialHealthExtractor creates a new financial health extractor
func NewFinancialHealthExtractor(config *FinancialHealthConfig, logger *observability.Logger, tracer trace.Tracer) *FinancialHealthExtractor {
	if config == nil {
		config = &FinancialHealthConfig{
			FundingDetectionEnabled: true,
			FundingPatterns: []string{
				`(?i)(funded|funding|investment|venture|capital|series|round)`,
				`(?i)(raised|secured|obtained|received)\s+\$?([0-9,]+[kmb]?)`,
				`(?i)(seed|series\s+[a-z]|angel|venture|private\s+equity)`,
			},
			FundingKeywords: []string{
				"funded", "funding", "investment", "venture", "capital", "series", "round",
				"raised", "secured", "obtained", "received", "seed", "angel", "equity",
			},
			FundingAmountPatterns: []string{
				`\$?([0-9,]+[kmb]?)`,
				`([0-9,]+[kmb]?)\s*(million|billion|thousand)`,
			},
			RevenueExtractionEnabled: true,
			RevenuePatterns: []string{
				`(?i)(revenue|sales|income|earnings|turnover)`,
				`(?i)(annual|yearly|monthly|quarterly)\s+(revenue|sales|income)`,
				`(?i)(revenue|sales)\s+of\s+\$?([0-9,]+[kmb]?)`,
			},
			RevenueKeywords: []string{
				"revenue", "sales", "income", "earnings", "turnover", "annual", "yearly",
				"monthly", "quarterly", "profit", "gross", "net",
			},
			RevenueAmountPatterns: []string{
				`\$?([0-9,]+[kmb]?)`,
				`([0-9,]+[kmb]?)\s*(million|billion|thousand)`,
			},
			StabilityDetectionEnabled: true,
			StabilityIndicators: []string{
				"profitable", "profitable", "growth", "expanding", "stable", "established",
				"successful", "thriving", "profitable", "cash flow", "liquidity",
			},
			StabilityKeywords: []string{
				"profitable", "growth", "expanding", "stable", "established", "successful",
				"thriving", "cash flow", "liquidity", "solvent", "viable",
			},
			StabilityThresholds: map[string]float64{
				"min_stability_score": 0.3,
				"max_risk_factors":    5.0,
			},
			CreditRiskDetectionEnabled: true,
			CreditRiskPatterns: []string{
				`(?i)(bankruptcy|insolvency|liquidation|receivership)`,
				`(?i)(debt|liability|obligation|credit\s+issue)`,
				`(?i)(payment\s+default|late\s+payment|credit\s+risk)`,
			},
			CreditRiskKeywords: []string{
				"bankruptcy", "insolvency", "liquidation", "receivership", "debt",
				"liability", "obligation", "credit issue", "payment default",
				"late payment", "credit risk", "financial distress",
			},
			CreditRiskThresholds: map[string]float64{
				"max_risk_score":   0.7,
				"min_credit_score": 300.0,
			},
			AnalysisEnabled:     true,
			ConfidenceThreshold: 0.6,
			MaxExtractionTime:   30 * time.Second,
		}
	}

	fhe := &FinancialHealthExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize components
	fhe.fundingDetector = &FundingDetector{
		enabled:        config.FundingDetectionEnabled,
		patterns:       compilePatterns(config.FundingPatterns),
		keywords:       config.FundingKeywords,
		amountPatterns: compilePatterns(config.FundingAmountPatterns),
	}

	fhe.revenueExtractor = &RevenueExtractor{
		enabled:        config.RevenueExtractionEnabled,
		patterns:       compilePatterns(config.RevenuePatterns),
		keywords:       config.RevenueKeywords,
		amountPatterns: compilePatterns(config.RevenueAmountPatterns),
	}

	fhe.stabilityDetector = &StabilityDetector{
		enabled:    config.StabilityDetectionEnabled,
		indicators: config.StabilityIndicators,
		keywords:   config.StabilityKeywords,
		thresholds: config.StabilityThresholds,
	}

	fhe.creditRiskDetector = &CreditRiskDetector{
		enabled:    config.CreditRiskDetectionEnabled,
		patterns:   compilePatterns(config.CreditRiskPatterns),
		keywords:   config.CreditRiskKeywords,
		thresholds: config.CreditRiskThresholds,
	}

	fhe.financialAnalyzer = &FinancialAnalyzer{
		enabled:             config.AnalysisEnabled,
		confidenceThreshold: config.ConfidenceThreshold,
	}

	return fhe
}

// ExtractFinancialHealth extracts financial health data from business information
func (fhe *FinancialHealthExtractor) ExtractFinancialHealth(ctx context.Context, businessName, websiteContent, description string) (*FinancialHealthData, error) {
	ctx, span := fhe.tracer.Start(ctx, "FinancialHealthExtractor.ExtractFinancialHealth")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessName),
		attribute.Int("content_length", len(websiteContent)),
		attribute.Int("description_length", len(description)),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, fhe.config.MaxExtractionTime)
	defer cancel()

	// Combine all text for analysis
	combinedText := strings.Join([]string{businessName, websiteContent, description}, " ")

	// Extract funding information
	var fundingInfo *FundingInfo
	if fhe.config.FundingDetectionEnabled {
		fundingInfo = fhe.fundingDetector.DetectFunding(ctx, combinedText)
	}

	// Extract revenue information
	var revenueInfo *RevenueInfo
	if fhe.config.RevenueExtractionEnabled {
		revenueInfo = fhe.revenueExtractor.ExtractRevenue(ctx, combinedText)
	}

	// Detect stability indicators
	var stabilityInfo *StabilityInfo
	if fhe.config.StabilityDetectionEnabled {
		stabilityInfo = fhe.stabilityDetector.DetectStability(ctx, combinedText)
	}

	// Detect credit risk indicators
	var creditRiskInfo *CreditRiskInfo
	if fhe.config.CreditRiskDetectionEnabled {
		creditRiskInfo = fhe.creditRiskDetector.DetectCreditRisk(ctx, combinedText)
	}

	// Perform comprehensive analysis
	var analysis *FinancialAnalysis
	if fhe.config.AnalysisEnabled {
		analysis = fhe.financialAnalyzer.AnalyzeFinancialHealth(ctx, fundingInfo, revenueInfo, stabilityInfo, creditRiskInfo)
	}

	// Calculate overall confidence
	confidence := fhe.calculateOverallConfidence(fundingInfo, revenueInfo, stabilityInfo, creditRiskInfo)

	// Collect sources
	sources := fhe.collectSources(fundingInfo, revenueInfo, stabilityInfo, creditRiskInfo)

	result := &FinancialHealthData{
		FundingInfo:    fundingInfo,
		RevenueInfo:    revenueInfo,
		StabilityInfo:  stabilityInfo,
		CreditRiskInfo: creditRiskInfo,
		Analysis:       analysis,
		ExtractionTime: time.Now(),
		Confidence:     confidence,
		Sources:        sources,
	}

	fhe.logger.Info("financial health extraction completed", map[string]interface{}{
		"business_name":   businessName,
		"confidence":      confidence,
		"has_funding":     fundingInfo != nil && fundingInfo.HasFunding,
		"has_revenue":     revenueInfo != nil && revenueInfo.RevenueAmount > 0,
		"stability_score": stabilityInfo != nil && stabilityInfo.StabilityScore,
		"risk_score":      creditRiskInfo != nil && creditRiskInfo.RiskScore,
	})

	return result, nil
}

// FundingDetector methods

func (fd *FundingDetector) DetectFunding(ctx context.Context, text string) *FundingInfo {
	fd.detectionMux.Lock()
	defer fd.detectionMux.Unlock()

	// Check for funding keywords
	hasFunding := fd.checkFundingKeywords(text)
	if !hasFunding {
		return &FundingInfo{
			HasFunding: false,
			Confidence: 0.0,
		}
	}

	// Extract funding amount
	amount, currency := fd.extractFundingAmount(text)

	// Determine funding type
	fundingType := fd.determineFundingType(text)

	// Extract funding date
	fundingDate := fd.extractFundingDate(text)

	// Extract investors
	investors := fd.extractInvestors(text)

	// Determine funding round
	fundingRound := fd.determineFundingRound(text)

	// Calculate confidence
	confidence := fd.calculateFundingConfidence(text, amount, fundingType)

	return &FundingInfo{
		HasFunding:      true,
		FundingAmount:   amount,
		FundingCurrency: currency,
		FundingType:     fundingType,
		FundingDate:     fundingDate,
		Investors:       investors,
		FundingRound:    fundingRound,
		Confidence:      confidence,
		Sources:         []string{"text_analysis"},
	}
}

func (fd *FundingDetector) checkFundingKeywords(text string) bool {
	text = strings.ToLower(text)
	for _, keyword := range fd.keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (fd *FundingDetector) extractFundingAmount(text string) (float64, string) {
	text = strings.ToLower(text)

	for _, pattern := range fd.amountPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			amountStr := matches[1]
			amount, currency := fd.parseAmount(amountStr)
			if amount > 0 {
				return amount, currency
			}
		}
	}

	return 0.0, "USD"
}

func (fd *FundingDetector) parseAmount(amountStr string) (float64, string) {
	amountStr = strings.ToLower(strings.TrimSpace(amountStr))

	// Remove common suffixes and convert to number
	var multiplier float64 = 1.0
	if strings.HasSuffix(amountStr, "b") || strings.HasSuffix(amountStr, "billion") {
		multiplier = 1000000000
		amountStr = strings.TrimSuffix(amountStr, "b")
		amountStr = strings.TrimSuffix(amountStr, "billion")
	} else if strings.HasSuffix(amountStr, "m") || strings.HasSuffix(amountStr, "million") {
		multiplier = 1000000
		amountStr = strings.TrimSuffix(amountStr, "m")
		amountStr = strings.TrimSuffix(amountStr, "million")
	} else if strings.HasSuffix(amountStr, "k") || strings.HasSuffix(amountStr, "thousand") {
		multiplier = 1000
		amountStr = strings.TrimSuffix(amountStr, "k")
		amountStr = strings.TrimSuffix(amountStr, "thousand")
	}

	// Remove commas and parse
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0.0, "USD"
	}

	return amount * multiplier, "USD"
}

func (fd *FundingDetector) determineFundingType(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "seed") {
		return "seed"
	} else if strings.Contains(text, "series a") {
		return "series_a"
	} else if strings.Contains(text, "series b") {
		return "series_b"
	} else if strings.Contains(text, "series c") {
		return "series_c"
	} else if strings.Contains(text, "angel") {
		return "angel"
	} else if strings.Contains(text, "venture") {
		return "venture"
	} else if strings.Contains(text, "private equity") {
		return "private_equity"
	}

	return "unknown"
}

func (fd *FundingDetector) extractFundingDate(text string) time.Time {
	// Simple date extraction - in production, use more sophisticated date parsing
	// For now, return current time as placeholder
	return time.Now()
}

func (fd *FundingDetector) extractInvestors(text string) []string {
	// Simple investor extraction - in production, use more sophisticated parsing
	// For now, return empty slice as placeholder
	return []string{}
}

func (fd *FundingDetector) determineFundingRound(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "seed") {
		return "seed"
	} else if strings.Contains(text, "series a") {
		return "series_a"
	} else if strings.Contains(text, "series b") {
		return "series_b"
	} else if strings.Contains(text, "series c") {
		return "series_c"
	}

	return "unknown"
}

func (fd *FundingDetector) calculateFundingConfidence(text string, amount float64, fundingType string) float64 {
	confidence := 0.0

	// Base confidence from keyword presence
	if fd.checkFundingKeywords(text) {
		confidence += 0.3
	}

	// Confidence from amount extraction
	if amount > 0 {
		confidence += 0.4
	}

	// Confidence from funding type identification
	if fundingType != "unknown" {
		confidence += 0.3
	}

	return confidence
}

// RevenueExtractor methods

func (re *RevenueExtractor) ExtractRevenue(ctx context.Context, text string) *RevenueInfo {
	re.extractionMux.Lock()
	defer re.extractionMux.Unlock()

	// Check for revenue keywords
	hasRevenue := re.checkRevenueKeywords(text)
	if !hasRevenue {
		return &RevenueInfo{
			Confidence: 0.0,
		}
	}

	// Extract revenue amount
	amount, currency := re.extractRevenueAmount(text)

	// Determine revenue period
	period := re.determineRevenuePeriod(text)

	// Calculate revenue growth
	growth := re.calculateRevenueGrowth(text)

	// Extract revenue sources
	sources := re.extractRevenueSources(text)

	// Calculate confidence
	confidence := re.calculateRevenueConfidence(text, amount, period)

	return &RevenueInfo{
		RevenueRange:    re.determineRevenueRange(amount),
		RevenueAmount:   amount,
		RevenueCurrency: currency,
		RevenuePeriod:   period,
		RevenueGrowth:   growth,
		RevenueSources:  sources,
		Confidence:      confidence,
		Sources:         []string{"text_analysis"},
	}
}

func (re *RevenueExtractor) checkRevenueKeywords(text string) bool {
	text = strings.ToLower(text)
	for _, keyword := range re.keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func (re *RevenueExtractor) extractRevenueAmount(text string) (float64, string) {
	text = strings.ToLower(text)

	for _, pattern := range re.amountPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			amountStr := matches[1]
			amount, currency := re.parseAmount(amountStr)
			if amount > 0 {
				return amount, currency
			}
		}
	}

	return 0.0, "USD"
}

func (re *RevenueExtractor) parseAmount(amountStr string) (float64, string) {
	// Same logic as funding detector
	return parseAmountHelper(amountStr)
}

func (re *RevenueExtractor) determineRevenuePeriod(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "annual") || strings.Contains(text, "yearly") {
		return "annual"
	} else if strings.Contains(text, "monthly") {
		return "monthly"
	} else if strings.Contains(text, "quarterly") {
		return "quarterly"
	}

	return "annual" // Default to annual
}

func (re *RevenueExtractor) calculateRevenueGrowth(text string) float64 {
	// Simple growth calculation - in production, use more sophisticated analysis
	// For now, return 0 as placeholder
	return 0.0
}

func (re *RevenueExtractor) extractRevenueSources(text string) []string {
	// Simple revenue source extraction - in production, use more sophisticated parsing
	// For now, return empty slice as placeholder
	return []string{}
}

func (re *RevenueExtractor) determineRevenueRange(amount float64) string {
	if amount == 0 {
		return "unknown"
	} else if amount < 1000000 {
		return "under_1m"
	} else if amount < 10000000 {
		return "1m_10m"
	} else if amount < 100000000 {
		return "10m_100m"
	} else {
		return "over_100m"
	}
}

func (re *RevenueExtractor) calculateRevenueConfidence(text string, amount float64, period string) float64 {
	confidence := 0.0

	// Base confidence from keyword presence
	if re.checkRevenueKeywords(text) {
		confidence += 0.3
	}

	// Confidence from amount extraction
	if amount > 0 {
		confidence += 0.4
	}

	// Confidence from period identification
	if period != "unknown" {
		confidence += 0.3
	}

	return confidence
}

// StabilityDetector methods

func (sd *StabilityDetector) DetectStability(ctx context.Context, text string) *StabilityInfo {
	sd.detectionMux.Lock()
	defer sd.detectionMux.Unlock()

	// Calculate stability score
	score := sd.calculateStabilityScore(text)

	// Determine stability level
	level := sd.determineStabilityLevel(score)

	// Identify stability factors
	factors := sd.identifyStabilityFactors(text)

	// Identify risk factors
	risks := sd.identifyRiskFactors(text)

	// Calculate confidence
	confidence := sd.calculateStabilityConfidence(text, score, factors)

	return &StabilityInfo{
		StabilityScore:   score,
		StabilityLevel:   level,
		StabilityFactors: factors,
		RiskFactors:      risks,
		Confidence:       confidence,
		Sources:          []string{"text_analysis"},
	}
}

func (sd *StabilityDetector) calculateStabilityScore(text string) float64 {
	text = strings.ToLower(text)
	score := 0.5 // Base score

	// Positive indicators
	for _, keyword := range sd.keywords {
		if strings.Contains(text, keyword) {
			score += 0.1
		}
	}

	// Negative indicators
	negativeKeywords := []string{"struggling", "failing", "bankruptcy", "insolvency", "debt"}
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			score -= 0.2
		}
	}

	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

func (sd *StabilityDetector) determineStabilityLevel(score float64) string {
	if score >= 0.8 {
		return "high"
	} else if score >= 0.6 {
		return "medium"
	} else if score >= 0.4 {
		return "low"
	} else {
		return "very_low"
	}
}

func (sd *StabilityDetector) identifyStabilityFactors(text string) []string {
	text = strings.ToLower(text)
	var factors []string

	for _, keyword := range sd.keywords {
		if strings.Contains(text, keyword) {
			factors = append(factors, keyword)
		}
	}

	return factors
}

func (sd *StabilityDetector) identifyRiskFactors(text string) []string {
	text = strings.ToLower(text)
	riskKeywords := []string{"struggling", "failing", "bankruptcy", "insolvency", "debt", "loss", "decline"}
	var risks []string

	for _, keyword := range riskKeywords {
		if strings.Contains(text, keyword) {
			risks = append(risks, keyword)
		}
	}

	return risks
}

func (sd *StabilityDetector) calculateStabilityConfidence(text string, score float64, factors []string) float64 {
	confidence := 0.0

	// Base confidence from score calculation
	confidence += 0.4

	// Confidence from factor identification
	if len(factors) > 0 {
		confidence += 0.3
	}

	// Confidence from text length (more text = more confidence)
	if len(text) > 1000 {
		confidence += 0.3
	} else if len(text) > 500 {
		confidence += 0.2
	} else if len(text) > 100 {
		confidence += 0.1
	}

	return confidence
}

// CreditRiskDetector methods

func (crd *CreditRiskDetector) DetectCreditRisk(ctx context.Context, text string) *CreditRiskInfo {
	crd.detectionMux.Lock()
	defer crd.detectionMux.Unlock()

	// Calculate risk score
	score := crd.calculateRiskScore(text)

	// Determine risk level
	level := crd.determineRiskLevel(score)

	// Identify risk factors
	factors := crd.identifyRiskFactors(text)

	// Determine credit history
	history := crd.determineCreditHistory(text)

	// Determine payment behavior
	behavior := crd.determinePaymentBehavior(text)

	// Calculate confidence
	confidence := crd.calculateRiskConfidence(text, score, factors)

	return &CreditRiskInfo{
		RiskScore:       score,
		RiskLevel:       level,
		RiskFactors:     factors,
		CreditHistory:   history,
		PaymentBehavior: behavior,
		Confidence:      confidence,
		Sources:         []string{"text_analysis"},
	}
}

func (crd *CreditRiskDetector) calculateRiskScore(text string) float64 {
	text = strings.ToLower(text)
	score := 0.3 // Base score (moderate risk)

	// High risk indicators
	highRiskKeywords := []string{"bankruptcy", "insolvency", "liquidation", "receivership"}
	for _, keyword := range highRiskKeywords {
		if strings.Contains(text, keyword) {
			score += 0.4
		}
	}

	// Medium risk indicators
	mediumRiskKeywords := []string{"debt", "liability", "obligation", "credit issue"}
	for _, keyword := range mediumRiskKeywords {
		if strings.Contains(text, keyword) {
			score += 0.2
		}
	}

	// Low risk indicators
	lowRiskKeywords := []string{"profitable", "stable", "successful", "thriving"}
	for _, keyword := range lowRiskKeywords {
		if strings.Contains(text, keyword) {
			score -= 0.1
		}
	}

	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

func (crd *CreditRiskDetector) determineRiskLevel(score float64) string {
	if score >= 0.7 {
		return "high"
	} else if score >= 0.5 {
		return "medium"
	} else if score >= 0.3 {
		return "low"
	} else {
		return "very_low"
	}
}

func (crd *CreditRiskDetector) identifyRiskFactors(text string) []string {
	text = strings.ToLower(text)
	var factors []string

	for _, keyword := range crd.keywords {
		if strings.Contains(text, keyword) {
			factors = append(factors, keyword)
		}
	}

	return factors
}

func (crd *CreditRiskDetector) determineCreditHistory(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "bankruptcy") || strings.Contains(text, "insolvency") {
		return "poor"
	} else if strings.Contains(text, "debt") || strings.Contains(text, "credit issue") {
		return "fair"
	} else if strings.Contains(text, "profitable") || strings.Contains(text, "stable") {
		return "good"
	}

	return "unknown"
}

func (crd *CreditRiskDetector) determinePaymentBehavior(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "payment default") || strings.Contains(text, "late payment") {
		return "poor"
	} else if strings.Contains(text, "on time") || strings.Contains(text, "prompt payment") {
		return "good"
	}

	return "unknown"
}

func (crd *CreditRiskDetector) calculateRiskConfidence(text string, score float64, factors []string) float64 {
	confidence := 0.0

	// Base confidence from score calculation
	confidence += 0.4

	// Confidence from factor identification
	if len(factors) > 0 {
		confidence += 0.3
	}

	// Confidence from text length
	if len(text) > 1000 {
		confidence += 0.3
	} else if len(text) > 500 {
		confidence += 0.2
	} else if len(text) > 100 {
		confidence += 0.1
	}

	return confidence
}

// FinancialAnalyzer methods

func (fa *FinancialAnalyzer) AnalyzeFinancialHealth(ctx context.Context, funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) *FinancialAnalysis {
	fa.analysisMux.Lock()
	defer fa.analysisMux.Unlock()

	// Calculate overall health score
	healthScore := fa.calculateHealthScore(funding, revenue, stability, risk)

	// Determine overall health
	overallHealth := fa.determineOverallHealth(healthScore)

	// Identify key strengths
	strengths := fa.identifyKeyStrengths(funding, revenue, stability, risk)

	// Identify key risks
	risks := fa.identifyKeyRisks(funding, revenue, stability, risk)

	// Generate recommendations
	recommendations := fa.generateRecommendations(funding, revenue, stability, risk)

	// Calculate confidence
	confidence := fa.calculateAnalysisConfidence(funding, revenue, stability, risk)

	return &FinancialAnalysis{
		OverallHealth:   overallHealth,
		HealthScore:     healthScore,
		KeyStrengths:    strengths,
		KeyRisks:        risks,
		Recommendations: recommendations,
		Confidence:      confidence,
	}
}

func (fa *FinancialAnalyzer) calculateHealthScore(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) float64 {
	score := 0.5 // Base score

	// Funding contribution
	if funding != nil && funding.HasFunding {
		score += 0.1
		if funding.FundingAmount > 10000000 { // $10M+
			score += 0.1
		}
	}

	// Revenue contribution
	if revenue != nil && revenue.RevenueAmount > 0 {
		score += 0.2
		if revenue.RevenueAmount > 10000000 { // $10M+
			score += 0.1
		}
	}

	// Stability contribution
	if stability != nil {
		score += stability.StabilityScore * 0.3
	}

	// Risk contribution (inverse)
	if risk != nil {
		score -= risk.RiskScore * 0.2
	}

	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

func (fa *FinancialAnalyzer) determineOverallHealth(score float64) string {
	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.7 {
		return "good"
	} else if score >= 0.6 {
		return "fair"
	} else if score >= 0.5 {
		return "poor"
	} else {
		return "critical"
	}
}

func (fa *FinancialAnalyzer) identifyKeyStrengths(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) []string {
	var strengths []string

	if funding != nil && funding.HasFunding {
		strengths = append(strengths, "has_funding")
		if funding.FundingAmount > 10000000 {
			strengths = append(strengths, "significant_funding")
		}
	}

	if revenue != nil && revenue.RevenueAmount > 0 {
		strengths = append(strengths, "has_revenue")
		if revenue.RevenueAmount > 10000000 {
			strengths = append(strengths, "high_revenue")
		}
	}

	if stability != nil && stability.StabilityScore > 0.7 {
		strengths = append(strengths, "financially_stable")
	}

	if risk != nil && risk.RiskScore < 0.3 {
		strengths = append(strengths, "low_credit_risk")
	}

	return strengths
}

func (fa *FinancialAnalyzer) identifyKeyRisks(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) []string {
	var risks []string

	if funding == nil || !funding.HasFunding {
		risks = append(risks, "no_funding_identified")
	}

	if revenue == nil || revenue.RevenueAmount == 0 {
		risks = append(risks, "no_revenue_identified")
	}

	if stability != nil && stability.StabilityScore < 0.4 {
		risks = append(risks, "low_stability")
	}

	if risk != nil && risk.RiskScore > 0.6 {
		risks = append(risks, "high_credit_risk")
	}

	return risks
}

func (fa *FinancialAnalyzer) generateRecommendations(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) []string {
	var recommendations []string

	if funding == nil || !funding.HasFunding {
		recommendations = append(recommendations, "consider_seeking_funding")
	}

	if revenue == nil || revenue.RevenueAmount == 0 {
		recommendations = append(recommendations, "focus_on_revenue_generation")
	}

	if stability != nil && stability.StabilityScore < 0.5 {
		recommendations = append(recommendations, "improve_financial_stability")
	}

	if risk != nil && risk.RiskScore > 0.5 {
		recommendations = append(recommendations, "address_credit_risk_factors")
	}

	return recommendations
}

func (fa *FinancialAnalyzer) calculateAnalysisConfidence(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, risk *CreditRiskInfo) float64 {
	confidence := 0.0
	count := 0

	if funding != nil {
		confidence += funding.Confidence
		count++
	}

	if revenue != nil {
		confidence += revenue.Confidence
		count++
	}

	if stability != nil {
		confidence += stability.Confidence
		count++
	}

	if risk != nil {
		confidence += risk.Confidence
		count++
	}

	if count > 0 {
		return confidence / float64(count)
	}

	return 0.0
}

// Helper methods

func (fhe *FinancialHealthExtractor) calculateOverallConfidence(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, creditRisk *CreditRiskInfo) float64 {
	confidence := 0.0
	count := 0

	if funding != nil {
		confidence += funding.Confidence
		count++
	}

	if revenue != nil {
		confidence += revenue.Confidence
		count++
	}

	if stability != nil {
		confidence += stability.Confidence
		count++
	}

	if creditRisk != nil {
		confidence += creditRisk.Confidence
		count++
	}

	if count > 0 {
		return confidence / float64(count)
	}

	return 0.0
}

func (fhe *FinancialHealthExtractor) collectSources(funding *FundingInfo, revenue *RevenueInfo, stability *StabilityInfo, creditRisk *CreditRiskInfo) []string {
	var sources []string

	if funding != nil {
		sources = append(sources, funding.Sources...)
	}

	if revenue != nil {
		sources = append(sources, revenue.Sources...)
	}

	if stability != nil {
		sources = append(sources, stability.Sources...)
	}

	if creditRisk != nil {
		sources = append(sources, creditRisk.Sources...)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueSources []string
	for _, source := range sources {
		if !seen[source] {
			seen[source] = true
			uniqueSources = append(uniqueSources, source)
		}
	}

	return uniqueSources
}

// Helper function for parsing amounts
func parseAmountHelper(amountStr string) (float64, string) {
	amountStr = strings.ToLower(strings.TrimSpace(amountStr))

	var multiplier float64 = 1.0
	if strings.HasSuffix(amountStr, "b") || strings.HasSuffix(amountStr, "billion") {
		multiplier = 1000000000
		amountStr = strings.TrimSuffix(amountStr, "b")
		amountStr = strings.TrimSuffix(amountStr, "billion")
	} else if strings.HasSuffix(amountStr, "m") || strings.HasSuffix(amountStr, "million") {
		multiplier = 1000000
		amountStr = strings.TrimSuffix(amountStr, "m")
		amountStr = strings.TrimSuffix(amountStr, "million")
	} else if strings.HasSuffix(amountStr, "k") || strings.HasSuffix(amountStr, "thousand") {
		multiplier = 1000
		amountStr = strings.TrimSuffix(amountStr, "k")
		amountStr = strings.TrimSuffix(amountStr, "thousand")
	}

	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0.0, "USD"
	}

	return amount * multiplier, "USD"
}

// Helper function to compile regex patterns
func compilePatterns(patterns []string) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, compiled)
		}
	}
	return compiled
}
