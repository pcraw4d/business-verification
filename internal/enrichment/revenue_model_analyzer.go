package enrichment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RevenueModelAnalyzer analyzes revenue models and pricing strategies from website content
type RevenueModelAnalyzer struct {
	config *RevenueModelConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// RevenueModelConfig configuration for revenue model analysis
type RevenueModelConfig struct {
	// Analysis thresholds
	MinConfidenceThreshold float64
	MinEvidenceCount       int
	MinContentLength       int

	// Scoring weights
	ModelWeight       float64
	PricingWeight     float64
	StrategyWeight    float64
	MarketWeight      float64
	CompetitiveWeight float64

	// Validation settings
	RequireMultipleIndicators bool
	EnableFallbackAnalysis    bool
	ValidateModels            bool
}

// RevenueModelResult comprehensive revenue model analysis result
type RevenueModelResult struct {
	// Primary classifications
	PrimaryRevenueModel   string            `json:"primary_revenue_model"`
	SecondaryRevenueModel string            `json:"secondary_revenue_model,omitempty"`
	PricingStrategies     []PricingStrategy `json:"pricing_strategies"`
	RevenueStreams        []RevenueStream   `json:"revenue_streams"`

	// Detailed analysis
	ModelDetails        RevenueModelDetails `json:"model_details"`
	PricingAnalysis     PricingAnalysis     `json:"pricing_analysis"`
	MarketPositioning   MarketPositioning   `json:"market_positioning"`
	CompetitiveAnalysis CompetitiveAnalysis `json:"competitive_analysis"`

	// Confidence and validation
	ConfidenceScore  float64                `json:"confidence_score"`
	ComponentScores  RevenueComponentScores `json:"component_scores"`
	Evidence         []string               `json:"evidence"`
	ExtractedPhrases []string               `json:"extracted_phrases"`

	// Quality assessment
	IsValidated      bool             `json:"is_validated"`
	ValidationStatus ValidationStatus `json:"validation_status"`
	DataQualityScore float64          `json:"data_quality_score"`
	Reasoning        string           `json:"reasoning"`

	// Metadata
	AnalyzedAt     time.Time     `json:"analyzed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	SourceURL      string        `json:"source_url,omitempty"`
}

// PricingStrategy represents a detailed pricing strategy
type PricingStrategy struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"` // subscription, freemium, marketplace, etc.
	Description      string   `json:"description"`
	Characteristics  []string `json:"characteristics"`
	TargetAudience   string   `json:"target_audience"`
	PriceRange       string   `json:"price_range"`
	BillingFrequency string   `json:"billing_frequency"`
	ConfidenceScore  float64  `json:"confidence_score"`
}

// RevenueStream represents a revenue stream
type RevenueStream struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	RevenueShare    float64 `json:"revenue_share"`
	GrowthPotential string  `json:"growth_potential"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// RevenueModelDetails contains detailed revenue model information
type RevenueModelDetails struct {
	ModelType          string   `json:"model_type"`
	RevenueSources     []string `json:"revenue_sources"`
	CustomerSegments   []string `json:"customer_segments"`
	ValueProposition   string   `json:"value_proposition"`
	MonetizationMethod string   `json:"monetization_method"`
	RevenueMultipliers []string `json:"revenue_multipliers"`
}

// PricingAnalysis contains detailed pricing analysis
type PricingAnalysis struct {
	PricingModel       string   `json:"pricing_model"`
	PricePoints        []string `json:"price_points"`
	DiscountStrategies []string `json:"discount_strategies"`
	PricingTiers       []string `json:"pricing_tiers"`
	CompetitivePricing string   `json:"competitive_pricing"`
	PriceOptimization  string   `json:"price_optimization"`
}

// MarketPositioning contains market positioning analysis
type MarketPositioning struct {
	MarketSegment        string   `json:"market_segment"`
	CompetitiveAdvantage []string `json:"competitive_advantage"`
	MarketShare          string   `json:"market_share"`
	GrowthStrategy       string   `json:"growth_strategy"`
	MarketMaturity       string   `json:"market_maturity"`
}

// CompetitiveAnalysis contains competitive analysis
type CompetitiveAnalysis struct {
	CompetitiveLandscape string   `json:"competitive_landscape"`
	Differentiators      []string `json:"differentiators"`
	CompetitiveThreats   []string `json:"competitive_threats"`
	MarketGaps           []string `json:"market_gaps"`
	CompetitiveResponse  string   `json:"competitive_response"`
}

// RevenueComponentScores detailed scoring breakdown
type RevenueComponentScores struct {
	ModelScore       float64 `json:"model_score"`
	PricingScore     float64 `json:"pricing_score"`
	StrategyScore    float64 `json:"strategy_score"`
	MarketScore      float64 `json:"market_score"`
	CompetitiveScore float64 `json:"competitive_score"`
}

// NewRevenueModelAnalyzer creates a new revenue model analyzer
func NewRevenueModelAnalyzer(config *RevenueModelConfig, logger *zap.Logger) *RevenueModelAnalyzer {
	if config == nil {
		config = GetDefaultRevenueModelConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RevenueModelAnalyzer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("revenue-model-analyzer"),
	}
}

// GetDefaultRevenueModelConfig returns default configuration
func GetDefaultRevenueModelConfig() *RevenueModelConfig {
	return &RevenueModelConfig{
		MinConfidenceThreshold:    0.3,
		MinEvidenceCount:          2,
		MinContentLength:          50,
		ModelWeight:               0.30,
		PricingWeight:             0.25,
		StrategyWeight:            0.20,
		MarketWeight:              0.15,
		CompetitiveWeight:         0.10,
		RequireMultipleIndicators: true,
		EnableFallbackAnalysis:    true,
		ValidateModels:            true,
	}
}

// AnalyzeRevenueModel performs comprehensive revenue model analysis
func (rma *RevenueModelAnalyzer) AnalyzeRevenueModel(ctx context.Context, content, sourceURL string) (*RevenueModelResult, error) {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_revenue_model")
	defer span.End()

	startTime := time.Now()

	// Input validation
	if err := rma.validateInput(content); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	result := &RevenueModelResult{
		SourceURL:        sourceURL,
		AnalyzedAt:       time.Now(),
		Evidence:         []string{},
		ExtractedPhrases: []string{},
		ComponentScores:  RevenueComponentScores{},
	}

	// Perform analysis components
	if err := rma.analyzeRevenueModels(ctx, content, result); err != nil {
		rma.logger.Warn("Revenue model analysis failed", zap.Error(err))
	}

	if err := rma.analyzePricingStrategies(ctx, content, result); err != nil {
		rma.logger.Warn("Pricing strategy analysis failed", zap.Error(err))
	}

	if err := rma.analyzeRevenueStreams(ctx, content, result); err != nil {
		rma.logger.Warn("Revenue stream analysis failed", zap.Error(err))
	}

	if err := rma.analyzeMarketPositioning(ctx, content, result); err != nil {
		rma.logger.Warn("Market positioning analysis failed", zap.Error(err))
	}

	if err := rma.analyzeCompetitiveLandscape(ctx, content, result); err != nil {
		rma.logger.Warn("Competitive analysis failed", zap.Error(err))
	}

	// Determine primary and secondary revenue models
	rma.determinePrimaryRevenueModel(result)

	// Calculate confidence scores
	rma.calculateConfidenceScores(result)

	// Validate results
	rma.validateResult(result)

	// Generate reasoning
	result.Reasoning = rma.generateReasoning(result)

	// Calculate data quality
	result.DataQualityScore = rma.calculateDataQuality(result)

	result.ProcessingTime = time.Since(startTime)

	rma.logger.Info("Revenue model analysis completed",
		zap.String("primary_model", result.PrimaryRevenueModel),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// analyzeRevenueModels identifies primary revenue models
func (rma *RevenueModelAnalyzer) analyzeRevenueModels(ctx context.Context, content string, result *RevenueModelResult) error {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_revenue_models")
	defer span.End()

	contentLower := strings.ToLower(content)
	revenueModels := make(map[string]float64)

	// Subscription model
	subscriptionIndicators := []string{
		"subscription", "monthly", "annual", "yearly", "recurring",
		"monthly plan", "annual plan", "yearly plan", "subscription plan",
		"monthly billing", "annual billing", "recurring billing",
		"subscription service", "monthly service", "annual service",
	}

	for _, indicator := range subscriptionIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["subscription"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Subscription indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Freemium model
	freemiumIndicators := []string{
		"freemium", "free plan", "free tier", "free version",
		"free with premium", "free upgrade", "free to premium",
		"basic free", "premium upgrade", "free features",
		"free account", "free trial", "free forever",
	}

	for _, indicator := range freemiumIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["freemium"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Freemium indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Marketplace model
	marketplaceIndicators := []string{
		"marketplace", "commission", "transaction fee", "platform fee",
		"marketplace fee", "service fee", "processing fee",
		"buyer fee", "seller fee", "marketplace commission",
		"transaction commission", "platform commission",
	}

	for _, indicator := range marketplaceIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["marketplace"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Marketplace indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// One-time purchase model
	oneTimeIndicators := []string{
		"one-time", "single purchase", "buy once", "one-time payment",
		"single payment", "one-time fee", "single fee",
		"one-time purchase", "single purchase", "buy now",
		"one-time cost", "single cost", "one-time price",
	}

	for _, indicator := range oneTimeIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["one_time"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("One-time indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Enterprise model
	enterpriseIndicators := []string{
		"enterprise pricing", "enterprise plan", "enterprise solution",
		"custom pricing", "contact sales", "enterprise quote",
		"enterprise license", "enterprise contract", "enterprise deal",
		"enterprise package", "enterprise tier", "enterprise features",
	}

	for _, indicator := range enterpriseIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["enterprise"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Enterprise indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Advertising model
	advertisingIndicators := []string{
		"advertising", "ads", "ad revenue", "advertising revenue",
		"display ads", "banner ads", "ad space", "advertising space",
		"ad network", "ad platform", "advertising platform",
		"sponsored content", "ad-supported", "free with ads",
	}

	for _, indicator := range advertisingIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["advertising"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Advertising indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Data monetization model
	dataIndicators := []string{
		"data monetization", "data revenue", "data licensing",
		"data insights", "data analytics", "data platform",
		"data marketplace", "data exchange", "data services",
		"data products", "data solutions", "data revenue stream",
	}

	for _, indicator := range dataIndicators {
		if strings.Contains(contentLower, indicator) {
			revenueModels["data_monetization"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Data monetization indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Convert to model details
	for modelType, score := range revenueModels {
		if score > 0 {
			result.ModelDetails.ModelType = modelType
			break
		}
	}

	return nil
}

// analyzePricingStrategies identifies pricing strategies
func (rma *RevenueModelAnalyzer) analyzePricingStrategies(ctx context.Context, content string, result *RevenueModelResult) error {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_pricing_strategies")
	defer span.End()

	contentLower := strings.ToLower(content)
	strategies := []PricingStrategy{}

	// Tiered pricing
	if strings.Contains(contentLower, "tier") || strings.Contains(contentLower, "plan") || strings.Contains(contentLower, "package") {
		strategies = append(strategies, PricingStrategy{
			Name:             "Tiered Pricing",
			Type:             "tiered",
			Description:      "Multiple pricing tiers with different feature sets",
			Characteristics:  []string{"Multiple tiers", "Feature differentiation", "Scalable pricing"},
			TargetAudience:   "Various customer segments",
			PriceRange:       "Variable",
			BillingFrequency: "Monthly/Annual",
			ConfidenceScore:  0.8,
		})
		result.Evidence = append(result.Evidence, "Tiered pricing strategy detected")
	}

	// Value-based pricing
	if strings.Contains(contentLower, "value") || strings.Contains(contentLower, "roi") || strings.Contains(contentLower, "benefit") {
		strategies = append(strategies, PricingStrategy{
			Name:             "Value-Based Pricing",
			Type:             "value_based",
			Description:      "Pricing based on customer value and ROI",
			Characteristics:  []string{"Value-focused", "ROI-driven", "Customer benefit"},
			TargetAudience:   "Enterprise/B2B",
			PriceRange:       "Premium",
			BillingFrequency: "Variable",
			ConfidenceScore:  0.7,
		})
		result.Evidence = append(result.Evidence, "Value-based pricing strategy detected")
	}

	// Penetration pricing
	if strings.Contains(contentLower, "competitive") || strings.Contains(contentLower, "low price") || strings.Contains(contentLower, "affordable") {
		strategies = append(strategies, PricingStrategy{
			Name:             "Penetration Pricing",
			Type:             "penetration",
			Description:      "Low initial pricing to gain market share",
			Characteristics:  []string{"Low initial price", "Market penetration", "Competitive advantage"},
			TargetAudience:   "Price-sensitive customers",
			PriceRange:       "Low to Medium",
			BillingFrequency: "Variable",
			ConfidenceScore:  0.6,
		})
		result.Evidence = append(result.Evidence, "Penetration pricing strategy detected")
	}

	// Premium pricing
	if strings.Contains(contentLower, "premium") || strings.Contains(contentLower, "luxury") || strings.Contains(contentLower, "high-end") {
		strategies = append(strategies, PricingStrategy{
			Name:             "Premium Pricing",
			Type:             "premium",
			Description:      "High pricing for premium positioning",
			Characteristics:  []string{"High price", "Premium quality", "Exclusive positioning"},
			TargetAudience:   "High-end customers",
			PriceRange:       "High",
			BillingFrequency: "Variable",
			ConfidenceScore:  0.8,
		})
		result.Evidence = append(result.Evidence, "Premium pricing strategy detected")
	}

	// Dynamic pricing
	if strings.Contains(contentLower, "dynamic") || strings.Contains(contentLower, "variable") || strings.Contains(contentLower, "flexible") {
		strategies = append(strategies, PricingStrategy{
			Name:             "Dynamic Pricing",
			Type:             "dynamic",
			Description:      "Variable pricing based on demand and market conditions",
			Characteristics:  []string{"Variable pricing", "Market-responsive", "Demand-based"},
			TargetAudience:   "Market-responsive customers",
			PriceRange:       "Variable",
			BillingFrequency: "Variable",
			ConfidenceScore:  0.7,
		})
		result.Evidence = append(result.Evidence, "Dynamic pricing strategy detected")
	}

	result.PricingStrategies = strategies
	return nil
}

// analyzeRevenueStreams identifies revenue streams
func (rma *RevenueModelAnalyzer) analyzeRevenueStreams(ctx context.Context, content string, result *RevenueModelResult) error {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_revenue_streams")
	defer span.End()

	contentLower := strings.ToLower(content)
	streams := []RevenueStream{}

	// Software licensing
	if strings.Contains(contentLower, "license") || strings.Contains(contentLower, "software") || strings.Contains(contentLower, "saas") {
		streams = append(streams, RevenueStream{
			Name:            "Software Licensing",
			Type:            "licensing",
			Description:     "Revenue from software licenses and subscriptions",
			RevenueShare:    0.4,
			GrowthPotential: "High",
			ConfidenceScore: 0.8,
		})
		result.Evidence = append(result.Evidence, "Software licensing revenue stream detected")
	}

	// Transaction fees
	if strings.Contains(contentLower, "transaction") || strings.Contains(contentLower, "commission") || strings.Contains(contentLower, "fee") {
		streams = append(streams, RevenueStream{
			Name:            "Transaction Fees",
			Type:            "transaction",
			Description:     "Revenue from transaction processing fees",
			RevenueShare:    0.3,
			GrowthPotential: "Medium",
			ConfidenceScore: 0.7,
		})
		result.Evidence = append(result.Evidence, "Transaction fees revenue stream detected")
	}

	// Advertising
	if strings.Contains(contentLower, "advertising") || strings.Contains(contentLower, "ads") || strings.Contains(contentLower, "ad revenue") {
		streams = append(streams, RevenueStream{
			Name:            "Advertising",
			Type:            "advertising",
			Description:     "Revenue from advertising and sponsored content",
			RevenueShare:    0.2,
			GrowthPotential: "Medium",
			ConfidenceScore: 0.6,
		})
		result.Evidence = append(result.Evidence, "Advertising revenue stream detected")
	}

	// Data monetization
	if strings.Contains(contentLower, "data") || strings.Contains(contentLower, "analytics") || strings.Contains(contentLower, "insights") {
		streams = append(streams, RevenueStream{
			Name:            "Data Monetization",
			Type:            "data",
			Description:     "Revenue from data insights and analytics",
			RevenueShare:    0.1,
			GrowthPotential: "High",
			ConfidenceScore: 0.5,
		})
		result.Evidence = append(result.Evidence, "Data monetization revenue stream detected")
	}

	result.RevenueStreams = streams
	return nil
}

// analyzeMarketPositioning analyzes market positioning
func (rma *RevenueModelAnalyzer) analyzeMarketPositioning(ctx context.Context, content string, result *RevenueModelResult) error {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_market_positioning")
	defer span.End()

	contentLower := strings.ToLower(content)
	positioning := &MarketPositioning{}

	// Market segment analysis
	if strings.Contains(contentLower, "enterprise") || strings.Contains(contentLower, "b2b") {
		positioning.MarketSegment = "Enterprise/B2B"
	} else if strings.Contains(contentLower, "consumer") || strings.Contains(contentLower, "b2c") {
		positioning.MarketSegment = "Consumer/B2C"
	} else if strings.Contains(contentLower, "marketplace") || strings.Contains(contentLower, "platform") {
		positioning.MarketSegment = "Marketplace/Platform"
	} else {
		positioning.MarketSegment = "Mixed"
	}

	// Competitive advantage
	advantages := []string{}
	if strings.Contains(contentLower, "innovative") || strings.Contains(contentLower, "cutting-edge") {
		advantages = append(advantages, "Innovation")
	}
	if strings.Contains(contentLower, "affordable") || strings.Contains(contentLower, "cost-effective") {
		advantages = append(advantages, "Cost-effectiveness")
	}
	if strings.Contains(contentLower, "premium") || strings.Contains(contentLower, "high-quality") {
		advantages = append(advantages, "Quality")
	}
	if strings.Contains(contentLower, "easy") || strings.Contains(contentLower, "simple") {
		advantages = append(advantages, "Ease of use")
	}

	positioning.CompetitiveAdvantage = advantages

	// Market maturity
	if strings.Contains(contentLower, "startup") || strings.Contains(contentLower, "emerging") {
		positioning.MarketMaturity = "Emerging"
	} else if strings.Contains(contentLower, "growing") || strings.Contains(contentLower, "expanding") {
		positioning.MarketMaturity = "Growing"
	} else if strings.Contains(contentLower, "mature") || strings.Contains(contentLower, "established") {
		positioning.MarketMaturity = "Mature"
	} else {
		positioning.MarketMaturity = "Unknown"
	}

	result.MarketPositioning = *positioning
	return nil
}

// analyzeCompetitiveLandscape analyzes competitive landscape
func (rma *RevenueModelAnalyzer) analyzeCompetitiveLandscape(ctx context.Context, content string, result *RevenueModelResult) error {
	ctx, span := rma.tracer.Start(ctx, "revenue_model_analyzer.analyze_competitive_landscape")
	defer span.End()

	contentLower := strings.ToLower(content)
	competitive := &CompetitiveAnalysis{}

	// Competitive landscape
	if strings.Contains(contentLower, "competitive") || strings.Contains(contentLower, "competition") {
		competitive.CompetitiveLandscape = "High competition"
	} else if strings.Contains(contentLower, "niche") || strings.Contains(contentLower, "specialized") {
		competitive.CompetitiveLandscape = "Niche market"
	} else {
		competitive.CompetitiveLandscape = "Moderate competition"
	}

	// Differentiators
	differentiators := []string{}
	if strings.Contains(contentLower, "unique") || strings.Contains(contentLower, "exclusive") {
		differentiators = append(differentiators, "Unique features")
	}
	if strings.Contains(contentLower, "technology") || strings.Contains(contentLower, "ai") {
		differentiators = append(differentiators, "Technology advantage")
	}
	if strings.Contains(contentLower, "customer service") || strings.Contains(contentLower, "support") {
		differentiators = append(differentiators, "Customer service")
	}

	competitive.Differentiators = differentiators

	// Market gaps
	gaps := []string{}
	if strings.Contains(contentLower, "unmet need") || strings.Contains(contentLower, "gap") {
		gaps = append(gaps, "Unmet customer needs")
	}
	if strings.Contains(contentLower, "inefficient") || strings.Contains(contentLower, "pain point") {
		gaps = append(gaps, "Process inefficiencies")
	}

	competitive.MarketGaps = gaps

	result.CompetitiveAnalysis = *competitive
	return nil
}

// determinePrimaryRevenueModel determines primary and secondary revenue models
func (rma *RevenueModelAnalyzer) determinePrimaryRevenueModel(result *RevenueModelResult) {
	if result.ModelDetails.ModelType == "" {
		result.PrimaryRevenueModel = "unknown"
		return
	}

	result.PrimaryRevenueModel = result.ModelDetails.ModelType

	// Set secondary model if multiple strategies exist
	if len(result.PricingStrategies) > 1 {
		result.SecondaryRevenueModel = result.PricingStrategies[1].Type
	}
}

// calculateConfidenceScores calculates confidence scores for all components
func (rma *RevenueModelAnalyzer) calculateConfidenceScores(result *RevenueModelResult) {
	// Calculate component scores
	result.ComponentScores.ModelScore = rma.calculateModelScore(result)
	result.ComponentScores.PricingScore = rma.calculatePricingScore(result)
	result.ComponentScores.StrategyScore = rma.calculateStrategyScore(result)
	result.ComponentScores.MarketScore = rma.calculateMarketScore(result)
	result.ComponentScores.CompetitiveScore = rma.calculateCompetitiveScore(result)

	// Calculate overall confidence
	scores := []float64{
		result.ComponentScores.ModelScore * rma.config.ModelWeight,
		result.ComponentScores.PricingScore * rma.config.PricingWeight,
		result.ComponentScores.StrategyScore * rma.config.StrategyWeight,
		result.ComponentScores.MarketScore * rma.config.MarketWeight,
		result.ComponentScores.CompetitiveScore * rma.config.CompetitiveWeight,
	}

	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	result.ConfidenceScore = totalScore
}

// Helper functions for component scoring
func (rma *RevenueModelAnalyzer) calculateModelScore(result *RevenueModelResult) float64 {
	if result.ModelDetails.ModelType == "" {
		return 0.0
	}
	return 0.8
}

func (rma *RevenueModelAnalyzer) calculatePricingScore(result *RevenueModelResult) float64 {
	if len(result.PricingStrategies) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, strategy := range result.PricingStrategies {
		totalConfidence += strategy.ConfidenceScore
	}

	return totalConfidence / float64(len(result.PricingStrategies))
}

func (rma *RevenueModelAnalyzer) calculateStrategyScore(result *RevenueModelResult) float64 {
	if len(result.RevenueStreams) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, stream := range result.RevenueStreams {
		totalConfidence += stream.ConfidenceScore
	}

	return totalConfidence / float64(len(result.RevenueStreams))
}

func (rma *RevenueModelAnalyzer) calculateMarketScore(result *RevenueModelResult) float64 {
	score := 0.0
	positioning := result.MarketPositioning

	if positioning.MarketSegment != "" {
		score += 0.3
	}
	if len(positioning.CompetitiveAdvantage) > 0 {
		score += 0.3
	}
	if positioning.MarketMaturity != "" {
		score += 0.2
	}
	if positioning.GrowthStrategy != "" {
		score += 0.2
	}

	return score
}

func (rma *RevenueModelAnalyzer) calculateCompetitiveScore(result *RevenueModelResult) float64 {
	score := 0.0
	competitive := result.CompetitiveAnalysis

	if competitive.CompetitiveLandscape != "" {
		score += 0.3
	}
	if len(competitive.Differentiators) > 0 {
		score += 0.3
	}
	if len(competitive.MarketGaps) > 0 {
		score += 0.2
	}
	if competitive.CompetitiveResponse != "" {
		score += 0.2
	}

	return score
}

// validateInput validates input parameters
func (rma *RevenueModelAnalyzer) validateInput(content string) error {
	if len(content) < rma.config.MinContentLength {
		return fmt.Errorf("content too short: %d characters (minimum: %d)", len(content), rma.config.MinContentLength)
	}
	return nil
}

// validateResult validates analysis results
func (rma *RevenueModelAnalyzer) validateResult(result *RevenueModelResult) {
	status := ValidationStatus{
		IsValid:          true,
		ValidationErrors: []string{},
		LastValidated:    time.Now(),
	}

	// Check confidence threshold
	if result.ConfidenceScore < rma.config.MinConfidenceThreshold {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, fmt.Sprintf("Confidence score %.2f below threshold %.2f",
			result.ConfidenceScore, rma.config.MinConfidenceThreshold))
	}

	// Check evidence count
	if len(result.Evidence) < rma.config.MinEvidenceCount {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, fmt.Sprintf("Insufficient evidence: %d items (minimum: %d)",
			len(result.Evidence), rma.config.MinEvidenceCount))
	}

	// Check primary revenue model
	if result.PrimaryRevenueModel == "" || result.PrimaryRevenueModel == "unknown" {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "No primary revenue model identified")
	}

	result.IsValidated = status.IsValid
	result.ValidationStatus = status
}

// calculateDataQuality calculates data quality score
func (rma *RevenueModelAnalyzer) calculateDataQuality(result *RevenueModelResult) float64 {
	qualityFactors := []float64{
		rma.calculateEvidenceQuality(result),
		rma.calculateAnalysisCompleteness(result),
		rma.calculateModelQuality(result),
	}

	total := 0.0
	for _, factor := range qualityFactors {
		total += factor
	}

	return total / float64(len(qualityFactors))
}

func (rma *RevenueModelAnalyzer) calculateEvidenceQuality(result *RevenueModelResult) float64 {
	if len(result.Evidence) == 0 {
		return 0.0
	}

	return minFloat64(float64(len(result.Evidence))*0.1, 1.0)
}

func (rma *RevenueModelAnalyzer) calculateAnalysisCompleteness(result *RevenueModelResult) float64 {
	completeness := 0.0

	if result.ModelDetails.ModelType != "" {
		completeness += 0.3
	}
	if len(result.PricingStrategies) > 0 {
		completeness += 0.3
	}
	if len(result.RevenueStreams) > 0 {
		completeness += 0.2
	}
	if result.MarketPositioning.MarketSegment != "" {
		completeness += 0.2
	}

	return completeness
}

func (rma *RevenueModelAnalyzer) calculateModelQuality(result *RevenueModelResult) float64 {
	if result.ModelDetails.ModelType == "" {
		return 0.0
	}

	quality := 0.0
	if len(result.ModelDetails.RevenueSources) > 0 {
		quality += 0.3
	}
	if result.ModelDetails.ValueProposition != "" {
		quality += 0.3
	}
	if result.ModelDetails.MonetizationMethod != "" {
		quality += 0.2
	}
	if len(result.ModelDetails.CustomerSegments) > 0 {
		quality += 0.2
	}

	return quality
}

// generateReasoning generates human-readable reasoning
func (rma *RevenueModelAnalyzer) generateReasoning(result *RevenueModelResult) string {
	if result.PrimaryRevenueModel == "" {
		return "No clear revenue model indicators found in the analyzed content."
	}

	reasoning := fmt.Sprintf("Primary revenue model identified as '%s' with %.1f%% confidence. ",
		result.PrimaryRevenueModel, result.ConfidenceScore*100)

	if len(result.PricingStrategies) > 0 {
		reasoning += fmt.Sprintf("Pricing strategies include: %v. ", rma.getStrategyNames(result.PricingStrategies))
	}

	if len(result.RevenueStreams) > 0 {
		reasoning += fmt.Sprintf("Revenue streams include: %v. ", rma.getStreamNames(result.RevenueStreams))
	}

	if result.MarketPositioning.MarketSegment != "" {
		reasoning += fmt.Sprintf("Market segment: %s. ", result.MarketPositioning.MarketSegment)
	}

	if len(result.MarketPositioning.CompetitiveAdvantage) > 0 {
		reasoning += fmt.Sprintf("Competitive advantages: %v. ", result.MarketPositioning.CompetitiveAdvantage)
	}

	reasoning += fmt.Sprintf("Analysis based on %d pieces of evidence.", len(result.Evidence))

	return reasoning
}

func (rma *RevenueModelAnalyzer) getStrategyNames(strategies []PricingStrategy) []string {
	names := make([]string, len(strategies))
	for i, strategy := range strategies {
		names[i] = strategy.Name
	}
	return names
}

func (rma *RevenueModelAnalyzer) getStreamNames(streams []RevenueStream) []string {
	names := make([]string, len(streams))
	for i, stream := range streams {
		names[i] = stream.Name
	}
	return names
}
