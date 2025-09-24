package enrichment

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// BusinessModelAnalyzer analyzes website content to determine business model type
type BusinessModelAnalyzer struct {
	config *BusinessModelConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// BusinessModelConfig contains configuration for business model analysis
type BusinessModelConfig struct {
	// Analysis settings
	EnableB2BAnalysis         bool `json:"enable_b2b_analysis"`
	EnableB2CAnalysis         bool `json:"enable_b2c_analysis"`
	EnableMarketplaceAnalysis bool `json:"enable_marketplace_analysis"`

	// Thresholds
	ConfidenceThreshold  float64 `json:"confidence_threshold"`
	MinimumEvidenceCount int     `json:"minimum_evidence_count"`

	// Weights for different indicators
	KeywordWeight          float64 `json:"keyword_weight"`
	ContentStructureWeight float64 `json:"content_structure_weight"`
	PricingModelWeight     float64 `json:"pricing_model_weight"`
	AudienceWeight         float64 `json:"audience_weight"`

	// Analysis depth
	MaxAnalysisLength  int  `json:"max_analysis_length"`
	EnableDeepAnalysis bool `json:"enable_deep_analysis"`
}

// BusinessModelResult contains the analysis results
type BusinessModelResult struct {
	// Primary classification
	BusinessModel   string  `json:"business_model"`   // "b2b", "b2c", "marketplace", "hybrid", "unknown"
	ConfidenceScore float64 `json:"confidence_score"` // 0.0-1.0
	PrimaryModel    string  `json:"primary_model"`

	// Detailed scores
	B2BScore         float64 `json:"b2b_score"`
	B2CScore         float64 `json:"b2c_score"`
	MarketplaceScore float64 `json:"marketplace_score"`

	// Evidence and indicators
	Evidence              []string `json:"evidence"`
	B2BIndicators         []string `json:"b2b_indicators"`
	B2CIndicators         []string `json:"b2c_indicators"`
	MarketplaceIndicators []string `json:"marketplace_indicators"`

	// Target audience analysis
	TargetAudience     string  `json:"target_audience"`
	AudienceConfidence float64 `json:"audience_confidence"`

	// Revenue model analysis
	RevenueModel    string `json:"revenue_model"`
	PricingStrategy string `json:"pricing_strategy"`

	// Analysis metadata
	ExtractionMethod string           `json:"extraction_method"`
	IsValidated      bool             `json:"is_validated"`
	ValidationStatus ValidationStatus `json:"validation_status"`
	ExtractedAt      time.Time        `json:"extracted_at"`
	SourceURL        string           `json:"source_url"`

	// Quality metrics
	DataQualityScore float64 `json:"data_quality_score"`
	Reasoning        string  `json:"reasoning"`
}

// NewBusinessModelAnalyzer creates a new business model analyzer
func NewBusinessModelAnalyzer(config *BusinessModelConfig, logger *zap.Logger) *BusinessModelAnalyzer {
	if config == nil {
		config = &BusinessModelConfig{
			EnableB2BAnalysis:         true,
			EnableB2CAnalysis:         true,
			EnableMarketplaceAnalysis: true,
			ConfidenceThreshold:       0.6,
			MinimumEvidenceCount:      2,
			KeywordWeight:             0.4,
			ContentStructureWeight:    0.3,
			PricingModelWeight:        0.2,
			AudienceWeight:            0.1,
			MaxAnalysisLength:         10000,
			EnableDeepAnalysis:        true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &BusinessModelAnalyzer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("business-model-analyzer"),
	}
}

// AnalyzeBusinessModel analyzes website content to determine business model
func (bma *BusinessModelAnalyzer) AnalyzeBusinessModel(ctx context.Context, content, sourceURL string) (*BusinessModelResult, error) {
	ctx, span := bma.tracer.Start(ctx, "AnalyzeBusinessModel",
		trace.WithAttributes(
			attribute.String("source_url", sourceURL),
			attribute.Int("content_length", len(content)),
			attribute.Bool("enable_b2b", bma.config.EnableB2BAnalysis),
			attribute.Bool("enable_b2c", bma.config.EnableB2CAnalysis),
			attribute.Bool("enable_marketplace", bma.config.EnableMarketplaceAnalysis),
		))
	defer span.End()

	bma.logger.Info("Starting business model analysis",
		zap.String("source_url", sourceURL),
		zap.Int("content_length", len(content)))

	// Truncate content if needed
	if len(content) > bma.config.MaxAnalysisLength {
		content = content[:bma.config.MaxAnalysisLength]
	}

	result := &BusinessModelResult{
		ExtractedAt:           time.Now(),
		SourceURL:             sourceURL,
		ExtractionMethod:      "content_analysis",
		Evidence:              []string{},
		B2BIndicators:         []string{},
		B2CIndicators:         []string{},
		MarketplaceIndicators: []string{},
	}

	// Analyze B2B indicators
	if bma.config.EnableB2BAnalysis {
		b2bScore, b2bIndicators := bma.analyzeB2BIndicators(content)
		result.B2BScore = b2bScore
		result.B2BIndicators = b2bIndicators
		result.Evidence = append(result.Evidence, b2bIndicators...)
	}

	// Analyze B2C indicators
	if bma.config.EnableB2CAnalysis {
		b2cScore, b2cIndicators := bma.analyzeB2CIndicators(content)
		result.B2CScore = b2cScore
		result.B2CIndicators = b2cIndicators
		result.Evidence = append(result.Evidence, b2cIndicators...)
	}

	// Analyze marketplace indicators
	if bma.config.EnableMarketplaceAnalysis {
		marketplaceScore, marketplaceIndicators := bma.analyzeMarketplaceIndicators(content)
		result.MarketplaceScore = marketplaceScore
		result.MarketplaceIndicators = marketplaceIndicators
		result.Evidence = append(result.Evidence, marketplaceIndicators...)
	}

	// Determine primary business model
	result.PrimaryModel = bma.determinePrimaryModel(result)
	result.BusinessModel = result.PrimaryModel

	// Calculate overall confidence
	result.ConfidenceScore = bma.calculateConfidence(result)

	// Analyze target audience
	result.TargetAudience, result.AudienceConfidence = bma.analyzeTargetAudience(content)

	// Analyze revenue model
	result.RevenueModel, result.PricingStrategy = bma.analyzeRevenueModel(content)

	// Generate reasoning
	result.Reasoning = bma.generateReasoning(result)

	// Validate result
	result.IsValidated = bma.validateResult(result)
	result.ValidationStatus = bma.createValidationStatus(result)

	// Calculate data quality
	result.DataQualityScore = bma.calculateDataQuality(result)

	bma.logger.Info("Business model analysis completed",
		zap.String("business_model", result.BusinessModel),
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.String("target_audience", result.TargetAudience),
		zap.String("revenue_model", result.RevenueModel))

	span.SetAttributes(
		attribute.String("business_model", result.BusinessModel),
		attribute.Float64("confidence_score", result.ConfidenceScore),
		attribute.String("target_audience", result.TargetAudience),
		attribute.String("revenue_model", result.RevenueModel))

	return result, nil
}

// analyzeB2BIndicators analyzes content for B2B business model indicators
func (bma *BusinessModelAnalyzer) analyzeB2BIndicators(content string) (float64, []string) {
	indicators := []string{}
	score := 0.0
	factorCount := 0

	// B2B keywords and phrases
	b2bKeywords := []string{
		"enterprise", "business", "corporate", "professional", "industry",
		"solution", "platform", "software", "service", "consulting",
		"partnership", "integration", "api", "sdk", "enterprise-grade",
		"scalable", "enterprise solution", "business solution", "corporate solution",
		"b2b", "business to business", "enterprise software", "saas", "cloud platform",
		"workflow", "automation", "enterprise tools", "business tools", "professional services",
		"consulting services", "enterprise support", "business support", "enterprise features",
		"enterprise pricing", "enterprise plan", "business plan", "corporate plan",
	}

	contentLower := strings.ToLower(content)

	// Check for B2B keywords
	keywordMatches := 0
	for _, keyword := range b2bKeywords {
		if strings.Contains(contentLower, keyword) {
			keywordMatches++
			indicators = append(indicators, fmt.Sprintf("B2B keyword: %s", keyword))
		}
	}

	if keywordMatches > 0 {
		keywordScore := math.Min(1.0, float64(keywordMatches)/float64(len(b2bKeywords)/2))
		score += keywordScore * bma.config.KeywordWeight
		factorCount++
	}

	// Check for enterprise-focused content structure
	if bma.analyzeEnterpriseContentStructure(content) {
		score += 0.8 * bma.config.ContentStructureWeight
		factorCount++
		indicators = append(indicators, "Enterprise-focused content structure")
	}

	// Check for B2B pricing models
	if bma.analyzeB2BPricingModels(content) {
		score += 0.9 * bma.config.PricingModelWeight
		factorCount++
		indicators = append(indicators, "B2B pricing model indicators")
	}

	// Check for enterprise audience targeting
	if bma.analyzeEnterpriseAudience(content) {
		score += 0.7 * bma.config.AudienceWeight
		factorCount++
		indicators = append(indicators, "Enterprise audience targeting")
	}

	if factorCount == 0 {
		return 0.0, indicators
	}

	return score / float64(factorCount), indicators
}

// analyzeB2CIndicators analyzes content for B2C business model indicators
func (bma *BusinessModelAnalyzer) analyzeB2CIndicators(content string) (float64, []string) {
	indicators := []string{}
	score := 0.0
	factorCount := 0

	// B2C keywords and phrases
	b2cKeywords := []string{
		"consumer", "personal", "individual", "customer", "user",
		"shopping", "retail", "e-commerce", "online store", "buy now",
		"personal use", "individual use", "consumer product", "personal product",
		"shopping cart", "checkout", "payment", "credit card", "personal account",
		"user account", "personal plan", "individual plan", "consumer plan",
		"personal pricing", "individual pricing", "consumer pricing", "retail price",
		"personal service", "individual service", "consumer service", "customer service",
		"personal support", "individual support", "consumer support", "customer support",
		"personal features", "individual features", "consumer features", "user features",
	}

	contentLower := strings.ToLower(content)

	// Check for B2C keywords
	keywordMatches := 0
	for _, keyword := range b2cKeywords {
		if strings.Contains(contentLower, keyword) {
			keywordMatches++
			indicators = append(indicators, fmt.Sprintf("B2C keyword: %s", keyword))
		}
	}

	if keywordMatches > 0 {
		keywordScore := math.Min(1.0, float64(keywordMatches)/float64(len(b2cKeywords)/2))
		score += keywordScore * bma.config.KeywordWeight
		factorCount++
	}

	// Check for consumer-focused content structure
	if bma.analyzeConsumerContentStructure(content) {
		score += 0.8 * bma.config.ContentStructureWeight
		factorCount++
		indicators = append(indicators, "Consumer-focused content structure")
	}

	// Check for B2C pricing models
	if bma.analyzeB2CPricingModels(content) {
		score += 0.9 * bma.config.PricingModelWeight
		factorCount++
		indicators = append(indicators, "B2C pricing model indicators")
	}

	// Check for consumer audience targeting
	if bma.analyzeConsumerAudience(content) {
		score += 0.7 * bma.config.AudienceWeight
		factorCount++
		indicators = append(indicators, "Consumer audience targeting")
	}

	if factorCount == 0 {
		return 0.0, indicators
	}

	return score / float64(factorCount), indicators
}

// analyzeMarketplaceIndicators analyzes content for marketplace business model indicators
func (bma *BusinessModelAnalyzer) analyzeMarketplaceIndicators(content string) (float64, []string) {
	indicators := []string{}
	score := 0.0
	factorCount := 0

	// Marketplace keywords and phrases
	marketplaceKeywords := []string{
		"marketplace", "platform", "connect", "buyer", "seller", "vendor",
		"listing", "product", "service", "market", "exchange", "trading",
		"commission", "fee", "transaction", "payment", "escrow", "verification",
		"review", "rating", "feedback", "trust", "safety", "dispute",
		"buyer protection", "seller protection", "marketplace fee", "commission fee",
		"transaction fee", "listing fee", "vendor fee", "seller fee", "buyer fee",
		"marketplace platform", "trading platform", "exchange platform", "connect buyers",
		"connect sellers", "buyer seller", "vendor marketplace", "product marketplace",
		"service marketplace", "online marketplace", "digital marketplace", "e-commerce marketplace",
	}

	contentLower := strings.ToLower(content)

	// Check for marketplace keywords
	keywordMatches := 0
	for _, keyword := range marketplaceKeywords {
		if strings.Contains(contentLower, keyword) {
			keywordMatches++
			indicators = append(indicators, fmt.Sprintf("Marketplace keyword: %s", keyword))
		}
	}

	if keywordMatches > 0 {
		keywordScore := math.Min(1.0, float64(keywordMatches)/float64(len(marketplaceKeywords)/2))
		score += keywordScore * bma.config.KeywordWeight
		factorCount++
	}

	// Check for marketplace content structure
	if bma.analyzeMarketplaceContentStructure(content) {
		score += 0.8 * bma.config.ContentStructureWeight
		factorCount++
		indicators = append(indicators, "Marketplace content structure")
	}

	// Check for marketplace pricing models
	if bma.analyzeMarketplacePricingModels(content) {
		score += 0.9 * bma.config.PricingModelWeight
		factorCount++
		indicators = append(indicators, "Marketplace pricing model indicators")
	}

	// Check for marketplace audience targeting
	if bma.analyzeMarketplaceAudience(content) {
		score += 0.7 * bma.config.AudienceWeight
		factorCount++
		indicators = append(indicators, "Marketplace audience targeting")
	}

	if factorCount == 0 {
		return 0.0, indicators
	}

	return score / float64(factorCount), indicators
}

// Helper methods for content structure analysis
func (bma *BusinessModelAnalyzer) analyzeEnterpriseContentStructure(content string) bool {
	enterprisePatterns := []string{
		"enterprise solution", "business solution", "corporate solution",
		"enterprise features", "business features", "corporate features",
		"enterprise pricing", "business pricing", "corporate pricing",
		"enterprise plan", "business plan", "corporate plan",
		"enterprise support", "business support", "corporate support",
		"enterprise integration", "business integration", "corporate integration",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range enterprisePatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

func (bma *BusinessModelAnalyzer) analyzeConsumerContentStructure(content string) bool {
	consumerPatterns := []string{
		"personal use", "individual use", "consumer use",
		"personal plan", "individual plan", "consumer plan",
		"personal pricing", "individual pricing", "consumer pricing",
		"personal features", "individual features", "consumer features",
		"personal support", "individual support", "consumer support",
		"shopping cart", "checkout", "buy now", "add to cart",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range consumerPatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

func (bma *BusinessModelAnalyzer) analyzeMarketplaceContentStructure(content string) bool {
	marketplacePatterns := []string{
		"buyer seller", "vendor marketplace", "product marketplace",
		"service marketplace", "connect buyers", "connect sellers",
		"listing fee", "commission fee", "transaction fee",
		"buyer protection", "seller protection", "dispute resolution",
		"review rating", "feedback system", "trust safety",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range marketplacePatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

// Helper methods for pricing model analysis
func (bma *BusinessModelAnalyzer) analyzeB2BPricingModels(content string) bool {
	b2bPricingPatterns := []string{
		"enterprise pricing", "business pricing", "corporate pricing",
		"enterprise plan", "business plan", "corporate plan",
		"annual contract", "enterprise contract", "business contract",
		"volume pricing", "enterprise discount", "business discount",
		"custom pricing", "enterprise quote", "business quote",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range b2bPricingPatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

func (bma *BusinessModelAnalyzer) analyzeB2CPricingModels(content string) bool {
	b2cPricingPatterns := []string{
		"personal pricing", "individual pricing", "consumer pricing",
		"personal plan", "individual plan", "consumer plan",
		"monthly subscription", "personal subscription", "individual subscription",
		"one-time purchase", "personal purchase", "individual purchase",
		"retail price", "consumer price", "personal price",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range b2cPricingPatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

func (bma *BusinessModelAnalyzer) analyzeMarketplacePricingModels(content string) bool {
	marketplacePricingPatterns := []string{
		"commission fee", "transaction fee", "listing fee",
		"marketplace fee", "vendor fee", "seller fee",
		"buyer fee", "platform fee", "service fee",
		"percentage fee", "flat fee", "processing fee",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range marketplacePricingPatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 2
}

// Helper methods for audience analysis
func (bma *BusinessModelAnalyzer) analyzeEnterpriseAudience(content string) bool {
	enterpriseAudiencePatterns := []string{
		"enterprise customers", "business customers", "corporate customers",
		"enterprise clients", "business clients", "corporate clients",
		"enterprise users", "business users", "corporate users",
		"enterprise companies", "business companies", "corporate companies",
		"enterprise organizations", "business organizations", "corporate organizations",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range enterpriseAudiencePatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 1
}

func (bma *BusinessModelAnalyzer) analyzeConsumerAudience(content string) bool {
	consumerAudiencePatterns := []string{
		"personal users", "individual users", "consumer users",
		"personal customers", "individual customers", "consumer customers",
		"personal clients", "individual clients", "consumer clients",
		"end users", "home users", "personal use",
		"individual use", "consumer use", "personal customers",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range consumerAudiencePatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 1
}

func (bma *BusinessModelAnalyzer) analyzeMarketplaceAudience(content string) bool {
	marketplaceAudiencePatterns := []string{
		"buyers and sellers", "vendors and customers", "buyers sellers",
		"marketplace users", "platform users", "trading users",
		"buyer community", "seller community", "vendor community",
		"marketplace participants", "platform participants", "trading participants",
	}

	contentLower := strings.ToLower(content)
	matches := 0

	for _, pattern := range marketplaceAudiencePatterns {
		if strings.Contains(contentLower, pattern) {
			matches++
		}
	}

	return matches >= 1
}

// determinePrimaryModel determines the primary business model based on scores
func (bma *BusinessModelAnalyzer) determinePrimaryModel(result *BusinessModelResult) string {
	scores := map[string]float64{
		"b2b":         result.B2BScore,
		"b2c":         result.B2CScore,
		"marketplace": result.MarketplaceScore,
	}

	// Find the highest score
	maxScore := 0.0
	primaryModel := "unknown"

	for model, score := range scores {
		if score > maxScore {
			maxScore = score
			primaryModel = model
		}
	}

	// Check if it's a hybrid model
	if maxScore > 0.6 {
		// Count models with significant scores
		significantModels := 0
		for _, score := range scores {
			if score > 0.4 {
				significantModels++
			}
		}

		if significantModels > 1 {
			return "hybrid"
		}
	}

	if maxScore < bma.config.ConfidenceThreshold {
		return "unknown"
	}

	return primaryModel
}

// calculateConfidence calculates the overall confidence score
func (bma *BusinessModelAnalyzer) calculateConfidence(result *BusinessModelResult) float64 {
	// Base confidence on the primary model score
	var baseScore float64
	switch result.PrimaryModel {
	case "b2b":
		baseScore = result.B2BScore
	case "b2c":
		baseScore = result.B2CScore
	case "marketplace":
		baseScore = result.MarketplaceScore
	case "hybrid":
		// For hybrid, take the average of the two highest scores
		scores := []float64{result.B2BScore, result.B2CScore, result.MarketplaceScore}
		// Sort scores in descending order
		for i := 0; i < len(scores)-1; i++ {
			for j := i + 1; j < len(scores); j++ {
				if scores[i] < scores[j] {
					scores[i], scores[j] = scores[j], scores[i]
				}
			}
		}
		baseScore = (scores[0] + scores[1]) / 2
	default:
		baseScore = 0.0
	}

	// Adjust confidence based on evidence count
	evidenceFactor := math.Min(1.0, float64(len(result.Evidence))/float64(bma.config.MinimumEvidenceCount))

	// Adjust confidence based on data quality
	qualityFactor := result.DataQualityScore

	return (baseScore * 0.6) + (evidenceFactor * 0.3) + (qualityFactor * 0.1)
}

// analyzeTargetAudience analyzes the target audience from content
func (bma *BusinessModelAnalyzer) analyzeTargetAudience(content string) (string, float64) {
	contentLower := strings.ToLower(content)

	// Check for enterprise audience
	enterprisePatterns := []string{"enterprise", "business", "corporate", "professional", "industry"}
	enterpriseMatches := 0
	for _, pattern := range enterprisePatterns {
		if strings.Contains(contentLower, pattern) {
			enterpriseMatches++
		}
	}

	// Check for consumer audience
	consumerPatterns := []string{"consumer", "personal", "individual", "customer", "user", "shopping"}
	consumerMatches := 0
	for _, pattern := range consumerPatterns {
		if strings.Contains(contentLower, pattern) {
			consumerMatches++
		}
	}

	// Check for marketplace audience
	marketplacePatterns := []string{"buyer", "seller", "vendor", "marketplace", "platform", "connect"}
	marketplaceMatches := 0
	for _, pattern := range marketplacePatterns {
		if strings.Contains(contentLower, pattern) {
			marketplaceMatches++
		}
	}

	// Determine primary audience
	if enterpriseMatches > consumerMatches && enterpriseMatches > marketplaceMatches {
		return "enterprise", math.Min(1.0, float64(enterpriseMatches)/5.0)
	} else if consumerMatches > enterpriseMatches && consumerMatches > marketplaceMatches {
		return "consumer", math.Min(1.0, float64(consumerMatches)/6.0)
	} else if marketplaceMatches > enterpriseMatches && marketplaceMatches > consumerMatches {
		return "marketplace", math.Min(1.0, float64(marketplaceMatches)/6.0)
	} else {
		return "mixed", 0.5
	}
}

// analyzeRevenueModel analyzes the revenue model from content
func (bma *BusinessModelAnalyzer) analyzeRevenueModel(content string) (string, string) {
	contentLower := strings.ToLower(content)

	// Check for subscription model
	if strings.Contains(contentLower, "subscription") || strings.Contains(contentLower, "monthly") || strings.Contains(contentLower, "annual") {
		return "subscription", "recurring"
	}

	// Check for marketplace model
	if strings.Contains(contentLower, "commission") || strings.Contains(contentLower, "transaction fee") || strings.Contains(contentLower, "marketplace fee") {
		return "marketplace", "commission-based"
	}

	// Check for one-time purchase
	if strings.Contains(contentLower, "one-time") || strings.Contains(contentLower, "single purchase") || strings.Contains(contentLower, "buy once") {
		return "one-time", "single purchase"
	}

	// Check for freemium
	if strings.Contains(contentLower, "free") && (strings.Contains(contentLower, "premium") || strings.Contains(contentLower, "upgrade")) {
		return "freemium", "free with premium upgrade"
	}

	// Check for enterprise pricing
	if strings.Contains(contentLower, "enterprise pricing") || strings.Contains(contentLower, "custom pricing") || strings.Contains(contentLower, "contact sales") {
		return "enterprise", "custom pricing"
	}

	return "unknown", "unknown"
}

// generateReasoning generates reasoning for the business model classification
func (bma *BusinessModelAnalyzer) generateReasoning(result *BusinessModelResult) string {
	reasoning := fmt.Sprintf("Business model classified as '%s' with %.1f%% confidence. ",
		result.BusinessModel, result.ConfidenceScore*100)

	if len(result.Evidence) > 0 {
		reasoning += fmt.Sprintf("Key indicators: %s. ", strings.Join(result.Evidence[:min(3, len(result.Evidence))], ", "))
	}

	if result.TargetAudience != "" {
		reasoning += fmt.Sprintf("Target audience: %s (%.1f%% confidence). ",
			result.TargetAudience, result.AudienceConfidence*100)
	}

	if result.RevenueModel != "unknown" {
		reasoning += fmt.Sprintf("Revenue model: %s (%s). ", result.RevenueModel, result.PricingStrategy)
	}

	reasoning += fmt.Sprintf("B2B score: %.2f, B2C score: %.2f, Marketplace score: %.2f.",
		result.B2BScore, result.B2CScore, result.MarketplaceScore)

	return reasoning
}

// validateResult validates the business model analysis result
func (bma *BusinessModelAnalyzer) validateResult(result *BusinessModelResult) bool {
	// Check if we have sufficient evidence
	if len(result.Evidence) < bma.config.MinimumEvidenceCount {
		return false
	}

	// Check if confidence is above threshold
	if result.ConfidenceScore < bma.config.ConfidenceThreshold {
		return false
	}

	// Check if business model is valid
	validModels := []string{"b2b", "b2c", "marketplace", "hybrid", "unknown"}
	isValid := false
	for _, model := range validModels {
		if result.BusinessModel == model {
			isValid = true
			break
		}
	}

	return isValid
}

// createValidationStatus creates validation status for the result
func (bma *BusinessModelAnalyzer) createValidationStatus(result *BusinessModelResult) ValidationStatus {
	errors := []string{}

	if len(result.Evidence) < bma.config.MinimumEvidenceCount {
		errors = append(errors, fmt.Sprintf("Insufficient evidence: %d < %d",
			len(result.Evidence), bma.config.MinimumEvidenceCount))
	}

	if result.ConfidenceScore < bma.config.ConfidenceThreshold {
		errors = append(errors, fmt.Sprintf("Low confidence: %.2f < %.2f",
			result.ConfidenceScore, bma.config.ConfidenceThreshold))
	}

	return ValidationStatus{
		IsValid:          len(errors) == 0,
		ValidationErrors: errors,
	}
}

// calculateDataQuality calculates the data quality score
func (bma *BusinessModelAnalyzer) calculateDataQuality(result *BusinessModelResult) float64 {
	score := 0.0
	factorCount := 0

	// Evidence quality
	if len(result.Evidence) > 0 {
		evidenceScore := math.Min(1.0, float64(len(result.Evidence))/10.0)
		score += evidenceScore
		factorCount++
	}

	// Confidence quality
	if result.ConfidenceScore > 0.0 {
		score += result.ConfidenceScore
		factorCount++
	}

	// Target audience quality
	if result.AudienceConfidence > 0.0 {
		score += result.AudienceConfidence
		factorCount++
	}

	// Revenue model quality
	if result.RevenueModel != "unknown" {
		score += 0.8
		factorCount++
	}

	if factorCount == 0 {
		return 0.5
	}

	return score / float64(factorCount)
}
