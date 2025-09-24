package enrichment

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RevenueAnalyzer analyzes website content to identify revenue indicators and financial health signals
type RevenueAnalyzer struct {
	config *RevenueConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// RevenueConfig contains configuration for revenue analysis
type RevenueConfig struct {
	// Analysis settings
	EnableRevenueExtraction   bool `json:"enable_revenue_extraction"`
	EnableFinancialIndicators bool `json:"enable_financial_indicators"`
	EnableConfidenceScoring   bool `json:"enable_confidence_scoring"`
	EnableValidation          bool `json:"enable_validation"`

	// Revenue thresholds for classification
	StartupRevenueThreshold    int64 `json:"startup_revenue_threshold"`    // $0-1M
	SMEMinRevenueThreshold     int64 `json:"sme_min_revenue_threshold"`    // $1M
	SMEMaxRevenueThreshold     int64 `json:"sme_max_revenue_threshold"`    // $10M
	EnterpriseRevenueThreshold int64 `json:"enterprise_revenue_threshold"` // $10M+

	// Confidence settings
	MinConfidenceThreshold float64 `json:"min_confidence_threshold"`
	MaxExtractionAttempts  int     `json:"max_extraction_attempts"`

	// Quality settings
	EnableDuplicateDetection bool `json:"enable_duplicate_detection"`
	EnableContextValidation  bool `json:"enable_context_validation"`
}

// RevenueResult contains the results of revenue analysis
type RevenueResult struct {
	// Extracted revenue information
	RevenueAmount    int64   `json:"revenue_amount"`
	RevenueRange     string  `json:"revenue_range"`
	ConfidenceScore  float64 `json:"confidence_score"`
	ExtractionMethod string  `json:"extraction_method"`
	IsValidated      bool    `json:"is_validated"`

	// Financial health indicators
	FinancialHealth  string  `json:"financial_health"`
	HealthConfidence float64 `json:"health_confidence"`
	HealthCategory   string  `json:"health_category"`

	// Evidence and reasoning
	Evidence         []string `json:"evidence"`
	Reasoning        string   `json:"reasoning"`
	ExtractedPhrases []string `json:"extracted_phrases"`

	// Quality metrics
	DataQuality      DataQualityMetrics `json:"data_quality"`
	ValidationStatus ValidationStatus   `json:"validation_status"`

	// Metadata
	ExtractedAt time.Time              `json:"extracted_at"`
	SourceURL   string                 `json:"source_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// FinancialHealthCategory represents financial health categories
type FinancialHealthCategory struct {
	Category        string   `json:"category"`
	MinRevenue      int64    `json:"min_revenue"`
	MaxRevenue      int64    `json:"max_revenue"`
	ConfidenceScore float64  `json:"confidence_score"`
	Evidence        []string `json:"evidence"`
}

// NewRevenueAnalyzer creates a new revenue analyzer with default configuration
func NewRevenueAnalyzer(config *RevenueConfig, logger *zap.Logger) *RevenueAnalyzer {
	if config == nil {
		config = &RevenueConfig{
			EnableRevenueExtraction:   true,
			EnableFinancialIndicators: true,
			EnableConfidenceScoring:   true,
			EnableValidation:          true,

			StartupRevenueThreshold:    1000000,  // $1M
			SMEMinRevenueThreshold:     1000000,  // $1M
			SMEMaxRevenueThreshold:     10000000, // $10M
			EnterpriseRevenueThreshold: 10000000, // $10M+

			MinConfidenceThreshold: 0.3,
			MaxExtractionAttempts:  5,

			EnableDuplicateDetection: true,
			EnableContextValidation:  true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &RevenueAnalyzer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("revenue-analyzer"),
	}
}

// AnalyzeContent analyzes website content to extract revenue indicators and financial health signals
func (ra *RevenueAnalyzer) AnalyzeContent(ctx context.Context, content string) (*RevenueResult, error) {
	ctx, span := ra.tracer.Start(ctx, "revenue_analyzer.analyze_content",
		trace.WithAttributes(
			attribute.String("content_length", fmt.Sprintf("%d", len(content))),
			attribute.Bool("enable_revenue_extraction", ra.config.EnableRevenueExtraction),
			attribute.Bool("enable_financial_indicators", ra.config.EnableFinancialIndicators),
		))
	defer span.End()

	ra.logger.Info("Starting revenue analysis",
		zap.String("content_length", fmt.Sprintf("%d", len(content))),
		zap.Bool("enable_revenue_extraction", ra.config.EnableRevenueExtraction),
		zap.Bool("enable_financial_indicators", ra.config.EnableFinancialIndicators))

	result := &RevenueResult{
		ExtractedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Extract revenue information
	if ra.config.EnableRevenueExtraction {
		if err := ra.extractRevenue(ctx, content, result); err != nil {
			ra.logger.Error("Failed to extract revenue", zap.Error(err))
			span.RecordError(err)
		}
	}

	// Analyze financial health indicators
	if ra.config.EnableFinancialIndicators {
		if err := ra.analyzeFinancialHealth(ctx, content, result); err != nil {
			ra.logger.Error("Failed to analyze financial health", zap.Error(err))
			span.RecordError(err)
		}
	}

	// Calculate confidence score
	if ra.config.EnableConfidenceScoring {
		result.ConfidenceScore = ra.calculateConfidence(result)
	}

	// Validate result
	if ra.config.EnableValidation {
		if err := ra.validateResult(result); err != nil {
			ra.logger.Error("Failed to validate result", zap.Error(err))
			span.RecordError(err)
		}
	}

	// Generate reasoning
	result.Reasoning = ra.generateReasoning(result)

	ra.logger.Info("Revenue analysis completed",
		zap.Int64("revenue_amount", result.RevenueAmount),
		zap.String("financial_health", result.FinancialHealth),
		zap.Float64("confidence_score", result.ConfidenceScore))

	return result, nil
}

// extractRevenue extracts revenue information from content
func (ra *RevenueAnalyzer) extractRevenue(ctx context.Context, content string, result *RevenueResult) error {
	ctx, span := ra.tracer.Start(ctx, "revenue_analyzer.extract_revenue")
	defer span.End()

	// Direct revenue mentions
	revenuePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\$?(\d{1,3}(?:,\d{3})*(?:\.\d{2})?))\s*(?:million|mil|m)\s*(?:in\s+)?revenue`),
		regexp.MustCompile(`(?i)revenue\s+of\s+\$?(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)`),
		regexp.MustCompile(`(?i)annual\s+revenue\s+\$?(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)`),
		regexp.MustCompile(`(?i)generated\s+\$?(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)\s*(?:million|mil|m)\s*in\s+revenue`),
		regexp.MustCompile(`(?i)revenue\s+reached\s+\$?(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)`),
	}

	for _, pattern := range revenuePatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 2 {
				// Use the second capture group (index 2) for the numeric value
				// Check if the full match contains "million" to determine the multiplier
				fullMatch := match[0]
				numericValue := match[2]

				var amount int64
				var err error

				// Check if the full match contains million indicators
				if strings.Contains(strings.ToLower(fullMatch), "million") ||
					strings.Contains(strings.ToLower(fullMatch), "mil") ||
					strings.Contains(strings.ToLower(fullMatch), "m") {
					// Parse as millions
					amount, err = ra.parseRevenueAmount(numericValue + " million")
				} else {
					// Parse as direct amount
					amount, err = ra.parseRevenueAmount(numericValue)
				}

				if err == nil && amount > 0 {
					result.RevenueAmount = amount
					result.ExtractionMethod = "direct_mention"
					result.Evidence = append(result.Evidence, fmt.Sprintf("Direct revenue mention: %s", match[0]))
					result.ExtractedPhrases = append(result.ExtractedPhrases, match[0])
					break
				}
			} else if len(match) > 1 {
				// Fallback to first capture group for patterns without the outer group
				amount, err := ra.parseRevenueAmount(match[1])
				if err == nil && amount > 0 {
					result.RevenueAmount = amount
					result.ExtractionMethod = "direct_mention"
					result.Evidence = append(result.Evidence, fmt.Sprintf("Direct revenue mention: %s", match[0]))
					result.ExtractedPhrases = append(result.ExtractedPhrases, match[0])
					break
				}
			}
		}
		if result.RevenueAmount > 0 {
			break
		}
	}

	// Revenue range indicators
	if result.RevenueAmount == 0 {
		ra.extractRevenueRange(content, result)
	}

	// Financial indicators
	if result.RevenueAmount == 0 {
		ra.extractFinancialIndicators(content, result)
	}

	return nil
}

// parseRevenueAmount parses revenue amount from string
func (ra *RevenueAnalyzer) parseRevenueAmount(amountStr string) (int64, error) {
	// Remove currency symbols and commas
	cleanStr := strings.ReplaceAll(amountStr, "$", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	cleanStr = strings.TrimSpace(cleanStr)

	// Check if it's in millions
	if strings.Contains(strings.ToLower(cleanStr), "million") ||
		strings.Contains(strings.ToLower(cleanStr), "mil") ||
		strings.Contains(strings.ToLower(cleanStr), "m") {
		// Extract numeric part
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(cleanStr)
		if len(matches) > 1 {
			amount, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			return int64(amount * 1000000), nil // Convert to dollars
		}
	}

	// Check if it's in thousands (K)
	if strings.Contains(strings.ToLower(cleanStr), "thousand") ||
		strings.Contains(strings.ToLower(cleanStr), "k") {
		// Extract numeric part
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(cleanStr)
		if len(matches) > 1 {
			amount, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			return int64(amount * 1000), nil // Convert to dollars
		}
	}

	// Try parsing as direct amount
	amount, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0, err
	}

	return int64(amount), nil
}

// extractRevenueRange extracts revenue range from content
func (ra *RevenueAnalyzer) extractRevenueRange(content string, result *RevenueResult) {
	// Revenue range keywords
	revenueRanges := map[string]int64{
		"under $1 million":  500000,
		"under $1m":         500000,
		"under $1 mil":      500000,
		"$1-5 million":      3000000,
		"$1-5m":             3000000,
		"$1-5 mil":          3000000,
		"$5-10 million":     7500000,
		"$5-10m":            7500000,
		"$5-10 mil":         7500000,
		"$10-50 million":    30000000,
		"$10-50m":           30000000,
		"$10-50 mil":        30000000,
		"$50-100 million":   75000000,
		"$50-100m":          75000000,
		"$50-100 mil":       75000000,
		"over $100 million": 150000000,
		"over $100m":        150000000,
		"over $100 mil":     150000000,
	}

	contentLower := strings.ToLower(content)
	for rangeStr, amount := range revenueRanges {
		if strings.Contains(contentLower, rangeStr) {
			result.RevenueAmount = amount
			result.ExtractionMethod = "revenue_range"
			result.Evidence = append(result.Evidence, fmt.Sprintf("Revenue range: %s", rangeStr))
			result.ExtractedPhrases = append(result.ExtractedPhrases, rangeStr)
			break
		}
	}
}

// extractFinancialIndicators extracts financial health indicators
func (ra *RevenueAnalyzer) extractFinancialIndicators(content string, result *RevenueResult) {
	// Financial health keywords
	financialIndicators := map[string]int64{
		"profitable":            5000000,
		"profitable company":    5000000,
		"profitable business":   5000000,
		"profitable startup":    2000000,
		"profitable sme":        3000000,
		"profitable enterprise": 50000000,
		"growing revenue":       3000000,
		"revenue growth":        3000000,
		"increasing revenue":    3000000,
		"strong revenue":        5000000,
		"healthy revenue":       5000000,
		"stable revenue":        4000000,
		"consistent revenue":    4000000,
		"annual growth":         3000000,
		"year over year":        3000000,
		"yoy growth":            3000000,
	}

	contentLower := strings.ToLower(content)
	for indicator, amount := range financialIndicators {
		if strings.Contains(contentLower, indicator) {
			if result.RevenueAmount == 0 {
				result.RevenueAmount = amount
				result.ExtractionMethod = "financial_indicator"
				result.Evidence = append(result.Evidence, fmt.Sprintf("Financial indicator: %s", indicator))
				result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
			}
		}
	}
}

// analyzeFinancialHealth analyzes financial health indicators
func (ra *RevenueAnalyzer) analyzeFinancialHealth(ctx context.Context, content string, result *RevenueResult) error {
	ctx, span := ra.tracer.Start(ctx, "revenue_analyzer.analyze_financial_health")
	defer span.End()

	// Positive financial indicators
	positiveIndicators := []string{
		"profitable", "profitable company", "profitable business",
		"growing revenue", "revenue growth", "increasing revenue",
		"strong revenue", "healthy revenue", "stable revenue",
		"consistent revenue", "annual growth", "year over year",
		"yoy growth", "financial stability", "strong financials",
		"healthy financials", "stable financials", "consistent financials",
		"positive cash flow", "positive ebitda", "positive earnings",
		"strong balance sheet", "healthy balance sheet", "stable balance sheet",
	}

	// Negative financial indicators
	negativeIndicators := []string{
		"loss", "losing money", "negative revenue", "declining revenue",
		"decreasing revenue", "weak revenue", "poor revenue", "unstable revenue",
		"inconsistent revenue", "financial loss", "negative cash flow",
		"negative ebitda", "negative earnings", "weak balance sheet",
		"poor balance sheet", "unstable balance sheet", "financial instability",
		"cash flow problems", "financial difficulties", "financial challenges",
	}

	contentLower := strings.ToLower(content)
	positiveCount := 0
	negativeCount := 0

	for _, indicator := range positiveIndicators {
		if strings.Contains(contentLower, indicator) {
			positiveCount++
			result.Evidence = append(result.Evidence, fmt.Sprintf("Positive indicator: %s", indicator))
		}
	}

	for _, indicator := range negativeIndicators {
		if strings.Contains(contentLower, indicator) {
			negativeCount++
			result.Evidence = append(result.Evidence, fmt.Sprintf("Negative indicator: %s", indicator))
		}
	}

	// Determine financial health
	if positiveCount > negativeCount {
		result.FinancialHealth = "healthy"
		result.HealthConfidence = float64(positiveCount) / float64(positiveCount+negativeCount+1)
	} else if negativeCount > positiveCount {
		result.FinancialHealth = "unhealthy"
		result.HealthConfidence = float64(negativeCount) / float64(positiveCount+negativeCount+1)
	} else {
		result.FinancialHealth = "neutral"
		result.HealthConfidence = 0.5
	}

	// Classify health category
	result.HealthCategory = ra.classifyFinancialHealth(result.RevenueAmount, result.FinancialHealth)

	return nil
}

// classifyFinancialHealth classifies financial health based on revenue and indicators
func (ra *RevenueAnalyzer) classifyFinancialHealth(revenue int64, health string) string {
	if health == "unhealthy" {
		return "at_risk"
	}

	if revenue == 0 {
		return "unknown"
	}

	if revenue < ra.config.StartupRevenueThreshold {
		return "startup"
	} else if revenue < ra.config.SMEMaxRevenueThreshold {
		return "sme"
	} else {
		return "enterprise"
	}
}

// calculateConfidence calculates confidence score for the analysis
func (ra *RevenueAnalyzer) calculateConfidence(result *RevenueResult) float64 {
	confidence := 0.0

	// Base confidence from extraction method
	switch result.ExtractionMethod {
	case "direct_mention":
		confidence += 0.4
	case "revenue_range":
		confidence += 0.3
	case "financial_indicator":
		confidence += 0.2
	default:
		confidence += 0.1
	}

	// Evidence quality
	if len(result.Evidence) > 0 {
		evidenceScore := float64(len(result.Evidence)) * 0.1
		if evidenceScore > 0.3 {
			evidenceScore = 0.3
		}
		confidence += evidenceScore
	}

	// Financial health confidence
	if result.HealthConfidence > 0 {
		confidence += result.HealthConfidence * 0.2
	}

	// Validation status
	if result.IsValidated {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// validateResult validates the analysis result
func (ra *RevenueAnalyzer) validateResult(result *RevenueResult) error {
	// Check minimum confidence threshold
	if result.ConfidenceScore < ra.config.MinConfidenceThreshold {
		return fmt.Errorf("confidence score %f below threshold %f",
			result.ConfidenceScore, ra.config.MinConfidenceThreshold)
	}

	// Validate revenue amount
	if result.RevenueAmount < 0 {
		return fmt.Errorf("invalid revenue amount: %d", result.RevenueAmount)
	}

	// Validate financial health
	if result.FinancialHealth != "" &&
		result.FinancialHealth != "healthy" &&
		result.FinancialHealth != "unhealthy" &&
		result.FinancialHealth != "neutral" {
		return fmt.Errorf("invalid financial health: %s", result.FinancialHealth)
	}

	result.IsValidated = true
	return nil
}

// generateReasoning generates human-readable reasoning for the analysis
func (ra *RevenueAnalyzer) generateReasoning(result *RevenueResult) string {
	reasoning := ""

	if result.RevenueAmount > 0 {
		reasoning += fmt.Sprintf("Revenue analysis identified $%d in annual revenue. ", result.RevenueAmount)
	}

	if result.ExtractionMethod != "" {
		reasoning += fmt.Sprintf("This was extracted using %s method. ", result.ExtractionMethod)
	}

	if result.FinancialHealth != "" {
		reasoning += fmt.Sprintf("Financial health indicators suggest a %s financial position. ", result.FinancialHealth)
	}

	if result.HealthCategory != "" {
		reasoning += fmt.Sprintf("The company appears to be in the %s category. ", result.HealthCategory)
	}

	if len(result.Evidence) > 0 {
		reasoning += fmt.Sprintf("Key evidence includes: %s. ", strings.Join(result.Evidence[:min(3, len(result.Evidence))], ", "))
	}

	reasoning += fmt.Sprintf("Overall confidence in this analysis is %d%%. ", int(result.ConfidenceScore*100))

	return reasoning
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
