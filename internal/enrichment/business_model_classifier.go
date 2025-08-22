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

// BusinessModelClassifier provides comprehensive business model classification
type BusinessModelClassifier struct {
	config *BusinessModelClassifierConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// BusinessModelClassifierConfig holds configuration for business model classification
type BusinessModelClassifierConfig struct {
	// Analysis thresholds
	MinConfidenceThreshold float64
	MinEvidenceCount       int
	MinContentLength       int

	// Component weights for classification
	ModelIndicatorWeight   float64
	AudienceAnalysisWeight float64
	RevenueModelWeight     float64
	ConsistencyWeight      float64
	EvidenceWeight         float64

	// Classification settings
	RequireMultipleIndicators bool
	EnableFallbackAnalysis    bool
	ValidateClassifications    bool
	EnableDetailedBreakdown    bool
}

// BusinessModelClassification represents comprehensive business model classification
type BusinessModelClassification struct {
	// Primary classification
	PrimaryBusinessModel   string `json:"primary_business_model"`
	SecondaryBusinessModel string `json:"secondary_business_model,omitempty"`
	BusinessModelType      string `json:"business_model_type"` // B2B, B2C, B2B2C, Marketplace, etc.

	// Detailed breakdown
	ModelIndicators    ModelIndicatorAnalysis    `json:"model_indicators"`
	AudienceAnalysis   AudienceAnalysis          `json:"audience_analysis"`
	RevenueModel       RevenueModelAnalysis      `json:"revenue_model"`
	MarketPositioning  MarketPositioningAnalysis `json:"market_positioning"`

	// Confidence and validation
	ConfidenceScore  float64                    `json:"confidence_score"`
	ComponentScores  BusinessComponentScores    `json:"component_scores"`
	Evidence         []string                   `json:"evidence"`
	ExtractedPhrases []string                   `json:"extracted_phrases"`

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

// ModelIndicatorAnalysis contains business model indicator analysis
type ModelIndicatorAnalysis struct {
	Indicators      []string `json:"indicators"`
	BusinessType    string   `json:"business_type"`
	MarketSegment   string   `json:"market_segment"`
	ConfidenceScore float64  `json:"confidence_score"`
	Evidence        []string `json:"evidence"`
}

// AudienceAnalysis contains target audience analysis
type AudienceAnalysis struct {
	PrimaryAudience   string         `json:"primary_audience"`
	CustomerTypes     []CustomerType `json:"customer_types"`
	Industries        []Industry     `json:"industries"`
	GeographicMarkets []string       `json:"geographic_markets"`
	ConfidenceScore   float64        `json:"confidence_score"`
	Evidence          []string       `json:"evidence"`
}

// RevenueModelAnalysis contains revenue model analysis
type RevenueModelAnalysis struct {
	PrimaryRevenueModel string            `json:"primary_revenue_model"`
	PricingStrategies   []PricingStrategy `json:"pricing_strategies"`
	RevenueStreams      []RevenueStream   `json:"revenue_streams"`
	ConfidenceScore     float64           `json:"confidence_score"`
	Evidence            []string          `json:"evidence"`
}

// MarketPositioningAnalysis contains market positioning analysis
type MarketPositioningAnalysis struct {
	MarketSegment        string   `json:"market_segment"`
	CompetitiveAdvantage []string `json:"competitive_advantage"`
	MarketMaturity       string   `json:"market_maturity"`
	GrowthStrategy       string   `json:"growth_strategy"`
	ConfidenceScore      float64  `json:"confidence_score"`
	Evidence             []string `json:"evidence"`
}

// CustomerType represents a customer type
type CustomerType struct {
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	ConfidenceScore float64 `json:"confidence_score"`
}



// BusinessComponentScores detailed scoring breakdown
type BusinessComponentScores struct {
	ModelIndicatorScore float64 `json:"model_indicator_score"`
	AudienceScore       float64 `json:"audience_score"`
	RevenueScore        float64 `json:"revenue_score"`
	ConsistencyScore    float64 `json:"consistency_score"`
	EvidenceScore       float64 `json:"evidence_score"`
}

// NewBusinessModelClassifier creates a new business model classifier
func NewBusinessModelClassifier(config *BusinessModelClassifierConfig, logger *zap.Logger) *BusinessModelClassifier {
	if config == nil {
		config = GetDefaultBusinessModelClassifierConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &BusinessModelClassifier{
		config: config,
		logger: logger,
		tracer: otel.Tracer("business-model-classifier"),
	}
}

// GetDefaultBusinessModelClassifierConfig returns default configuration
func GetDefaultBusinessModelClassifierConfig() *BusinessModelClassifierConfig {
	return &BusinessModelClassifierConfig{
		MinConfidenceThreshold:    0.3,
		MinEvidenceCount:          3,
		MinContentLength:          100,
		ModelIndicatorWeight:      0.25,
		AudienceAnalysisWeight:    0.30,
		RevenueModelWeight:        0.25,
		ConsistencyWeight:         0.15,
		EvidenceWeight:            0.05,
		RequireMultipleIndicators: true,
		EnableFallbackAnalysis:    true,
		ValidateClassifications:   true,
		EnableDetailedBreakdown:   true,
	}
}

// ClassifyBusinessModel performs comprehensive business model classification
func (bmc *BusinessModelClassifier) ClassifyBusinessModel(ctx context.Context, content, sourceURL string) (*BusinessModelClassification, error) {
	ctx, span := bmc.tracer.Start(ctx, "business_model_classifier.classify_business_model")
	defer span.End()

	startTime := time.Now()

	// Input validation
	if err := bmc.validateInput(content); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	result := &BusinessModelClassification{
		SourceURL:        sourceURL,
		AnalyzedAt:       time.Now(),
		Evidence:         []string{},
		ExtractedPhrases: []string{},
		ComponentScores:  BusinessComponentScores{},
	}

	// Perform component analyses
	if err := bmc.analyzeModelIndicators(ctx, content, result); err != nil {
		bmc.logger.Warn("Model indicator analysis failed", zap.Error(err))
	}

	if err := bmc.analyzeAudience(ctx, content, result); err != nil {
		bmc.logger.Warn("Audience analysis failed", zap.Error(err))
	}

	if err := bmc.analyzeRevenueModel(ctx, content, result); err != nil {
		bmc.logger.Warn("Revenue model analysis failed", zap.Error(err))
	}

	if err := bmc.analyzeMarketPositioning(ctx, content, result); err != nil {
		bmc.logger.Warn("Market positioning analysis failed", zap.Error(err))
	}

	// Determine primary and secondary business models
	bmc.determineBusinessModels(result)

	// Calculate confidence scores
	bmc.calculateConfidenceScores(result)

	// Validate results
	bmc.validateResult(result)

	// Generate reasoning
	result.Reasoning = bmc.generateReasoning(result)

	// Calculate data quality
	result.DataQualityScore = bmc.calculateDataQuality(result)

	// Record processing time
	result.ProcessingTime = time.Since(startTime)

	// Log completion
	bmc.logger.Info("Business model classification completed",
		zap.String("primary_model", result.PrimaryBusinessModel),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// validateInput validates input parameters
func (bmc *BusinessModelClassifier) validateInput(content string) error {
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if len(content) < bmc.config.MinContentLength {
		return fmt.Errorf("content too short: %d characters (minimum: %d)", len(content), bmc.config.MinContentLength)
	}

	return nil
}

// analyzeModelIndicators analyzes business model indicators
func (bmc *BusinessModelClassifier) analyzeModelIndicators(ctx context.Context, content string, result *BusinessModelClassification) error {
	ctx, span := bmc.tracer.Start(ctx, "business_model_classifier.analyze_model_indicators")
	defer span.End()

	contentLower := strings.ToLower(content)
	indicators := []string{}
	evidence := []string{}

	// B2B indicators
	if strings.Contains(contentLower, "enterprise") || strings.Contains(contentLower, "business") || strings.Contains(contentLower, "b2b") {
		indicators = append(indicators, "B2B")
		evidence = append(evidence, "B2B indicators found in content")
	}

	// B2C indicators
	if strings.Contains(contentLower, "consumer") || strings.Contains(contentLower, "individual") || strings.Contains(contentLower, "personal") {
		indicators = append(indicators, "B2C")
		evidence = append(evidence, "B2C indicators found in content")
	}

	// Marketplace indicators
	if strings.Contains(contentLower, "marketplace") || strings.Contains(contentLower, "platform") || strings.Contains(contentLower, "connect") {
		indicators = append(indicators, "Marketplace")
		evidence = append(evidence, "Marketplace indicators found in content")
	}

	// SaaS indicators
	if strings.Contains(contentLower, "saas") || strings.Contains(contentLower, "software as a service") || strings.Contains(contentLower, "subscription") {
		indicators = append(indicators, "SaaS")
		evidence = append(evidence, "SaaS indicators found in content")
	}

	// E-commerce indicators
	if strings.Contains(contentLower, "e-commerce") || strings.Contains(contentLower, "online store") || strings.Contains(contentLower, "shopping") {
		indicators = append(indicators, "E-commerce")
		evidence = append(evidence, "E-commerce indicators found in content")
	}

	// Determine business type
	businessType := bmc.determineBusinessType(indicators)

	result.ModelIndicators = ModelIndicatorAnalysis{
		Indicators:      indicators,
		BusinessType:    businessType,
		MarketSegment:   bmc.determineMarketSegment(indicators),
		ConfidenceScore: bmc.calculateIndicatorConfidence(indicators, evidence),
		Evidence:        evidence,
	}

	result.Evidence = append(result.Evidence, evidence...)
	result.ExtractedPhrases = append(result.ExtractedPhrases, indicators...)

	return nil
}

// analyzeAudience analyzes target audience
func (bmc *BusinessModelClassifier) analyzeAudience(ctx context.Context, content string, result *BusinessModelClassification) error {
	ctx, span := bmc.tracer.Start(ctx, "business_model_classifier.analyze_audience")
	defer span.End()

	contentLower := strings.ToLower(content)
	evidence := []string{}

	// Analyze customer types
	customerTypes := bmc.analyzeCustomerTypes(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d customer types", len(customerTypes)))

	// Analyze industries
	industries := bmc.analyzeIndustries(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d target industries", len(industries)))

	// Analyze geographic markets
	geographicMarkets := bmc.analyzeGeographicMarkets(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d geographic markets", len(geographicMarkets)))

	// Determine primary audience
	primaryAudience := bmc.determinePrimaryAudience(customerTypes, industries)

	result.AudienceAnalysis = AudienceAnalysis{
		PrimaryAudience:   primaryAudience,
		CustomerTypes:     customerTypes,
		Industries:        industries,
		GeographicMarkets: geographicMarkets,
		ConfidenceScore:   bmc.calculateAudienceConfidence(customerTypes, industries, geographicMarkets),
		Evidence:          evidence,
	}

	result.Evidence = append(result.Evidence, evidence...)

	return nil
}

// analyzeRevenueModel analyzes revenue model
func (bmc *BusinessModelClassifier) analyzeRevenueModel(ctx context.Context, content string, result *BusinessModelClassification) error {
	ctx, span := bmc.tracer.Start(ctx, "business_model_classifier.analyze_revenue_model")
	defer span.End()

	contentLower := strings.ToLower(content)
	evidence := []string{}

	// Analyze revenue models
	revenueModels := bmc.analyzeRevenueModels(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d revenue models", len(revenueModels)))

	// Analyze pricing strategies
	pricingStrategies := bmc.analyzePricingStrategies(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d pricing strategies", len(pricingStrategies)))

	// Analyze revenue streams
	revenueStreams := bmc.analyzeRevenueStreams(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d revenue streams", len(revenueStreams)))

	// Determine primary revenue model
	primaryRevenueModel := bmc.determinePrimaryRevenueModel(revenueModels)

	result.RevenueModel = RevenueModelAnalysis{
		PrimaryRevenueModel: primaryRevenueModel,
		PricingStrategies:   pricingStrategies,
		RevenueStreams:      revenueStreams,
		ConfidenceScore:     bmc.calculateRevenueConfidence(revenueModels, pricingStrategies, revenueStreams),
		Evidence:            evidence,
	}

	result.Evidence = append(result.Evidence, evidence...)

	return nil
}

// analyzeMarketPositioning analyzes market positioning
func (bmc *BusinessModelClassifier) analyzeMarketPositioning(ctx context.Context, content string, result *BusinessModelClassification) error {
	ctx, span := bmc.tracer.Start(ctx, "business_model_classifier.analyze_market_positioning")
	defer span.End()

	contentLower := strings.ToLower(content)
	evidence := []string{}

	// Analyze market segment
	marketSegment := bmc.analyzeMarketSegment(contentLower)
	evidence = append(evidence, fmt.Sprintf("Market segment: %s", marketSegment))

	// Analyze competitive advantages
	competitiveAdvantages := bmc.analyzeCompetitiveAdvantages(contentLower)
	evidence = append(evidence, fmt.Sprintf("Identified %d competitive advantages", len(competitiveAdvantages)))

	// Analyze market maturity
	marketMaturity := bmc.analyzeMarketMaturity(contentLower)
	evidence = append(evidence, fmt.Sprintf("Market maturity: %s", marketMaturity))

	// Analyze growth strategy
	growthStrategy := bmc.analyzeGrowthStrategy(contentLower)
	evidence = append(evidence, fmt.Sprintf("Growth strategy: %s", growthStrategy))

	result.MarketPositioning = MarketPositioningAnalysis{
		MarketSegment:        marketSegment,
		CompetitiveAdvantage: competitiveAdvantages,
		MarketMaturity:       marketMaturity,
		GrowthStrategy:       growthStrategy,
		ConfidenceScore:      bmc.calculateMarketPositioningConfidence(marketSegment, competitiveAdvantages, marketMaturity),
		Evidence:             evidence,
	}

	result.Evidence = append(result.Evidence, evidence...)

	return nil
}

// determineBusinessModels determines primary and secondary business models
func (bmc *BusinessModelClassifier) determineBusinessModels(result *BusinessModelClassification) {
	// Combine indicators from all analyses
	models := []string{}

	// Add model indicators
	models = append(models, result.ModelIndicators.Indicators...)

	// Add audience-based models
	if result.AudienceAnalysis.PrimaryAudience != "" {
		models = append(models, result.AudienceAnalysis.PrimaryAudience)
	}

	// Add revenue model
	if result.RevenueModel.PrimaryRevenueModel != "" {
		models = append(models, result.RevenueModel.PrimaryRevenueModel)
	}

	// Determine primary and secondary models
	if len(models) > 0 {
		result.PrimaryBusinessModel = models[0]
		if len(models) > 1 {
			result.SecondaryBusinessModel = models[1]
		}
	}

	// Determine business model type
	result.BusinessModelType = bmc.determineBusinessModelType(result)
}

// determineBusinessModelType determines the overall business model type
func (bmc *BusinessModelClassifier) determineBusinessModelType(result *BusinessModelClassification) string {
	// Analyze patterns across all components
	indicators := []string{}

	// Model indicators
	indicators = append(indicators, result.ModelIndicators.BusinessType)

	// Audience analysis
	if result.AudienceAnalysis.PrimaryAudience != "" {
		indicators = append(indicators, result.AudienceAnalysis.PrimaryAudience)
	}

	// Revenue model
	if result.RevenueModel.PrimaryRevenueModel != "" {
		indicators = append(indicators, result.RevenueModel.PrimaryRevenueModel)
	}

	// Determine type based on patterns
	if bmc.containsAll(indicators, []string{"B2B", "B2C"}) {
		return "B2B2C"
	} else if bmc.containsAny(indicators, []string{"Marketplace", "Platform"}) {
		return "Marketplace"
	} else if bmc.containsAny(indicators, []string{"B2B", "Enterprise"}) {
		return "B2B"
	} else if bmc.containsAny(indicators, []string{"B2C", "Consumer"}) {
		return "B2C"
	} else if bmc.containsAny(indicators, []string{"SaaS", "Subscription"}) {
		return "SaaS"
	} else {
		return "Unknown"
	}
}

// calculateConfidenceScores calculates confidence scores for all components
func (bmc *BusinessModelClassifier) calculateConfidenceScores(result *BusinessModelClassification) {
	// Calculate component scores
	result.ComponentScores.ModelIndicatorScore = result.ModelIndicators.ConfidenceScore
	result.ComponentScores.AudienceScore = result.AudienceAnalysis.ConfidenceScore
	result.ComponentScores.RevenueScore = result.RevenueModel.ConfidenceScore
	result.ComponentScores.ConsistencyScore = bmc.calculateConsistencyScore(result)
	result.ComponentScores.EvidenceScore = bmc.calculateEvidenceScore(result)

	// Calculate overall confidence
	scores := []float64{
		result.ComponentScores.ModelIndicatorScore * bmc.config.ModelIndicatorWeight,
		result.ComponentScores.AudienceScore * bmc.config.AudienceAnalysisWeight,
		result.ComponentScores.RevenueScore * bmc.config.RevenueModelWeight,
		result.ComponentScores.ConsistencyScore * bmc.config.ConsistencyWeight,
		result.ComponentScores.EvidenceScore * bmc.config.EvidenceWeight,
	}

	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	result.ConfidenceScore = totalScore
}

// validateResult validates the classification result
func (bmc *BusinessModelClassifier) validateResult(result *BusinessModelClassification) {
	validationErrors := []string{}

	// Check primary business model
	if result.PrimaryBusinessModel == "" {
		validationErrors = append(validationErrors, "No primary business model identified")
	}

	// Check confidence score
	if result.ConfidenceScore < bmc.config.MinConfidenceThreshold {
		validationErrors = append(validationErrors, fmt.Sprintf("Confidence score too low: %.2f (minimum: %.2f)", result.ConfidenceScore, bmc.config.MinConfidenceThreshold))
	}

	// Check evidence count
	if len(result.Evidence) < bmc.config.MinEvidenceCount {
		validationErrors = append(validationErrors, fmt.Sprintf("Insufficient evidence: %d (minimum: %d)", len(result.Evidence), bmc.config.MinEvidenceCount))
	}

	// Set validation status
	result.ValidationStatus = ValidationStatus{
		IsValid:          len(validationErrors) == 0,
		ValidationErrors: validationErrors,
		LastValidated:    time.Now(),
	}

	result.IsValidated = result.ValidationStatus.IsValid
}

// generateReasoning generates human-readable reasoning
func (bmc *BusinessModelClassifier) generateReasoning(result *BusinessModelClassification) string {
	if result.PrimaryBusinessModel == "" {
		return "No clear business model indicators found in the analyzed content."
	}

	reasoning := fmt.Sprintf("Primary business model identified as '%s' (%s) with %.1f%% confidence. ",
		result.PrimaryBusinessModel, result.BusinessModelType, result.ConfidenceScore*100)

	// Add model indicators
	if len(result.ModelIndicators.Indicators) > 0 {
		reasoning += fmt.Sprintf("Model indicators: %v. ", result.ModelIndicators.Indicators)
	}

	// Add audience analysis
	if result.AudienceAnalysis.PrimaryAudience != "" {
		reasoning += fmt.Sprintf("Primary audience: %s. ", result.AudienceAnalysis.PrimaryAudience)
	}

	// Add revenue model
	if result.RevenueModel.PrimaryRevenueModel != "" {
		reasoning += fmt.Sprintf("Revenue model: %s. ", result.RevenueModel.PrimaryRevenueModel)
	}

	// Add market positioning
	if result.MarketPositioning.MarketSegment != "" {
		reasoning += fmt.Sprintf("Market segment: %s. ", result.MarketPositioning.MarketSegment)
	}

	reasoning += fmt.Sprintf("Analysis based on %d pieces of evidence.", len(result.Evidence))

	return reasoning
}

// calculateDataQuality calculates data quality score
func (bmc *BusinessModelClassifier) calculateDataQuality(result *BusinessModelClassification) float64 {
	qualityFactors := []float64{
		bmc.calculateCompletenessScore(result),
		bmc.calculateConsistencyScore(result),
		bmc.calculateEvidenceQuality(result),
		bmc.calculateAnalysisCompleteness(result),
	}

	total := 0.0
	for _, factor := range qualityFactors {
		total += factor
	}

	return total / float64(len(qualityFactors))
}

// Helper functions for analysis components
func (bmc *BusinessModelClassifier) determineBusinessType(indicators []string) string {
	if bmc.containsAny(indicators, []string{"B2B", "Enterprise"}) {
		return "B2B"
	} else if bmc.containsAny(indicators, []string{"B2C", "Consumer"}) {
		return "B2C"
	} else if bmc.containsAny(indicators, []string{"Marketplace", "Platform"}) {
		return "Marketplace"
	} else if bmc.containsAny(indicators, []string{"SaaS"}) {
		return "SaaS"
	} else {
		return "Unknown"
	}
}

func (bmc *BusinessModelClassifier) determineMarketSegment(indicators []string) string {
	if bmc.containsAny(indicators, []string{"Enterprise", "B2B"}) {
		return "Enterprise"
	} else if bmc.containsAny(indicators, []string{"SMB", "Small Business"}) {
		return "SMB"
	} else if bmc.containsAny(indicators, []string{"Consumer", "B2C"}) {
		return "Consumer"
	} else {
		return "Mixed"
	}
}

func (bmc *BusinessModelClassifier) calculateIndicatorConfidence(indicators []string, evidence []string) float64 {
	if len(indicators) == 0 {
		return 0.0
	}

	baseConfidence := float64(len(indicators)) * 0.2
	evidenceBonus := float64(len(evidence)) * 0.1

	return minFloat64(baseConfidence+evidenceBonus, 1.0)
}

func (bmc *BusinessModelClassifier) analyzeCustomerTypes(content string) []CustomerType {
	// Simplified customer type analysis
	customerTypes := []CustomerType{}

	if strings.Contains(content, "enterprise") {
		customerTypes = append(customerTypes, CustomerType{
			Type:            "Enterprise",
			Description:     "Large enterprise customers",
			ConfidenceScore: 0.8,
		})
	}

	if strings.Contains(content, "small business") || strings.Contains(content, "smb") {
		customerTypes = append(customerTypes, CustomerType{
			Type:            "SMB",
			Description:     "Small and medium businesses",
			ConfidenceScore: 0.7,
		})
	}

	if strings.Contains(content, "consumer") || strings.Contains(content, "individual") {
		customerTypes = append(customerTypes, CustomerType{
			Type:            "Consumer",
			Description:     "Individual consumers",
			ConfidenceScore: 0.8,
		})
	}

	return customerTypes
}

func (bmc *BusinessModelClassifier) analyzeIndustries(content string) []Industry {
	// Simplified industry analysis
	industries := []Industry{}

	if strings.Contains(content, "technology") || strings.Contains(content, "software") {
		industries = append(industries, Industry{
			Name:            "Technology",
			Sector:          "Technology",
			ConfidenceScore: 0.8,
		})
	}

	if strings.Contains(content, "finance") || strings.Contains(content, "banking") {
		industries = append(industries, Industry{
			Name:            "Financial Services",
			Sector:          "Finance",
			ConfidenceScore: 0.8,
		})
	}

	if strings.Contains(content, "healthcare") || strings.Contains(content, "medical") {
		industries = append(industries, Industry{
			Name:            "Healthcare",
			Sector:          "Healthcare",
			ConfidenceScore: 0.8,
		})
	}

	return industries
}

func (bmc *BusinessModelClassifier) analyzeGeographicMarkets(content string) []string {
	markets := []string{}

	if strings.Contains(content, "global") || strings.Contains(content, "worldwide") {
		markets = append(markets, "Global")
	}

	if strings.Contains(content, "north america") || strings.Contains(content, "united states") {
		markets = append(markets, "North America")
	}

	if strings.Contains(content, "europe") || strings.Contains(content, "european") {
		markets = append(markets, "Europe")
	}

	return markets
}

func (bmc *BusinessModelClassifier) determinePrimaryAudience(customerTypes []CustomerType, industries []Industry) string {
	if len(customerTypes) == 0 {
		return ""
	}

	// Return the first customer type as primary
	return customerTypes[0].Type
}

func (bmc *BusinessModelClassifier) calculateAudienceConfidence(customerTypes []CustomerType, industries []Industry, markets []string) float64 {
	score := 0.0

	if len(customerTypes) > 0 {
		score += 0.4
	}
	if len(industries) > 0 {
		score += 0.3
	}
	if len(markets) > 0 {
		score += 0.3
	}

	return score
}

func (bmc *BusinessModelClassifier) analyzeRevenueModels(content string) []string {
	models := []string{}

	if strings.Contains(content, "subscription") {
		models = append(models, "subscription")
	}
	if strings.Contains(content, "freemium") {
		models = append(models, "freemium")
	}
	if strings.Contains(content, "marketplace") {
		models = append(models, "marketplace")
	}
	if strings.Contains(content, "enterprise") {
		models = append(models, "enterprise")
	}
	if strings.Contains(content, "advertising") {
		models = append(models, "advertising")
	}

	return models
}

func (bmc *BusinessModelClassifier) analyzePricingStrategies(content string) []PricingStrategy {
	strategies := []PricingStrategy{}

	if strings.Contains(content, "tier") || strings.Contains(content, "plan") {
		strategies = append(strategies, PricingStrategy{
			Name:            "Tiered Pricing",
			Type:            "tiered",
			ConfidenceScore: 0.8,
		})
	}

	if strings.Contains(content, "premium") {
		strategies = append(strategies, PricingStrategy{
			Name:            "Premium Pricing",
			Type:            "premium",
			ConfidenceScore: 0.8,
		})
	}

	return strategies
}

func (bmc *BusinessModelClassifier) analyzeRevenueStreams(content string) []RevenueStream {
	streams := []RevenueStream{}

	if strings.Contains(content, "license") || strings.Contains(content, "software") {
		streams = append(streams, RevenueStream{
			Name:            "Software Licensing",
			Type:            "licensing",
			ConfidenceScore: 0.8,
		})
	}

	if strings.Contains(content, "transaction") || strings.Contains(content, "commission") {
		streams = append(streams, RevenueStream{
			Name:            "Transaction Fees",
			Type:            "transaction",
			ConfidenceScore: 0.7,
		})
	}

	return streams
}

func (bmc *BusinessModelClassifier) determinePrimaryRevenueModel(models []string) string {
	if len(models) == 0 {
		return ""
	}
	return models[0]
}

func (bmc *BusinessModelClassifier) calculateRevenueConfidence(models []string, strategies []PricingStrategy, streams []RevenueStream) float64 {
	score := 0.0

	if len(models) > 0 {
		score += 0.4
	}
	if len(strategies) > 0 {
		score += 0.3
	}
	if len(streams) > 0 {
		score += 0.3
	}

	return score
}

func (bmc *BusinessModelClassifier) analyzeMarketSegment(content string) string {
	if strings.Contains(content, "enterprise") {
		return "Enterprise"
	} else if strings.Contains(content, "small business") || strings.Contains(content, "smb") {
		return "SMB"
	} else if strings.Contains(content, "consumer") {
		return "Consumer"
	} else {
		return "Mixed"
	}
}

func (bmc *BusinessModelClassifier) analyzeCompetitiveAdvantages(content string) []string {
	advantages := []string{}

	if strings.Contains(content, "innovation") {
		advantages = append(advantages, "Innovation")
	}
	if strings.Contains(content, "quality") {
		advantages = append(advantages, "Quality")
	}
	if strings.Contains(content, "price") || strings.Contains(content, "affordable") {
		advantages = append(advantages, "Price")
	}

	return advantages
}

func (bmc *BusinessModelClassifier) analyzeMarketMaturity(content string) string {
	if strings.Contains(content, "emerging") {
		return "Emerging"
	} else if strings.Contains(content, "mature") {
		return "Mature"
	} else if strings.Contains(content, "growing") {
		return "Growing"
	} else {
		return "Unknown"
	}
}

func (bmc *BusinessModelClassifier) analyzeGrowthStrategy(content string) string {
	if strings.Contains(content, "expansion") {
		return "Market Expansion"
	} else if strings.Contains(content, "acquisition") {
		return "Acquisition"
	} else if strings.Contains(content, "innovation") {
		return "Innovation"
	} else {
		return "Organic Growth"
	}
}

func (bmc *BusinessModelClassifier) calculateMarketPositioningConfidence(segment string, advantages []string, maturity string) float64 {
	score := 0.0

	if segment != "" {
		score += 0.4
	}
	if len(advantages) > 0 {
		score += 0.3
	}
	if maturity != "" {
		score += 0.3
	}

	return score
}

func (bmc *BusinessModelClassifier) calculateConsistencyScore(result *BusinessModelClassification) float64 {
	// Check consistency across different analyses
	consistency := 0.0
	checks := 0

	// Check if model indicators align with audience analysis
	if result.ModelIndicators.BusinessType != "" && result.AudienceAnalysis.PrimaryAudience != "" {
		if bmc.areConsistent(result.ModelIndicators.BusinessType, result.AudienceAnalysis.PrimaryAudience) {
			consistency += 1.0
		}
		checks++
	}

	// Check if revenue model aligns with business type
	if result.RevenueModel.PrimaryRevenueModel != "" && result.ModelIndicators.BusinessType != "" {
		if bmc.areConsistent(result.RevenueModel.PrimaryRevenueModel, result.ModelIndicators.BusinessType) {
			consistency += 1.0
		}
		checks++
	}

	if checks == 0 {
		return 0.0
	}

	return consistency / float64(checks)
}

func (bmc *BusinessModelClassifier) calculateEvidenceScore(result *BusinessModelClassification) float64 {
	if len(result.Evidence) == 0 {
		return 0.0
	}

	return minFloat64(float64(len(result.Evidence))*0.1, 1.0)
}

func (bmc *BusinessModelClassifier) calculateCompletenessScore(result *BusinessModelClassification) float64 {
	completeness := 0.0

	if result.ModelIndicators.BusinessType != "" {
		completeness += 0.25
	}
	if result.AudienceAnalysis.PrimaryAudience != "" {
		completeness += 0.25
	}
	if result.RevenueModel.PrimaryRevenueModel != "" {
		completeness += 0.25
	}
	if result.MarketPositioning.MarketSegment != "" {
		completeness += 0.25
	}

	return completeness
}

func (bmc *BusinessModelClassifier) calculateAnalysisCompleteness(result *BusinessModelClassification) float64 {
	completeness := 0.0

	if len(result.ModelIndicators.Indicators) > 0 {
		completeness += 0.25
	}
	if len(result.AudienceAnalysis.CustomerTypes) > 0 {
		completeness += 0.25
	}
	if len(result.RevenueModel.PricingStrategies) > 0 {
		completeness += 0.25
	}
	if len(result.MarketPositioning.CompetitiveAdvantage) > 0 {
		completeness += 0.25
	}

	return completeness
}

func (bmc *BusinessModelClassifier) calculateEvidenceQuality(result *BusinessModelClassification) float64 {
	if len(result.Evidence) == 0 {
		return 0.0
	}

	return minFloat64(float64(len(result.Evidence))*0.1, 1.0)
}

// Utility functions
func (bmc *BusinessModelClassifier) containsAny(slice []string, items []string) bool {
	for _, item := range items {
		for _, s := range slice {
			if strings.Contains(strings.ToLower(s), strings.ToLower(item)) {
				return true
			}
		}
	}
	return false
}

func (bmc *BusinessModelClassifier) containsAll(slice []string, items []string) bool {
	for _, item := range items {
		found := false
		for _, s := range slice {
			if strings.Contains(strings.ToLower(s), strings.ToLower(item)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (bmc *BusinessModelClassifier) areConsistent(item1, item2 string) bool {
	// Simple consistency check - can be enhanced
	item1Lower := strings.ToLower(item1)
	item2Lower := strings.ToLower(item2)

	// B2B consistency
	if (strings.Contains(item1Lower, "b2b") || strings.Contains(item1Lower, "enterprise")) &&
		(strings.Contains(item2Lower, "b2b") || strings.Contains(item2Lower, "enterprise")) {
		return true
	}

	// B2C consistency
	if (strings.Contains(item1Lower, "b2c") || strings.Contains(item1Lower, "consumer")) &&
		(strings.Contains(item2Lower, "b2c") || strings.Contains(item2Lower, "consumer")) {
		return true
	}

	// Marketplace consistency
	if (strings.Contains(item1Lower, "marketplace") || strings.Contains(item1Lower, "platform")) &&
		(strings.Contains(item2Lower, "marketplace") || strings.Contains(item2Lower, "platform")) {
		return true
	}

	return false
}
