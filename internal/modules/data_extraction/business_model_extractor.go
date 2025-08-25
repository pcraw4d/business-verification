package data_extraction

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BusinessModelExtractor extracts business model information from business data
type BusinessModelExtractor struct {
	// Configuration
	config *BusinessModelConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Pattern matching
	b2bPatterns          []*regexp.Regexp
	b2cPatterns          []*regexp.Regexp
	b2b2cPatterns        []*regexp.Regexp
	marketplacePatterns  []*regexp.Regexp
	saasPatterns         []*regexp.Regexp
	subscriptionPatterns []*regexp.Regexp
	oneTimePatterns      []*regexp.Regexp
	freemiumPatterns     []*regexp.Regexp
	enterprisePatterns   []*regexp.Regexp
	consumerPatterns     []*regexp.Regexp
	pricingPatterns      []*regexp.Regexp
}

// BusinessModelConfig holds configuration for the business model extractor
type BusinessModelConfig struct {
	// Pattern matching settings
	CaseSensitive bool
	MaxPatterns   int

	// Confidence scoring settings
	MinConfidenceThreshold float64
	MaxConfidenceThreshold float64

	// ML model settings
	EnableMLModel bool
	MLModelPath   string

	// Processing settings
	Timeout time.Duration
}

// BusinessModel represents extracted business model information
type BusinessModel struct {
	// Business model type
	BusinessModelType string  `json:"business_model_type"` // B2B, B2C, B2B2C, Marketplace, SaaS
	ModelConfidence   float64 `json:"model_confidence"`

	// Revenue model
	RevenueModel      string  `json:"revenue_model"` // subscription, one-time, freemium, etc.
	RevenueConfidence float64 `json:"revenue_confidence"`

	// Target market
	TargetMarket     string  `json:"target_market"` // enterprise, consumer, both, etc.
	MarketConfidence float64 `json:"market_confidence"`

	// Pricing model
	PricingModel      string  `json:"pricing_model"` // tiered, usage-based, flat-rate, etc.
	PricingConfidence float64 `json:"pricing_confidence"`

	// Additional details
	ModelDetails       map[string]interface{} `json:"model_details,omitempty"`
	SupportingEvidence []string               `json:"supporting_evidence,omitempty"`

	// Overall assessment
	OverallConfidence float64 `json:"overall_confidence"`

	// Metadata
	ExtractedAt time.Time `json:"extracted_at"`
	DataSources []string  `json:"data_sources"`
}

// Business model type constants
const (
	BusinessModelB2B           = "B2B"
	BusinessModelB2C           = "B2C"
	BusinessModelB2B2C         = "B2B2C"
	BusinessModelMarketplace   = "Marketplace"
	BusinessModelSaaS          = "SaaS"
	BusinessModelEcommerce     = "E-commerce"
	BusinessModelConsulting    = "Consulting"
	BusinessModelAgency        = "Agency"
	BusinessModelManufacturing = "Manufacturing"
	BusinessModelRetail        = "Retail"
)

// Revenue model constants
const (
	RevenueModelSubscription = "subscription"
	RevenueModelOneTime      = "one-time"
	RevenueModelFreemium     = "freemium"
	RevenueModelUsageBased   = "usage-based"
	RevenueModelTiered       = "tiered"
	RevenueModelCommission   = "commission"
	RevenueModelAdvertising  = "advertising"
	RevenueModelLicensing    = "licensing"
	RevenueModelServices     = "services"
)

// Target market constants
const (
	TargetMarketEnterprise = "enterprise"
	TargetMarketConsumer   = "consumer"
	TargetMarketBoth       = "both"
	TargetMarketSMB        = "small_medium_business"
	TargetMarketStartup    = "startup"
	TargetMarketGovernment = "government"
	TargetMarketEducation  = "education"
)

// Pricing model constants
const (
	PricingModelTiered       = "tiered"
	PricingModelUsageBased   = "usage-based"
	PricingModelFlatRate     = "flat-rate"
	PricingModelPerUser      = "per-user"
	PricingModelPerFeature   = "per-feature"
	PricingModelFreemium     = "freemium"
	PricingModelPayPerUse    = "pay-per-use"
	PricingModelSubscription = "subscription"
)

// NewBusinessModelExtractor creates a new business model extractor
func NewBusinessModelExtractor(
	config *BusinessModelConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *BusinessModelExtractor {
	// Set default configuration
	if config == nil {
		config = &BusinessModelConfig{
			CaseSensitive:          false,
			MaxPatterns:            100,
			MinConfidenceThreshold: 0.3,
			MaxConfidenceThreshold: 1.0,
			EnableMLModel:          false,
			Timeout:                30 * time.Second,
		}
	}

	extractor := &BusinessModelExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize pattern matching
	extractor.initializePatterns()

	return extractor
}

// initializePatterns initializes all pattern matching regexes
func (bme *BusinessModelExtractor) initializePatterns() {
	// B2B patterns
	bme.b2bPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)business\s+to\s+business`),
		regexp.MustCompile(`(?i)b2b`),
		regexp.MustCompile(`(?i)enterprise\s+solution`),
		regexp.MustCompile(`(?i)corporate\s+client`),
		regexp.MustCompile(`(?i)business\s+client`),
		regexp.MustCompile(`(?i)enterprise\s+software`),
		regexp.MustCompile(`(?i)business\s+software`),
		regexp.MustCompile(`(?i)corporate\s+software`),
		regexp.MustCompile(`(?i)business\s+service`),
		regexp.MustCompile(`(?i)enterprise\s+service`),
		regexp.MustCompile(`(?i)business\s+platform`),
		regexp.MustCompile(`(?i)enterprise\s+platform`),
	}

	// B2C patterns
	bme.b2cPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)business\s+to\s+consumer`),
		regexp.MustCompile(`(?i)b2c`),
		regexp.MustCompile(`(?i)consumer\s+app`),
		regexp.MustCompile(`(?i)consumer\s+product`),
		regexp.MustCompile(`(?i)consumer\s+service`),
		regexp.MustCompile(`(?i)retail\s+customer`),
		regexp.MustCompile(`(?i)individual\s+customer`),
		regexp.MustCompile(`(?i)personal\s+use`),
		regexp.MustCompile(`(?i)consumer\s+market`),
		regexp.MustCompile(`(?i)end\s+user`),
	}

	// B2B2C patterns
	bme.b2b2cPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)business\s+to\s+business\s+to\s+consumer`),
		regexp.MustCompile(`(?i)b2b2c`),
		regexp.MustCompile(`(?i)platform\s+for\s+businesses`),
		regexp.MustCompile(`(?i)marketplace\s+for\s+businesses`),
		regexp.MustCompile(`(?i)business\s+marketplace`),
		regexp.MustCompile(`(?i)business\s+platform`),
	}

	// Marketplace patterns
	bme.marketplacePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)marketplace`),
		regexp.MustCompile(`(?i)platform\s+connecting`),
		regexp.MustCompile(`(?i)connecting\s+buyers\s+and\s+sellers`),
		regexp.MustCompile(`(?i)peer\s+to\s+peer`),
		regexp.MustCompile(`(?i)p2p`),
		regexp.MustCompile(`(?i)multi-sided\s+platform`),
		regexp.MustCompile(`(?i)two-sided\s+marketplace`),
		regexp.MustCompile(`(?i)exchange\s+platform`),
		regexp.MustCompile(`(?i)brokerage`),
		regexp.MustCompile(`(?i)intermediary`),
	}

	// SaaS patterns
	bme.saasPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)software\s+as\s+a\s+service`),
		regexp.MustCompile(`(?i)saas`),
		regexp.MustCompile(`(?i)cloud\s+software`),
		regexp.MustCompile(`(?i)web-based\s+software`),
		regexp.MustCompile(`(?i)online\s+software`),
		regexp.MustCompile(`(?i)subscription\s+software`),
		regexp.MustCompile(`(?i)cloud-based\s+platform`),
		regexp.MustCompile(`(?i)web\s+application`),
		regexp.MustCompile(`(?i)online\s+platform`),
		regexp.MustCompile(`(?i)cloud\s+platform`),
	}

	// Subscription patterns
	bme.subscriptionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)subscription`),
		regexp.MustCompile(`(?i)monthly\s+plan`),
		regexp.MustCompile(`(?i)annual\s+plan`),
		regexp.MustCompile(`(?i)recurring\s+revenue`),
		regexp.MustCompile(`(?i)monthly\s+subscription`),
		regexp.MustCompile(`(?i)yearly\s+subscription`),
		regexp.MustCompile(`(?i)subscription\s+model`),
		regexp.MustCompile(`(?i)subscription-based`),
	}

	// One-time patterns
	bme.oneTimePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)one-time\s+purchase`),
		regexp.MustCompile(`(?i)one\s+time\s+purchase`),
		regexp.MustCompile(`(?i)single\s+purchase`),
		regexp.MustCompile(`(?i)perpetual\s+license`),
		regexp.MustCompile(`(?i)lifetime\s+license`),
		regexp.MustCompile(`(?i)one-time\s+payment`),
		regexp.MustCompile(`(?i)upfront\s+payment`),
		regexp.MustCompile(`(?i)license\s+purchase`),
	}

	// Freemium patterns
	bme.freemiumPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)freemium`),
		regexp.MustCompile(`(?i)free\s+tier`),
		regexp.MustCompile(`(?i)free\s+plan`),
		regexp.MustCompile(`(?i)free\s+version`),
		regexp.MustCompile(`(?i)basic\s+plan`),
		regexp.MustCompile(`(?i)premium\s+plan`),
		regexp.MustCompile(`(?i)pro\s+plan`),
		regexp.MustCompile(`(?i)upgrade\s+to\s+premium`),
		regexp.MustCompile(`(?i)free\s+and\s+paid`),
	}

	// Enterprise patterns
	bme.enterprisePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)enterprise`),
		regexp.MustCompile(`(?i)corporate`),
		regexp.MustCompile(`(?i)business\s+client`),
		regexp.MustCompile(`(?i)large\s+organization`),
		regexp.MustCompile(`(?i)fortune\s+500`),
		regexp.MustCompile(`(?i)enterprise\s+grade`),
		regexp.MustCompile(`(?i)enterprise\s+level`),
		regexp.MustCompile(`(?i)business\s+to\s+business`),
		regexp.MustCompile(`(?i)b2b`),
	}

	// Consumer patterns
	bme.consumerPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)consumer`),
		regexp.MustCompile(`(?i)individual`),
		regexp.MustCompile(`(?i)personal`),
		regexp.MustCompile(`(?i)retail\s+customer`),
		regexp.MustCompile(`(?i)end\s+user`),
		regexp.MustCompile(`(?i)consumer\s+market`),
		regexp.MustCompile(`(?i)personal\s+use`),
		regexp.MustCompile(`(?i)individual\s+user`),
		regexp.MustCompile(`(?i)consumer\s+app`),
	}

	// Pricing patterns
	bme.pricingPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)tiered\s+pricing`),
		regexp.MustCompile(`(?i)usage-based\s+pricing`),
		regexp.MustCompile(`(?i)per-user\s+pricing`),
		regexp.MustCompile(`(?i)per-seat\s+pricing`),
		regexp.MustCompile(`(?i)flat-rate`),
		regexp.MustCompile(`(?i)pay-per-use`),
		regexp.MustCompile(`(?i)commission\s+model`),
		regexp.MustCompile(`(?i)transaction\s+fee`),
		regexp.MustCompile(`(?i)percentage\s+commission`),
		regexp.MustCompile(`(?i)advertising\s+revenue`),
	}
}

// ExtractBusinessModel extracts business model information from business data
func (bme *BusinessModelExtractor) ExtractBusinessModel(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
) (*BusinessModel, error) {
	ctx, span := bme.tracer.Start(ctx, "BusinessModelExtractor.ExtractBusinessModel")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessData.BusinessName),
		attribute.String("website", businessData.WebsiteURL),
	)

	// Create result structure
	result := &BusinessModel{
		ExtractedAt:  time.Now(),
		DataSources:  []string{"text_analysis", "pattern_matching"},
		ModelDetails: make(map[string]interface{}),
	}

	// Extract business model type
	if err := bme.extractBusinessModelType(ctx, businessData, result); err != nil {
		bme.logger.Warn("failed to extract business model type", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract revenue model
	if err := bme.extractRevenueModel(ctx, businessData, result); err != nil {
		bme.logger.Warn("failed to extract revenue model", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract target market
	if err := bme.extractTargetMarket(ctx, businessData, result); err != nil {
		bme.logger.Warn("failed to extract target market", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract pricing model
	if err := bme.extractPricingModel(ctx, businessData, result); err != nil {
		bme.logger.Warn("failed to extract pricing model", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Calculate overall confidence
	bme.calculateOverallConfidence(result)

	// Validate results
	if err := bme.validateResults(result); err != nil {
		bme.logger.Warn("business model validation failed", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	bme.logger.Info("business model extraction completed", map[string]interface{}{
		"business_name":       businessData.BusinessName,
		"business_model_type": result.BusinessModelType,
		"revenue_model":       result.RevenueModel,
		"target_market":       result.TargetMarket,
		"pricing_model":       result.PricingModel,
		"overall_confidence":  result.OverallConfidence,
	})

	return result, nil
}

// extractBusinessModelType extracts the business model type
func (bme *BusinessModelExtractor) extractBusinessModelType(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *BusinessModel,
) error {
	ctx, span := bme.tracer.Start(ctx, "BusinessModelExtractor.extractBusinessModelType")
	defer span.End()

	// Combine all text for analysis
	text := bme.combineText(businessData)

	// Check for different business model types
	modelScores := make(map[string]float64)

	// Check B2B patterns
	for _, pattern := range bme.b2bPatterns {
		if pattern.MatchString(text) {
			modelScores[BusinessModelB2B] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check B2C patterns
	for _, pattern := range bme.b2cPatterns {
		if pattern.MatchString(text) {
			modelScores[BusinessModelB2C] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check B2B2C patterns
	for _, pattern := range bme.b2b2cPatterns {
		if pattern.MatchString(text) {
			modelScores[BusinessModelB2B2C] += 0.9
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check Marketplace patterns
	for _, pattern := range bme.marketplacePatterns {
		if pattern.MatchString(text) {
			modelScores[BusinessModelMarketplace] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check SaaS patterns
	for _, pattern := range bme.saasPatterns {
		if pattern.MatchString(text) {
			modelScores[BusinessModelSaaS] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Determine the best model type
	var bestModel string
	var bestScore float64

	for model, score := range modelScores {
		if score > bestScore {
			bestModel = model
			bestScore = score
		}
	}

	// Set the business model type
	if bestScore > 0 {
		result.BusinessModelType = bestModel
		result.ModelConfidence = bme.normalizeConfidence(bestScore)
	} else {
		// Try to infer from context
		result.BusinessModelType, result.ModelConfidence = bme.inferBusinessModelType(text)
	}

	span.SetAttributes(
		attribute.String("business_model_type", result.BusinessModelType),
		attribute.Float64("confidence", result.ModelConfidence),
	)

	return nil
}

// extractRevenueModel extracts the revenue model
func (bme *BusinessModelExtractor) extractRevenueModel(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *BusinessModel,
) error {
	ctx, span := bme.tracer.Start(ctx, "BusinessModelExtractor.extractRevenueModel")
	defer span.End()

	// Combine all text for analysis
	text := bme.combineText(businessData)

	// Check for different revenue models
	revenueScores := make(map[string]float64)

	// Check subscription patterns
	for _, pattern := range bme.subscriptionPatterns {
		if pattern.MatchString(text) {
			revenueScores[RevenueModelSubscription] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check one-time patterns
	for _, pattern := range bme.oneTimePatterns {
		if pattern.MatchString(text) {
			revenueScores[RevenueModelOneTime] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check freemium patterns
	for _, pattern := range bme.freemiumPatterns {
		if pattern.MatchString(text) {
			revenueScores[RevenueModelFreemium] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check usage-based patterns
	usagePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)usage-based`),
		regexp.MustCompile(`(?i)pay-per-use`),
		regexp.MustCompile(`(?i)metered`),
		regexp.MustCompile(`(?i)consumption-based`),
	}

	for _, pattern := range usagePatterns {
		if pattern.MatchString(text) {
			revenueScores[RevenueModelUsageBased] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check commission patterns
	commissionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)commission`),
		regexp.MustCompile(`(?i)transaction\s+fee`),
		regexp.MustCompile(`(?i)percentage\s+cut`),
		regexp.MustCompile(`(?i)take\s+rate`),
	}

	for _, pattern := range commissionPatterns {
		if pattern.MatchString(text) {
			revenueScores[RevenueModelCommission] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Determine the best revenue model
	var bestModel string
	var bestScore float64

	for model, score := range revenueScores {
		if score > bestScore {
			bestModel = model
			bestScore = score
		}
	}

	// Set the revenue model
	if bestScore > 0 {
		result.RevenueModel = bestModel
		result.RevenueConfidence = bme.normalizeConfidence(bestScore)
	} else {
		// Try to infer from context
		result.RevenueModel, result.RevenueConfidence = bme.inferRevenueModel(text)
	}

	span.SetAttributes(
		attribute.String("revenue_model", result.RevenueModel),
		attribute.Float64("confidence", result.RevenueConfidence),
	)

	return nil
}

// extractTargetMarket extracts the target market
func (bme *BusinessModelExtractor) extractTargetMarket(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *BusinessModel,
) error {
	ctx, span := bme.tracer.Start(ctx, "BusinessModelExtractor.extractTargetMarket")
	defer span.End()

	// Combine all text for analysis
	text := bme.combineText(businessData)

	// Check for different target markets
	marketScores := make(map[string]float64)

	// Check enterprise patterns
	for _, pattern := range bme.enterprisePatterns {
		if pattern.MatchString(text) {
			marketScores[TargetMarketEnterprise] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check consumer patterns
	for _, pattern := range bme.consumerPatterns {
		if pattern.MatchString(text) {
			marketScores[TargetMarketConsumer] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check SMB patterns
	smbPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)small\s+business`),
		regexp.MustCompile(`(?i)medium\s+business`),
		regexp.MustCompile(`(?i)sme`),
		regexp.MustCompile(`(?i)small\s+and\s+medium`),
	}

	for _, pattern := range smbPatterns {
		if pattern.MatchString(text) {
			marketScores[TargetMarketSMB] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check startup patterns
	startupPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)startup`),
		regexp.MustCompile(`(?i)early-stage`),
		regexp.MustCompile(`(?i)seed-stage`),
	}

	for _, pattern := range startupPatterns {
		if pattern.MatchString(text) {
			marketScores[TargetMarketStartup] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Determine the best target market
	var bestMarket string
	var bestScore float64

	for market, score := range marketScores {
		if score > bestScore {
			bestMarket = market
			bestScore = score
		}
	}

	// Set the target market
	if bestScore > 0 {
		result.TargetMarket = bestMarket
		result.MarketConfidence = bme.normalizeConfidence(bestScore)
	} else {
		// Try to infer from context
		result.TargetMarket, result.MarketConfidence = bme.inferTargetMarket(text)
	}

	span.SetAttributes(
		attribute.String("target_market", result.TargetMarket),
		attribute.Float64("confidence", result.MarketConfidence),
	)

	return nil
}

// extractPricingModel extracts the pricing model
func (bme *BusinessModelExtractor) extractPricingModel(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *BusinessModel,
) error {
	ctx, span := bme.tracer.Start(ctx, "BusinessModelExtractor.extractPricingModel")
	defer span.End()

	// Combine all text for analysis
	text := bme.combineText(businessData)

	// Check for different pricing models
	pricingScores := make(map[string]float64)

	// Check pricing patterns
	for _, pattern := range bme.pricingPatterns {
		if pattern.MatchString(text) {
			// Determine which pricing model this pattern indicates
			patternStr := pattern.String()
			if strings.Contains(strings.ToLower(patternStr), "tiered") {
				pricingScores[PricingModelTiered] += 0.8
			} else if strings.Contains(strings.ToLower(patternStr), "usage") {
				pricingScores[PricingModelUsageBased] += 0.8
			} else if strings.Contains(strings.ToLower(patternStr), "flat") {
				pricingScores[PricingModelFlatRate] += 0.8
			} else if strings.Contains(strings.ToLower(patternStr), "per-user") {
				pricingScores[PricingModelPerUser] += 0.8
			} else if strings.Contains(strings.ToLower(patternStr), "commission") {
				pricingScores[PricingModelTiered] += 0.8
			} else if strings.Contains(strings.ToLower(patternStr), "advertising") {
				pricingScores[PricingModelTiered] += 0.8
			}
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Check freemium patterns
	for _, pattern := range bme.freemiumPatterns {
		if pattern.MatchString(text) {
			pricingScores[PricingModelFreemium] += 0.8
			result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
		}
	}

	// Determine the best pricing model
	var bestModel string
	var bestScore float64

	for model, score := range pricingScores {
		if score > bestScore {
			bestModel = model
			bestScore = score
		}
	}

	// Set the pricing model
	if bestScore > 0 {
		result.PricingModel = bestModel
		result.PricingConfidence = bme.normalizeConfidence(bestScore)
	} else {
		// Try to infer from context
		result.PricingModel, result.PricingConfidence = bme.inferPricingModel(text)
	}

	span.SetAttributes(
		attribute.String("pricing_model", result.PricingModel),
		attribute.Float64("confidence", result.PricingConfidence),
	)

	return nil
}

// combineText combines all available text for analysis
func (bme *BusinessModelExtractor) combineText(businessData *shared.BusinessClassificationRequest) string {
	var parts []string

	// Add business name
	if businessData.BusinessName != "" {
		parts = append(parts, businessData.BusinessName)
	}

	// Add description
	if businessData.Description != "" {
		parts = append(parts, businessData.Description)
	}

	// Add keywords
	if len(businessData.Keywords) > 0 {
		parts = append(parts, strings.Join(businessData.Keywords, " "))
	}

	// Add address
	if businessData.Address != "" {
		parts = append(parts, businessData.Address)
	}

	// Combine all parts
	text := strings.Join(parts, " ")

	// Normalize text
	if !bme.config.CaseSensitive {
		text = strings.ToLower(text)
	}

	return text
}

// Inference methods for when explicit patterns are not found
func (bme *BusinessModelExtractor) inferBusinessModelType(text string) (string, float64) {
	// Check for industry-based inference
	if strings.Contains(text, "software") || strings.Contains(text, "platform") {
		return BusinessModelSaaS, 0.6
	}
	if strings.Contains(text, "marketplace") || strings.Contains(text, "exchange") {
		return BusinessModelMarketplace, 0.6
	}
	if strings.Contains(text, "consulting") || strings.Contains(text, "agency") {
		return BusinessModelConsulting, 0.6
	}
	if strings.Contains(text, "retail") || strings.Contains(text, "store") {
		return BusinessModelRetail, 0.6
	}
	if strings.Contains(text, "manufacturing") || strings.Contains(text, "factory") {
		return BusinessModelManufacturing, 0.6
	}

	// Default to B2B for business-focused text
	if strings.Contains(text, "business") || strings.Contains(text, "enterprise") {
		return BusinessModelB2B, 0.5
	}

	// Default to B2C for consumer-focused text
	if strings.Contains(text, "consumer") || strings.Contains(text, "personal") {
		return BusinessModelB2C, 0.5
	}

	return "unknown", 0.0
}

func (bme *BusinessModelExtractor) inferRevenueModel(text string) (string, float64) {
	// Check for subscription indicators
	if strings.Contains(text, "monthly") || strings.Contains(text, "annual") {
		return RevenueModelSubscription, 0.6
	}

	// Check for one-time indicators
	if strings.Contains(text, "purchase") || strings.Contains(text, "buy") {
		return RevenueModelOneTime, 0.6
	}

	// Check for freemium indicators
	if strings.Contains(text, "free") && strings.Contains(text, "premium") {
		return RevenueModelFreemium, 0.6
	}

	// Check for service-based indicators
	if strings.Contains(text, "service") || strings.Contains(text, "consulting") {
		return RevenueModelServices, 0.6
	}

	// Default to subscription for SaaS-like businesses
	if strings.Contains(text, "software") || strings.Contains(text, "platform") {
		return RevenueModelSubscription, 0.5
	}

	return "unknown", 0.0
}

func (bme *BusinessModelExtractor) inferTargetMarket(text string) (string, float64) {
	// Check for enterprise indicators
	if strings.Contains(text, "enterprise") || strings.Contains(text, "corporate") {
		return TargetMarketEnterprise, 0.6
	}

	// Check for consumer indicators
	if strings.Contains(text, "consumer") || strings.Contains(text, "personal") {
		return TargetMarketConsumer, 0.6
	}

	// Check for SMB indicators
	if strings.Contains(text, "small") || strings.Contains(text, "medium") {
		return TargetMarketSMB, 0.6
	}

	// Check for startup indicators
	if strings.Contains(text, "startup") || strings.Contains(text, "early-stage") {
		return TargetMarketStartup, 0.6
	}

	// Default to both if unclear
	return TargetMarketBoth, 0.3
}

func (bme *BusinessModelExtractor) inferPricingModel(text string) (string, float64) {
	// Check for tiered indicators
	if strings.Contains(text, "tier") || strings.Contains(text, "plan") {
		return PricingModelTiered, 0.6
	}

	// Check for usage-based indicators
	if strings.Contains(text, "usage") || strings.Contains(text, "metered") {
		return PricingModelUsageBased, 0.6
	}

	// Check for per-user indicators
	if strings.Contains(text, "per user") || strings.Contains(text, "per seat") {
		return PricingModelPerUser, 0.6
	}

	// Default to tiered for most businesses
	return PricingModelTiered, 0.4
}

// normalizeConfidence normalizes confidence scores to 0-1 range
func (bme *BusinessModelExtractor) normalizeConfidence(score float64) float64 {
	if score > bme.config.MaxConfidenceThreshold {
		return bme.config.MaxConfidenceThreshold
	}
	if score < bme.config.MinConfidenceThreshold {
		return bme.config.MinConfidenceThreshold
	}
	return score
}

// calculateOverallConfidence calculates the overall confidence score
func (bme *BusinessModelExtractor) calculateOverallConfidence(result *BusinessModel) {
	var scores []float64
	var weights []float64

	// Business model type confidence
	if result.ModelConfidence > 0 {
		scores = append(scores, result.ModelConfidence)
		weights = append(weights, 0.3) // 30% weight
	}

	// Revenue model confidence
	if result.RevenueConfidence > 0 {
		scores = append(scores, result.RevenueConfidence)
		weights = append(weights, 0.3) // 30% weight
	}

	// Target market confidence
	if result.MarketConfidence > 0 {
		scores = append(scores, result.MarketConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	// Pricing model confidence
	if result.PricingConfidence > 0 {
		scores = append(scores, result.PricingConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	// Calculate weighted average
	if len(scores) > 0 {
		totalWeight := 0.0
		weightedSum := 0.0

		for i, score := range scores {
			weight := weights[i]
			weightedSum += score * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			result.OverallConfidence = weightedSum / totalWeight
		}
	} else {
		result.OverallConfidence = 0.0
	}
}

// validateResults validates the extracted results
func (bme *BusinessModelExtractor) validateResults(result *BusinessModel) error {
	// Validate confidence scores
	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		return fmt.Errorf("overall confidence score %f is out of range [0,1]", result.OverallConfidence)
	}

	// Validate business model type
	if result.BusinessModelType != "" {
		validModels := []string{
			BusinessModelB2B, BusinessModelB2C, BusinessModelB2B2C,
			BusinessModelMarketplace, BusinessModelSaaS, BusinessModelEcommerce,
			BusinessModelConsulting, BusinessModelAgency, BusinessModelManufacturing,
			BusinessModelRetail,
		}
		valid := false
		for _, model := range validModels {
			if result.BusinessModelType == model {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid business model type: %s", result.BusinessModelType)
		}
	}

	// Validate revenue model
	if result.RevenueModel != "" {
		validModels := []string{
			RevenueModelSubscription, RevenueModelOneTime, RevenueModelFreemium,
			RevenueModelUsageBased, RevenueModelTiered, RevenueModelCommission,
			RevenueModelAdvertising, RevenueModelLicensing, RevenueModelServices,
		}
		valid := false
		for _, model := range validModels {
			if result.RevenueModel == model {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid revenue model: %s", result.RevenueModel)
		}
	}

	return nil
}
