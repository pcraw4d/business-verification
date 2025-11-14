package risk

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/external"
	"kyb-platform/internal/modules/website_analysis"

	"go.uber.org/zap"
)

// RiskDetectionService provides comprehensive risk detection capabilities
// that integrate with existing website scraping and analysis infrastructure
type RiskDetectionService struct {
	db             database.Database
	logger         *zap.Logger
	websiteScraper *external.WebsiteScraper
	analysisModule *website_analysis.WebsiteAnalysisModule

	// Risk detection components
	keywordMatcher  *RiskKeywordMatcher
	riskScorer      *RiskScorer
	patternDetector *RiskPatternDetector

	// Caching for performance
	riskKeywordsCache map[string][]RiskKeyword
	cacheMutex        sync.RWMutex
	cacheTTL          time.Duration
	lastCacheUpdate   time.Time

	// Configuration
	config *RiskDetectionConfig
}

// RiskDetectionConfig contains configuration for risk detection
type RiskDetectionConfig struct {
	EnableWebsiteScraping  bool          `json:"enable_website_scraping"`
	EnableContentAnalysis  bool          `json:"enable_content_analysis"`
	EnablePatternDetection bool          `json:"enable_pattern_detection"`
	MaxConcurrentRequests  int           `json:"max_concurrent_requests"`
	RequestTimeout         time.Duration `json:"request_timeout"`
	CacheTTL               time.Duration `json:"cache_ttl"`
	MinConfidenceThreshold float64       `json:"min_confidence_threshold"`
	HighRiskThreshold      float64       `json:"high_risk_threshold"`
	CriticalRiskThreshold  float64       `json:"critical_risk_threshold"`
	EnableDetailedLogging  bool          `json:"enable_detailed_logging"`
	MaxContentLength       int           `json:"max_content_length"`
	EnableRegexPatterns    bool          `json:"enable_regex_patterns"`
	EnableSynonymMatching  bool          `json:"enable_synonym_matching"`
}

// DefaultRiskDetectionConfig returns default configuration for risk detection
func DefaultRiskDetectionConfig() *RiskDetectionConfig {
	return &RiskDetectionConfig{
		EnableWebsiteScraping:  true,
		EnableContentAnalysis:  true,
		EnablePatternDetection: true,
		MaxConcurrentRequests:  10,
		RequestTimeout:         30 * time.Second,
		CacheTTL:               1 * time.Hour,
		MinConfidenceThreshold: 0.3,
		HighRiskThreshold:      0.7,
		CriticalRiskThreshold:  0.9,
		EnableDetailedLogging:  true,
		MaxContentLength:       100000, // 100KB
		EnableRegexPatterns:    true,
		EnableSynonymMatching:  true,
	}
}

// Using existing RiskKeyword type from keywords_service.go

// RiskDetectionRequest represents a request for risk detection
type RiskDetectionRequest struct {
	BusinessID              string `json:"business_id"`
	BusinessName            string `json:"business_name"`
	BusinessDescription     string `json:"business_description,omitempty"`
	WebsiteURL              string `json:"website_url,omitempty"`
	IndustryCode            string `json:"industry_code,omitempty"`
	MCCCode                 string `json:"mcc_code,omitempty"`
	NAICSCode               string `json:"naics_code,omitempty"`
	SICCode                 string `json:"sic_code,omitempty"`
	IncludeWebsiteAnalysis  bool   `json:"include_website_analysis"`
	IncludeContentAnalysis  bool   `json:"include_content_analysis"`
	IncludePatternDetection bool   `json:"include_pattern_detection"`
}

// EnhancedRiskDetectionResult represents the enhanced result of risk detection
type EnhancedRiskDetectionResult struct {
	RequestID           string                        `json:"request_id"`
	BusinessID          string                        `json:"business_id"`
	BusinessName        string                        `json:"business_name"`
	OverallRiskScore    float64                       `json:"overall_risk_score"`
	OverallRiskLevel    RiskLevel                     `json:"overall_risk_level"`
	RiskCategories      map[string]RiskCategoryResult `json:"risk_categories"`
	DetectedKeywords    []DetectedRiskKeyword         `json:"detected_keywords"`
	WebsiteAnalysis     *WebsiteRiskAnalysis          `json:"website_analysis,omitempty"`
	ContentAnalysis     *ContentRiskAnalysis          `json:"content_analysis,omitempty"`
	PatternAnalysis     *PatternRiskAnalysis          `json:"pattern_analysis,omitempty"`
	Recommendations     []RiskRecommendation          `json:"recommendations"`
	Alerts              []RiskAlert                   `json:"alerts,omitempty"`
	AssessmentTimestamp time.Time                     `json:"assessment_timestamp"`
	ProcessingTime      time.Duration                 `json:"processing_time"`
	Metadata            map[string]interface{}        `json:"metadata,omitempty"`
}

// RiskCategoryResult represents risk analysis for a specific category
type RiskCategoryResult struct {
	Category        RiskCategory `json:"category"`
	Score           float64      `json:"score"`
	Level           RiskLevel    `json:"level"`
	DetectedCount   int          `json:"detected_count"`
	Keywords        []string     `json:"keywords"`
	Confidence      float64      `json:"confidence"`
	Recommendations []string     `json:"recommendations"`
}

// DetectedRiskKeyword represents a detected risk keyword with context
type DetectedRiskKeyword struct {
	Keyword          string    `json:"keyword"`
	Category         string    `json:"category"`
	Severity         string    `json:"severity"`
	Confidence       float64   `json:"confidence"`
	Context          string    `json:"context"`
	Source           string    `json:"source"` // "business_name", "description", "website_content", "pattern"
	Position         int       `json:"position"`
	MCCCodes         []string  `json:"mcc_codes"`
	CardRestrictions []string  `json:"card_restrictions"`
	DetectedAt       time.Time `json:"detected_at"`
}

// WebsiteRiskAnalysis represents risk analysis of website content
type WebsiteRiskAnalysis struct {
	URL               string                `json:"url"`
	StatusCode        int                   `json:"status_code"`
	ContentLength     int64                 `json:"content_length"`
	DetectedKeywords  []DetectedRiskKeyword `json:"detected_keywords"`
	RiskScore         float64               `json:"risk_score"`
	RiskLevel         RiskLevel             `json:"risk_level"`
	ContentQuality    float64               `json:"content_quality"`
	SecurityIssues    []string              `json:"security_issues,omitempty"`
	AnalysisTimestamp time.Time             `json:"analysis_timestamp"`
	ProcessingTime    time.Duration         `json:"processing_time"`
}

// ContentRiskAnalysis represents risk analysis of business content
type ContentRiskAnalysis struct {
	BusinessNameKeywords []DetectedRiskKeyword `json:"business_name_keywords"`
	DescriptionKeywords  []DetectedRiskKeyword `json:"description_keywords"`
	IndustryCodeKeywords []DetectedRiskKeyword `json:"industry_code_keywords"`
	OverallRiskScore     float64               `json:"overall_risk_score"`
	OverallRiskLevel     RiskLevel             `json:"overall_risk_level"`
	AnalysisTimestamp    time.Time             `json:"analysis_timestamp"`
	ProcessingTime       time.Duration         `json:"processing_time"`
}

// PatternRiskAnalysis represents pattern-based risk analysis
type PatternRiskAnalysis struct {
	DetectedPatterns  []DetectedPattern `json:"detected_patterns"`
	PatternRiskScore  float64           `json:"pattern_risk_score"`
	PatternRiskLevel  RiskLevel         `json:"pattern_risk_level"`
	AnalysisTimestamp time.Time         `json:"analysis_timestamp"`
	ProcessingTime    time.Duration     `json:"processing_time"`
}

// DetectedPattern represents a detected risk pattern
type DetectedPattern struct {
	PatternName string    `json:"pattern_name"`
	PatternType string    `json:"pattern_type"`
	Confidence  float64   `json:"confidence"`
	Context     string    `json:"context"`
	Source      string    `json:"source"`
	DetectedAt  time.Time `json:"detected_at"`
}

// NewRiskDetectionService creates a new risk detection service
func NewRiskDetectionService(
	db database.Database,
	logger *zap.Logger,
	websiteScraper *external.WebsiteScraper,
	analysisModule *website_analysis.WebsiteAnalysisModule,
	config *RiskDetectionConfig,
) *RiskDetectionService {
	if config == nil {
		config = DefaultRiskDetectionConfig()
	}

	service := &RiskDetectionService{
		db:                db,
		logger:            logger,
		websiteScraper:    websiteScraper,
		analysisModule:    analysisModule,
		riskKeywordsCache: make(map[string][]RiskKeyword),
		cacheTTL:          config.CacheTTL,
		config:            config,
	}

	// Initialize risk detection components
	service.keywordMatcher = NewRiskKeywordMatcher(logger, config)
	service.riskScorer = NewRiskScorer(logger, config)
	service.patternDetector = NewRiskPatternDetector(logger, config)

	return service
}

// DetectRisk performs comprehensive risk detection for a business
func (rds *RiskDetectionService) DetectRisk(ctx context.Context, req *RiskDetectionRequest) (*EnhancedRiskDetectionResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	rds.logger.Info("Starting risk detection",
		zap.String("request_id", requestID),
		zap.String("business_id", req.BusinessID),
		zap.String("business_name", req.BusinessName))

	// Initialize result
	result := &EnhancedRiskDetectionResult{
		RequestID:           requestID,
		BusinessID:          req.BusinessID,
		BusinessName:        req.BusinessName,
		RiskCategories:      make(map[string]RiskCategoryResult),
		DetectedKeywords:    make([]DetectedRiskKeyword, 0),
		Recommendations:     make([]RiskRecommendation, 0),
		Alerts:              make([]RiskAlert, 0),
		AssessmentTimestamp: time.Now(),
		Metadata:            make(map[string]interface{}),
	}

	// Load risk keywords from database
	riskKeywords, err := rds.loadRiskKeywords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load risk keywords: %w", err)
	}

	// Perform content-based risk analysis
	if req.IncludeContentAnalysis {
		contentAnalysis, err := rds.performContentRiskAnalysis(ctx, req, riskKeywords)
		if err != nil {
			rds.logger.Warn("Content risk analysis failed", zap.Error(err))
		} else {
			result.ContentAnalysis = contentAnalysis
			result.DetectedKeywords = append(result.DetectedKeywords, contentAnalysis.BusinessNameKeywords...)
			result.DetectedKeywords = append(result.DetectedKeywords, contentAnalysis.DescriptionKeywords...)
			result.DetectedKeywords = append(result.DetectedKeywords, contentAnalysis.IndustryCodeKeywords...)
		}
	}

	// Perform website-based risk analysis
	if req.IncludeWebsiteAnalysis && req.WebsiteURL != "" {
		websiteAnalysis, err := rds.performWebsiteRiskAnalysis(ctx, req, riskKeywords)
		if err != nil {
			rds.logger.Warn("Website risk analysis failed", zap.Error(err))
		} else {
			result.WebsiteAnalysis = websiteAnalysis
			result.DetectedKeywords = append(result.DetectedKeywords, websiteAnalysis.DetectedKeywords...)
		}
	}

	// Perform pattern-based risk analysis
	if req.IncludePatternDetection {
		patternAnalysis, err := rds.performPatternRiskAnalysis(ctx, req, riskKeywords)
		if err != nil {
			rds.logger.Warn("Pattern risk analysis failed", zap.Error(err))
		} else {
			result.PatternAnalysis = patternAnalysis
		}
	}

	// Calculate overall risk score and level
	// TODO: Fix type mismatch - calculateOverallRisk expects *RiskDetectionResult but we have *EnhancedRiskDetectionResult
	// For now, calculate directly from result
	overallScore := 0.0
	if len(result.DetectedKeywords) > 0 {
		sourceWeights := map[string]float64{
			"business_name":    0.3,
			"description":      0.25,
			"website_content":   0.25,
			"pattern":          0.2,
		}
		totalWeight := 0.0
		for _, keyword := range result.DetectedKeywords {
			weight := sourceWeights[keyword.Source]
			if weight == 0 {
				weight = 0.1
			}
			severityScore := rds.getSeverityScore(keyword.Severity)
			score := severityScore * keyword.Confidence * weight
			overallScore += score
			totalWeight += weight
		}
		if totalWeight > 0 {
			overallScore = overallScore / totalWeight
		}
	}
	// Determine risk level based on score
	var overallLevel RiskLevel
	switch {
	case overallScore >= 0.8:
		overallLevel = RiskLevelCritical
	case overallScore >= 0.6:
		overallLevel = RiskLevelHigh
	case overallScore >= 0.4:
		overallLevel = RiskLevelMedium
	case overallScore >= 0.2:
		overallLevel = RiskLevelLow
	default:
		overallLevel = RiskLevelMinimal
	}
	result.OverallRiskScore = overallScore
	result.OverallRiskLevel = overallLevel

	// Generate risk categories summary
	result.RiskCategories = rds.generateRiskCategories(result.DetectedKeywords)

	// Generate recommendations and alerts
	// TODO: Fix type mismatch - these functions expect *RiskDetectionResult but we have *EnhancedRiskDetectionResult
	// For now, generate empty lists
	result.Recommendations = []RiskRecommendation{}
	result.Alerts = []RiskAlert{}

	// Set processing time
	result.ProcessingTime = time.Since(startTime)

	// Log completion
	rds.logger.Info("Risk detection completed",
		zap.String("request_id", requestID),
		zap.Float64("overall_risk_score", result.OverallRiskScore),
		zap.String("overall_risk_level", string(result.OverallRiskLevel)),
		zap.Int("detected_keywords", len(result.DetectedKeywords)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// loadRiskKeywords loads risk keywords from database with caching
func (rds *RiskDetectionService) loadRiskKeywords(ctx context.Context) ([]RiskKeyword, error) {
	// Check cache first
	rds.cacheMutex.RLock()
	if time.Since(rds.lastCacheUpdate) < rds.cacheTTL && len(rds.riskKeywordsCache["all"]) > 0 {
		keywords := rds.riskKeywordsCache["all"]
		rds.cacheMutex.RUnlock()
		return keywords, nil
	}
	rds.cacheMutex.RUnlock()

	// TODO: database.Database interface doesn't have QueryContext method
	// Need to check the actual database interface and use the correct method
	// For now, return empty list - this needs proper implementation
	return []RiskKeyword{}, nil
	// The following code is commented out until we can properly query the database
	// query := `
	// 	SELECT id, keyword, risk_category, risk_severity, description,
	// 	       mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
	// 	       detection_patterns, synonyms, is_active, created_at, updated_at
	// 	FROM risk_keywords 
	// 	WHERE is_active = true
	// 	ORDER BY risk_severity DESC, risk_category
	// `
	// rows, err := rds.db.QueryContext(ctx, query)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to query risk keywords: %w", err)
	// }
	// defer rows.Close()
	// ... rest of the code ...

	// Update cache (stub - no keywords loaded)
	rds.cacheMutex.Lock()
	rds.riskKeywordsCache["all"] = []RiskKeyword{}
	rds.lastCacheUpdate = time.Now()
	rds.cacheMutex.Unlock()

	rds.logger.Info("Loaded risk keywords from database",
		zap.Int("count", 0)) // Stub - no keywords loaded

	return []RiskKeyword{}, nil
}

// performContentRiskAnalysis performs risk analysis on business content
func (rds *RiskDetectionService) performContentRiskAnalysis(
	ctx context.Context,
	req *RiskDetectionRequest,
	riskKeywords []RiskKeyword,
) (*ContentRiskAnalysis, error) {
	startTime := time.Now()

	analysis := &ContentRiskAnalysis{
		BusinessNameKeywords: make([]DetectedRiskKeyword, 0),
		DescriptionKeywords:  make([]DetectedRiskKeyword, 0),
		IndustryCodeKeywords: make([]DetectedRiskKeyword, 0),
		AnalysisTimestamp:    time.Now(),
	}

	// Analyze business name
	if req.BusinessName != "" {
		nameKeywords := rds.keywordMatcher.MatchKeywords(req.BusinessName, riskKeywords, "business_name")
		analysis.BusinessNameKeywords = nameKeywords
	}

	// Analyze business description
	if req.BusinessDescription != "" {
		descKeywords := rds.keywordMatcher.MatchKeywords(req.BusinessDescription, riskKeywords, "description")
		analysis.DescriptionKeywords = descKeywords
	}

	// Analyze industry codes
	industryKeywords := rds.analyzeIndustryCodes(req, riskKeywords)
	analysis.IndustryCodeKeywords = industryKeywords

	// Calculate overall risk score for content
	allKeywords := append(analysis.BusinessNameKeywords, analysis.DescriptionKeywords...)
	allKeywords = append(allKeywords, analysis.IndustryCodeKeywords...)

	analysis.OverallRiskScore, analysis.OverallRiskLevel = rds.riskScorer.CalculateRiskScore(allKeywords)
	analysis.ProcessingTime = time.Since(startTime)

	return analysis, nil
}

// performWebsiteRiskAnalysis performs risk analysis on website content
func (rds *RiskDetectionService) performWebsiteRiskAnalysis(
	ctx context.Context,
	req *RiskDetectionRequest,
	riskKeywords []RiskKeyword,
) (*WebsiteRiskAnalysis, error) {
	startTime := time.Now()

	analysis := &WebsiteRiskAnalysis{
		URL:               req.WebsiteURL,
		DetectedKeywords:  make([]DetectedRiskKeyword, 0),
		AnalysisTimestamp: time.Now(),
	}

	// Use existing website scraper to get content
	scrapingResult, err := rds.websiteScraper.ScrapeWebsite(ctx, req.WebsiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape website: %w", err)
	}

	analysis.StatusCode = scrapingResult.StatusCode
	analysis.ContentLength = scrapingResult.ContentLength

	// Extract text content for analysis
	content := rds.extractTextContent(scrapingResult.Content)
	if len(content) > rds.config.MaxContentLength {
		content = content[:rds.config.MaxContentLength]
	}

	// Match risk keywords against website content
	websiteKeywords := rds.keywordMatcher.MatchKeywords(content, riskKeywords, "website_content")
	analysis.DetectedKeywords = websiteKeywords

	// Calculate risk score for website
	analysis.RiskScore, analysis.RiskLevel = rds.riskScorer.CalculateRiskScore(websiteKeywords)

	// Calculate content quality (basic implementation)
	analysis.ContentQuality = rds.calculateContentQuality(content)

	analysis.ProcessingTime = time.Since(startTime)

	return analysis, nil
}

// performPatternRiskAnalysis performs pattern-based risk analysis
func (rds *RiskDetectionService) performPatternRiskAnalysis(
	ctx context.Context,
	req *RiskDetectionRequest,
	riskKeywords []RiskKeyword,
) (*PatternRiskAnalysis, error) {
	startTime := time.Now()

	analysis := &PatternRiskAnalysis{
		DetectedPatterns:  make([]DetectedPattern, 0),
		AnalysisTimestamp: time.Now(),
	}

	// Use pattern detector to find risk patterns
	patterns := rds.patternDetector.DetectPatterns(req, riskKeywords)
	analysis.DetectedPatterns = patterns

	// Calculate pattern-based risk score
	analysis.PatternRiskScore, analysis.PatternRiskLevel = rds.riskScorer.CalculatePatternRiskScore(patterns)

	analysis.ProcessingTime = time.Since(startTime)

	return analysis, nil
}

// analyzeIndustryCodes analyzes industry codes for risk indicators
func (rds *RiskDetectionService) analyzeIndustryCodes(req *RiskDetectionRequest, riskKeywords []RiskKeyword) []DetectedRiskKeyword {
	var detectedKeywords []DetectedRiskKeyword

	// Check MCC code against risk keywords
	if req.MCCCode != "" {
		for _, keyword := range riskKeywords {
			for _, mccCode := range keyword.MCCCodes {
				if strings.EqualFold(mccCode, req.MCCCode) {
					detected := DetectedRiskKeyword{
						Keyword:          keyword.Keyword,
						Category:         keyword.RiskCategory,
						Severity:         keyword.RiskSeverity,
						Confidence:       1.0, // Direct match
						Context:          fmt.Sprintf("MCC Code: %s", req.MCCCode),
						Source:           "mcc_code",
						MCCCodes:         []string{req.MCCCode},
						CardRestrictions: keyword.CardBrandRestrictions,
						DetectedAt:       time.Now(),
					}
					detectedKeywords = append(detectedKeywords, detected)
				}
			}
		}
	}

	// Similar logic for NAICS and SIC codes
	if req.NAICSCode != "" {
		for _, keyword := range riskKeywords {
			for _, naicsCode := range keyword.NAICSCodes {
				if strings.EqualFold(naicsCode, req.NAICSCode) {
					detected := DetectedRiskKeyword{
						Keyword:    keyword.Keyword,
						Category:   keyword.RiskCategory,
						Severity:   keyword.RiskSeverity,
						Confidence: 1.0,
						Context:    fmt.Sprintf("NAICS Code: %s", req.NAICSCode),
						Source:     "naics_code",
						DetectedAt: time.Now(),
					}
					detectedKeywords = append(detectedKeywords, detected)
				}
			}
		}
	}

	if req.SICCode != "" {
		for _, keyword := range riskKeywords {
			for _, sicCode := range keyword.SICCodes {
				if strings.EqualFold(sicCode, req.SICCode) {
					detected := DetectedRiskKeyword{
						Keyword:    keyword.Keyword,
						Category:   keyword.RiskCategory,
						Severity:   keyword.RiskSeverity,
						Confidence: 1.0,
						Context:    fmt.Sprintf("SIC Code: %s", req.SICCode),
						Source:     "sic_code",
						DetectedAt: time.Now(),
					}
					detectedKeywords = append(detectedKeywords, detected)
				}
			}
		}
	}

	return detectedKeywords
}

// calculateOverallRisk calculates the overall risk score and level
func (rds *RiskDetectionService) calculateOverallRisk(result *RiskDetectionResult) (float64, RiskLevel) {
	var totalScore float64
	var totalWeight float64

	// Calculate weighted score from detected keywords
	// TODO: RiskDetectionResult.DetectedKeywords is []string, not []DetectedRiskKeyword
	// This function needs to be updated to work with the actual RiskDetectionResult type
	// For now, return default values
	if len(result.DetectedKeywords) > 0 {
		// Stub: assume average risk if keywords detected
		totalScore = 0.5
		totalWeight = 1.0
	}

	// Normalize score
	var overallScore float64
	if totalWeight > 0 {
		overallScore = totalScore / totalWeight
	}

	// Determine risk level
	var overallLevel RiskLevel
	switch {
	case overallScore >= rds.config.CriticalRiskThreshold:
		overallLevel = RiskLevelCritical
	case overallScore >= rds.config.HighRiskThreshold:
		overallLevel = RiskLevelHigh
	case overallScore >= 0.5:
		overallLevel = RiskLevelMedium
	case overallScore >= 0.2:
		overallLevel = RiskLevelLow
	default:
		overallLevel = RiskLevelMinimal
	}

	return overallScore, overallLevel
}

// generateRiskCategories generates risk category summary
func (rds *RiskDetectionService) generateRiskCategories(keywords []DetectedRiskKeyword) map[string]RiskCategoryResult {
	categories := make(map[string]RiskCategoryResult)

	// Group keywords by category
	categoryKeywords := make(map[string][]DetectedRiskKeyword)
	for _, keyword := range keywords {
		categoryKeywords[keyword.Category] = append(categoryKeywords[keyword.Category], keyword)
	}

	// Calculate category scores
	for category, keywords := range categoryKeywords {
		var totalScore float64
		var totalConfidence float64

		for _, keyword := range keywords {
			severityScore := rds.getSeverityScore(keyword.Severity)
			score := severityScore * keyword.Confidence
			totalScore += score
			totalConfidence += keyword.Confidence
		}

		avgScore := totalScore / float64(len(keywords))
		avgConfidence := totalConfidence / float64(len(keywords))

		// Determine category level
		var level RiskLevel
		switch {
		case avgScore >= 0.9:
			level = RiskLevelCritical
		case avgScore >= 0.7:
			level = RiskLevelHigh
		case avgScore >= 0.5:
			level = RiskLevelMedium
		case avgScore >= 0.2:
			level = RiskLevelLow
		default:
			level = RiskLevelMinimal
		}

		// Extract keyword names
		keywordNames := make([]string, len(keywords))
		for i, keyword := range keywords {
			keywordNames[i] = keyword.Keyword
		}

		categories[category] = RiskCategoryResult{
			Category:        RiskCategory(category),
			Score:           avgScore,
			Level:           level,
			DetectedCount:   len(keywords),
			Keywords:        keywordNames,
			Confidence:      avgConfidence,
			Recommendations: rds.generateCategoryRecommendations(category, level),
		}
	}

	return categories
}

// generateRecommendations generates risk mitigation recommendations
// generateRecommendations generates recommendations from RiskDetectionResult
// TODO: This function expects *RiskDetectionResult but is being called with *EnhancedRiskDetectionResult
func (rds *RiskDetectionService) generateRecommendations(result *RiskDetectionResult) []RiskRecommendation {
	var recommendations []RiskRecommendation

	// Generate recommendations based on risk level
	// TODO: RiskDetectionResult has RiskLevel string, not OverallRiskLevel RiskLevel
	// This function needs to be updated to work with the actual RiskDetectionResult type
	if result.RiskLevel == "critical" {
		recommendations = append(recommendations, RiskRecommendation{
			ID:          generateRequestID(),
			RiskFactor:  "overall_risk",
			Title:       "Immediate Risk Review Required",
			Description: "Critical risk indicators detected. Immediate manual review and potential business rejection recommended.",
			Priority:    RiskLevelCritical,
			Action:      "Manual review and risk assessment",
			Impact:      "High - Business may be rejected",
			Timeline:    "Immediate",
			CreatedAt:   time.Now(),
		})
	}

	// Generate category-specific recommendations
	// TODO: RiskDetectionResult.RiskCategories is []string, not map[string]RiskCategoryResult
	// This function needs to be updated to work with the actual RiskDetectionResult type
	for _, category := range result.RiskCategories {
		recommendations = append(recommendations, RiskRecommendation{
			ID:          generateRequestID(),
			RiskFactor:  category,
			Title:       fmt.Sprintf("Risk in %s Category", category),
			Description: fmt.Sprintf("Risk indicators detected in %s category. Additional verification required.", category),
			Priority:    RiskLevelMedium, // Stub - default priority
			Action:      "Additional verification and documentation",
			Impact:      "Medium - Additional review required",
			Timeline:    "Within 24 hours",
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}

// generateAlerts generates risk alerts
// generateAlerts generates alerts from RiskDetectionResult
// TODO: This function expects *RiskDetectionResult but is being called with *EnhancedRiskDetectionResult
func (rds *RiskDetectionService) generateAlerts(result *RiskDetectionResult) []RiskAlert {
	var alerts []RiskAlert

	// Generate alerts for critical risks
	// TODO: RiskDetectionResult has RiskLevel string, not OverallRiskLevel RiskLevel
	// This function needs to be updated to work with the actual RiskDetectionResult type
	if result.RiskLevel == "critical" {
		alerts = append(alerts, RiskAlert{
			ID:          generateRequestID(),
			BusinessID:  "", // TODO: RiskDetectionResult doesn't have BusinessID field
			RiskFactor:  "overall_risk",
			Level:       RiskLevelCritical,
			Message:     "Critical risk indicators detected requiring immediate attention",
			Score:       1.0,
			Threshold:   0.8,
			TriggeredAt: time.Now(),
			Acknowledged: false,
		})
	}

	// Generate alerts for high-risk categories
	// TODO: RiskDetectionResult.RiskCategories is []string, not map[string]RiskCategoryResult
	// This function needs to be updated to work with the actual RiskDetectionResult type
	for _, category := range result.RiskCategories {
		alerts = append(alerts, RiskAlert{
			ID:          generateRequestID(),
			BusinessID:  "", // TODO: RiskDetectionResult doesn't have BusinessID field
			RiskFactor:  category,
			Level:       RiskLevelMedium, // Stub - default level
			Message:     fmt.Sprintf("Risk indicators detected in %s category.", category),
			Score:       0.5,
			Threshold:   0.4,
			TriggeredAt: time.Now(),
			Acknowledged: false,
		})
	}

	return alerts
}

// Helper functions

func (rds *RiskDetectionService) getSeverityScore(severity string) float64 {
	switch strings.ToLower(severity) {
	case "critical":
		return 1.0
	case "high":
		return 0.8
	case "medium":
		return 0.6
	case "low":
		return 0.4
	default:
		return 0.2
	}
}

func (rds *RiskDetectionService) extractTextContent(htmlContent string) string {
	// Basic HTML tag removal - in production, use a proper HTML parser
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, " ")

	// Clean up whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func (rds *RiskDetectionService) calculateContentQuality(content string) float64 {
	// Basic content quality calculation
	// In production, this would be more sophisticated
	length := len(content)
	if length == 0 {
		return 0.0
	}

	// Simple quality score based on content length and structure
	quality := 0.5 // Base quality

	if length > 1000 {
		quality += 0.2
	}
	if length > 5000 {
		quality += 0.2
	}
	if length > 10000 {
		quality += 0.1
	}

	// Check for common quality indicators
	if strings.Contains(content, "about") || strings.Contains(content, "contact") {
		quality += 0.1
	}

	if quality > 1.0 {
		quality = 1.0
	}

	return quality
}

func (rds *RiskDetectionService) generateCategoryRecommendations(category string, level RiskLevel) []string {
	var recommendations []string

	switch category {
	case "illegal":
		recommendations = append(recommendations, "Immediate business rejection recommended")
		recommendations = append(recommendations, "Report to compliance team")
	case "prohibited":
		recommendations = append(recommendations, "Card brand restrictions apply")
		recommendations = append(recommendations, "Additional verification required")
	case "high_risk":
		recommendations = append(recommendations, "Enhanced monitoring required")
		recommendations = append(recommendations, "Regular risk assessment")
	case "tbml":
		recommendations = append(recommendations, "Anti-money laundering review")
		recommendations = append(recommendations, "Enhanced due diligence")
	case "sanctions":
		recommendations = append(recommendations, "Sanctions screening required")
		recommendations = append(recommendations, "OFAC compliance check")
	case "fraud":
		recommendations = append(recommendations, "Fraud prevention measures")
		recommendations = append(recommendations, "Identity verification")
	}

	return recommendations
}

// Utility functions

func parseStringArray(arrayStr string) []string {
	if arrayStr == "" {
		return []string{}
	}

	// Remove curly braces and split by comma
	arrayStr = strings.Trim(arrayStr, "{}")
	if arrayStr == "" {
		return []string{}
	}

	parts := strings.Split(arrayStr, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func generateRequestID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}
