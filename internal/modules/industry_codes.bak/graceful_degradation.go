package industry_codes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kyb-platform/internal/integrations"
	"go.uber.org/zap"
)

// GracefulDegradationService provides graceful degradation strategies for industry code operations
type GracefulDegradationService struct {
	database          *IndustryCodeDatabase
	fallbackData      *FallbackDataProvider
	alternativeScorer *AlternativeScorer
	logger            *zap.Logger
	config            *DegradationConfig
}

// DegradationConfig defines configuration for graceful degradation
type DegradationConfig struct {
	EnableFallback         bool                  `json:"enable_fallback"`
	EnablePartialResults   bool                  `json:"enable_partial_results"`
	EnableCachedResults    bool                  `json:"enable_cached_results"`
	PartialResultThreshold float64               `json:"partial_result_threshold"`
	MinimalResultThreshold float64               `json:"minimal_result_threshold"`
	FallbackTimeout        time.Duration         `json:"fallback_timeout"`
	CacheTimeout           time.Duration         `json:"cache_timeout"`
	MaxPartialResults      int                   `json:"max_partial_results"`
	FallbackStrategies     []string              `json:"fallback_strategies"`
	QualityThresholds      map[string]float64    `json:"quality_thresholds"`
	NotificationSettings   *NotificationSettings `json:"notification_settings"`
}

// NotificationSettings defines notification settings for degradation events
type NotificationSettings struct {
	EnableNotifications bool     `json:"enable_notifications"`
	Recipients          []string `json:"recipients"`
	Channels            []string `json:"channels"`
	Severity            string   `json:"severity"`
}

// DegradationStrategy represents different degradation strategies
type DegradationStrategy string

const (
	StrategyFallbackData     DegradationStrategy = "fallback_data"
	StrategyCachedResults    DegradationStrategy = "cached_results"
	StrategyPartialResults   DegradationStrategy = "partial_results"
	StrategyAlternativeLogic DegradationStrategy = "alternative_logic"
	StrategyMinimalResponse  DegradationStrategy = "minimal_response"
	StrategyStaticResponse   DegradationStrategy = "static_response"
)

// DegradationLevel represents the level of service degradation
type DegradationLevel string

const (
	LevelNone     DegradationLevel = "none"
	LevelPartial  DegradationLevel = "partial"
	LevelMinimal  DegradationLevel = "minimal"
	LevelFallback DegradationLevel = "fallback"
	LevelCritical DegradationLevel = "critical"
)

// DegradationResult represents the result of a degraded operation
type DegradationResult struct {
	Success          bool                   `json:"success"`
	Data             interface{}            `json:"data"`
	DegradationLevel DegradationLevel       `json:"degradation_level"`
	Strategy         DegradationStrategy    `json:"strategy"`
	Confidence       float64                `json:"confidence"`
	QualityScore     float64                `json:"quality_score"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	Fallbacks        []FallbackAttempt      `json:"fallbacks"`
	Warnings         []string               `json:"warnings"`
	Recommendations  []string               `json:"recommendations"`
	Metadata         map[string]interface{} `json:"metadata"`
	Timestamp        time.Time              `json:"timestamp"`
}

// FallbackAttempt represents an attempt to use a fallback strategy
type FallbackAttempt struct {
	Strategy    DegradationStrategy `json:"strategy"`
	Success     bool                `json:"success"`
	Error       string              `json:"error,omitempty"`
	Duration    time.Duration       `json:"duration"`
	DataQuality float64             `json:"data_quality"`
}

// FallbackDataProvider provides fallback data for industry codes
type FallbackDataProvider struct {
	staticData   map[string]*IndustryCode
	cachedData   map[string]*CachedResult
	logger       *zap.Logger
	lastUpdated  time.Time
	cacheTimeout time.Duration
}

// CachedResult represents a cached industry code result
type CachedResult struct {
	Code       *IndustryCode `json:"code"`
	Confidence float64       `json:"confidence"`
	Timestamp  time.Time     `json:"timestamp"`
	Source     string        `json:"source"`
}

// AlternativeScorer provides alternative scoring logic when primary methods fail
type AlternativeScorer struct {
	simpleRules []SimpleRule
	logger      *zap.Logger
}

// SimpleRule represents a simple classification rule
type SimpleRule struct {
	Keywords    []string `json:"keywords"`
	Code        string   `json:"code"`
	Type        CodeType `json:"type"`
	Confidence  float64  `json:"confidence"`
	Description string   `json:"description"`
}

// NewGracefulDegradationService creates a new graceful degradation service
func NewGracefulDegradationService(
	database *IndustryCodeDatabase,
	logger *zap.Logger,
	config *DegradationConfig,
) *GracefulDegradationService {
	if config == nil {
		config = &DegradationConfig{
			EnableFallback:         true,
			EnablePartialResults:   true,
			EnableCachedResults:    true,
			PartialResultThreshold: 0.6,
			MinimalResultThreshold: 0.3,
			FallbackTimeout:        5 * time.Second,
			CacheTimeout:           1 * time.Hour,
			MaxPartialResults:      3,
			FallbackStrategies:     []string{"cached_results", "static_response", "partial_results"},
			QualityThresholds: map[string]float64{
				"excellent": 0.9,
				"good":      0.7,
				"fair":      0.5,
				"poor":      0.3,
			},
		}
	}

	fallbackData := &FallbackDataProvider{
		staticData:   make(map[string]*IndustryCode),
		cachedData:   make(map[string]*CachedResult),
		logger:       logger,
		lastUpdated:  time.Now(),
		cacheTimeout: config.CacheTimeout,
	}

	alternativeScorer := &AlternativeScorer{
		simpleRules: generateSimpleRules(),
		logger:      logger,
	}

	// Initialize static fallback data
	fallbackData.initializeStaticData()

	return &GracefulDegradationService{
		database:          database,
		fallbackData:      fallbackData,
		alternativeScorer: alternativeScorer,
		logger:            logger,
		config:            config,
	}
}

// ExecuteWithDegradation executes an industry code operation with graceful degradation
func (gds *GracefulDegradationService) ExecuteWithDegradation(
	ctx context.Context,
	operation func() (*IndustryCode, error),
	fallbackData interface{},
) *DegradationResult {
	start := time.Now()
	result := &DegradationResult{
		Timestamp: start,
		Fallbacks: []FallbackAttempt{},
		Warnings:  []string{},
		Metadata:  make(map[string]interface{}),
	}

	// Try primary operation first
	gds.logger.Info("attempting primary operation")
	data, err := operation()
	if err == nil && data != nil {
		result.Success = true
		result.Data = data
		result.DegradationLevel = LevelNone
		result.Strategy = DegradationStrategy("primary")
		result.Confidence = 1.0
		result.QualityScore = gds.calculateQualityScore(data)
		result.ProcessingTime = time.Since(start)
		return result
	}

	gds.logger.Warn("primary operation failed, attempting graceful degradation", zap.Error(err))

	// Try degradation strategies in order of preference
	strategies := gds.getDegradationStrategies()
	for _, strategy := range strategies {
		attempt := gds.attemptDegradationStrategy(ctx, strategy, fallbackData, err)
		result.Fallbacks = append(result.Fallbacks, attempt)

		if attempt.Success {
			result.Success = true
			result.Strategy = strategy
			result.DegradationLevel = gds.getDegradationLevel(strategy)
			result.Confidence = gds.calculateConfidence(strategy, attempt.DataQuality)
			result.QualityScore = attempt.DataQuality
			result.ProcessingTime = time.Since(start)
			result.Data = gds.getLastSuccessfulData()
			result.Warnings = gds.generateWarnings(strategy, attempt)
			result.Recommendations = gds.generateRecommendations(strategy, err)
			return result
		}
	}

	// All strategies failed
	gds.logger.Error("all degradation strategies failed")
	result.Success = false
	result.DegradationLevel = LevelCritical
	result.Strategy = DegradationStrategy("none")
	result.Confidence = 0.0
	result.QualityScore = 0.0
	result.ProcessingTime = time.Since(start)
	result.Warnings = append(result.Warnings, "All degradation strategies failed")
	result.Recommendations = []string{
		"Check system health and connectivity",
		"Verify database availability",
		"Review error logs for root cause",
		"Consider manual intervention",
	}

	return result
}

// getDegradationStrategies returns the ordered list of degradation strategies to try
func (gds *GracefulDegradationService) getDegradationStrategies() []DegradationStrategy {
	strategies := []DegradationStrategy{}

	if gds.config.EnableCachedResults {
		strategies = append(strategies, StrategyCachedResults)
	}

	if gds.config.EnableFallback {
		strategies = append(strategies, StrategyFallbackData)
	}

	if gds.config.EnablePartialResults {
		strategies = append(strategies, StrategyPartialResults)
	}

	strategies = append(strategies, StrategyAlternativeLogic, StrategyStaticResponse)

	return strategies
}

// attemptDegradationStrategy attempts a specific degradation strategy
func (gds *GracefulDegradationService) attemptDegradationStrategy(
	ctx context.Context,
	strategy DegradationStrategy,
	fallbackData interface{},
	originalError error,
) FallbackAttempt {
	start := time.Now()
	attempt := FallbackAttempt{
		Strategy: strategy,
		Success:  false,
	}

	gds.logger.Info("attempting degradation strategy", zap.String("strategy", string(strategy)))

	switch strategy {
	case StrategyCachedResults:
		attempt.Success, attempt.DataQuality = gds.tryCachedResults(ctx, fallbackData)
	case StrategyFallbackData:
		attempt.Success, attempt.DataQuality = gds.tryFallbackData(ctx, fallbackData)
	case StrategyPartialResults:
		attempt.Success, attempt.DataQuality = gds.tryPartialResults(ctx, fallbackData)
	case StrategyAlternativeLogic:
		attempt.Success, attempt.DataQuality = gds.tryAlternativeLogic(ctx, fallbackData)
	case StrategyStaticResponse:
		attempt.Success, attempt.DataQuality = gds.tryStaticResponse(ctx, fallbackData)
	default:
		attempt.Error = fmt.Sprintf("unknown strategy: %s", strategy)
	}

	attempt.Duration = time.Since(start)

	if !attempt.Success && attempt.Error == "" {
		attempt.Error = "strategy execution failed"
	}

	gds.logger.Info("degradation strategy completed",
		zap.String("strategy", string(strategy)),
		zap.Bool("success", attempt.Success),
		zap.Float64("data_quality", attempt.DataQuality),
		zap.Duration("duration", attempt.Duration))

	return attempt
}

// tryCachedResults attempts to use cached results
func (gds *GracefulDegradationService) tryCachedResults(ctx context.Context, fallbackData interface{}) (bool, float64) {
	if !gds.config.EnableCachedResults {
		return false, 0.0
	}

	// Try to extract query parameters for cache lookup
	query := gds.extractQueryFromFallbackData(fallbackData)
	if query == "" {
		return false, 0.0
	}

	// Check cache for recent results
	cached, exists := gds.fallbackData.cachedData[query]
	if !exists {
		return false, 0.0
	}

	// Check if cache is still valid
	if time.Since(cached.Timestamp) > gds.config.CacheTimeout {
		delete(gds.fallbackData.cachedData, query)
		return false, 0.0
	}

	gds.logger.Info("using cached result", zap.String("query", query))
	return true, cached.Confidence
}

// tryFallbackData attempts to use static fallback data
func (gds *GracefulDegradationService) tryFallbackData(ctx context.Context, fallbackData interface{}) (bool, float64) {
	if !gds.config.EnableFallback {
		return false, 0.0
	}

	query := gds.extractQueryFromFallbackData(fallbackData)
	if query == "" {
		return false, 0.0
	}

	// Try exact match first
	if code, exists := gds.fallbackData.staticData[query]; exists {
		gds.logger.Info("using static fallback data", zap.String("query", query))
		return true, code.Confidence
	}

	// Try partial matching
	for key, code := range gds.fallbackData.staticData {
		if strings.Contains(strings.ToLower(query), strings.ToLower(key)) ||
			strings.Contains(strings.ToLower(key), strings.ToLower(query)) {
			gds.logger.Info("using partial match fallback data",
				zap.String("query", query),
				zap.String("match", key))
			return true, code.Confidence * 0.8 // Reduced confidence for partial match
		}
	}

	return false, 0.0
}

// tryPartialResults attempts to provide partial results
func (gds *GracefulDegradationService) tryPartialResults(ctx context.Context, fallbackData interface{}) (bool, float64) {
	if !gds.config.EnablePartialResults {
		return false, 0.0
	}

	// Try to get partial data from alternative sources
	query := gds.extractQueryFromFallbackData(fallbackData)
	if query == "" {
		return false, 0.0
	}

	// Generate partial result based on simple heuristics
	partialCode := gds.generatePartialCode(query)
	if partialCode != nil {
		confidence := gds.config.PartialResultThreshold
		if confidence >= gds.config.MinimalResultThreshold {
			gds.logger.Info("providing partial result", zap.String("query", query))
			return true, confidence
		}
	}

	return false, 0.0
}

// tryAlternativeLogic attempts to use alternative scoring logic
func (gds *GracefulDegradationService) tryAlternativeLogic(ctx context.Context, fallbackData interface{}) (bool, float64) {
	query := gds.extractQueryFromFallbackData(fallbackData)
	if query == "" {
		return false, 0.0
	}

	// Use simple rule-based matching
	code := gds.alternativeScorer.scoreWithSimpleRules(query)
	if code != nil && code.Confidence >= gds.config.MinimalResultThreshold {
		gds.logger.Info("using alternative logic", zap.String("query", query))
		return true, code.Confidence
	}

	return false, 0.0
}

// tryStaticResponse attempts to provide a static response
func (gds *GracefulDegradationService) tryStaticResponse(ctx context.Context, fallbackData interface{}) (bool, float64) {
	// Check if we can meet the minimal threshold
	if gds.config.MinimalResultThreshold > 1.0 {
		return false, 0.0 // Impossible threshold
	}

	// Always succeeds with minimal quality
	gds.logger.Info("providing static response")
	return true, gds.config.MinimalResultThreshold
}

// Helper methods

// extractQueryFromFallbackData extracts query string from fallback data
func (gds *GracefulDegradationService) extractQueryFromFallbackData(fallbackData interface{}) string {
	if fallbackData == nil {
		return ""
	}

	switch data := fallbackData.(type) {
	case string:
		return data
	case map[string]interface{}:
		if query, ok := data["query"].(string); ok {
			return query
		}
		if name, ok := data["name"].(string); ok {
			return name
		}
		if business, ok := data["business_name"].(string); ok {
			return business
		}
	case integrations.BusinessData:
		return data.CompanyName
	}

	return ""
}

// generatePartialCode generates a partial industry code based on heuristics
func (gds *GracefulDegradationService) generatePartialCode(query string) *IndustryCode {
	// Simple heuristic-based code generation
	query = strings.ToLower(strings.TrimSpace(query))

	// Basic industry classifications
	if strings.Contains(query, "restaurant") || strings.Contains(query, "food") || strings.Contains(query, "dining") {
		return &IndustryCode{
			ID:          "fallback-food-001",
			Code:        "5812",
			Type:        CodeTypeMCC,
			Description: "Eating Places and Restaurants (Fallback)",
			Category:    "Food Service",
			Confidence:  0.6,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	if strings.Contains(query, "retail") || strings.Contains(query, "store") || strings.Contains(query, "shop") {
		return &IndustryCode{
			ID:          "fallback-retail-001",
			Code:        "5999",
			Type:        CodeTypeMCC,
			Description: "Miscellaneous and Specialty Retail Stores (Fallback)",
			Category:    "Retail",
			Confidence:  0.6,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	if strings.Contains(query, "consulting") || strings.Contains(query, "advisory") || strings.Contains(query, "professional") {
		return &IndustryCode{
			ID:          "fallback-professional-001",
			Code:        "541611",
			Type:        CodeTypeNAICS,
			Description: "Administrative Management and General Management Consulting Services (Fallback)",
			Category:    "Professional Services",
			Confidence:  0.6,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// Generic business classification
	return &IndustryCode{
		ID:          "fallback-generic-001",
		Code:        "9999",
		Type:        CodeTypeMCC,
		Description: "Miscellaneous Business (Fallback)",
		Category:    "General Business",
		Confidence:  0.4,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// calculateQualityScore calculates quality score for industry code data
func (gds *GracefulDegradationService) calculateQualityScore(code *IndustryCode) float64 {
	if code == nil {
		return 0.0
	}

	score := 0.0
	factors := 0

	// Check confidence
	if code.Confidence > 0 {
		score += code.Confidence
		factors++
	}

	// Check completeness
	completeness := 0.0
	if code.Code != "" {
		completeness += 0.3
	}
	if code.Description != "" {
		completeness += 0.3
	}
	if code.Category != "" {
		completeness += 0.2
	}
	if code.Type != "" {
		completeness += 0.2
	}

	score += completeness
	factors++

	// Check recency
	if !code.UpdatedAt.IsZero() {
		age := time.Since(code.UpdatedAt)
		if age < 30*24*time.Hour { // Less than 30 days
			score += 1.0
		} else if age < 90*24*time.Hour { // Less than 90 days
			score += 0.7
		} else {
			score += 0.3
		}
		factors++
	}

	if factors > 0 {
		return score / float64(factors)
	}

	return 0.5 // Default score
}

// calculateConfidence calculates confidence based on strategy and data quality
func (gds *GracefulDegradationService) calculateConfidence(strategy DegradationStrategy, dataQuality float64) float64 {
	baseConfidence := dataQuality

	// Adjust confidence based on strategy
	switch strategy {
	case StrategyCachedResults:
		return baseConfidence * 0.9 // High confidence for cached data
	case StrategyFallbackData:
		return baseConfidence * 0.8 // Good confidence for fallback data
	case StrategyPartialResults:
		return baseConfidence * 0.7 // Moderate confidence for partial results
	case StrategyAlternativeLogic:
		return baseConfidence * 0.6 // Lower confidence for alternative logic
	case StrategyStaticResponse:
		return baseConfidence * 0.4 // Low confidence for static response
	default:
		return baseConfidence * 0.3 // Very low confidence for unknown strategy
	}
}

// getDegradationLevel maps strategy to degradation level
func (gds *GracefulDegradationService) getDegradationLevel(strategy DegradationStrategy) DegradationLevel {
	switch strategy {
	case StrategyCachedResults:
		return LevelPartial
	case StrategyFallbackData:
		return LevelPartial
	case StrategyPartialResults:
		return LevelMinimal
	case StrategyAlternativeLogic:
		return LevelMinimal
	case StrategyStaticResponse:
		return LevelFallback
	default:
		return LevelCritical
	}
}

// generateWarnings generates warnings for degraded responses
func (gds *GracefulDegradationService) generateWarnings(strategy DegradationStrategy, attempt FallbackAttempt) []string {
	warnings := []string{}

	switch strategy {
	case StrategyCachedResults:
		warnings = append(warnings, "Using cached data - may not reflect recent changes")
	case StrategyFallbackData:
		warnings = append(warnings, "Using fallback data - accuracy may be reduced")
	case StrategyPartialResults:
		warnings = append(warnings, "Providing partial results - some information may be missing")
	case StrategyAlternativeLogic:
		warnings = append(warnings, "Using simplified classification logic - confidence reduced")
	case StrategyStaticResponse:
		warnings = append(warnings, "Using generic fallback response - manual review recommended")
	}

	if attempt.DataQuality < 0.7 {
		warnings = append(warnings, "Data quality below recommended threshold")
	}

	if attempt.Duration > gds.config.FallbackTimeout {
		warnings = append(warnings, "Fallback operation exceeded timeout threshold")
	}

	return warnings
}

// generateRecommendations generates recommendations based on degradation
func (gds *GracefulDegradationService) generateRecommendations(strategy DegradationStrategy, originalError error) []string {
	recommendations := []string{}

	// Common recommendations
	recommendations = append(recommendations, "Monitor system health and resolve underlying issues")
	recommendations = append(recommendations, "Review and validate the degraded response")

	// Strategy-specific recommendations
	switch strategy {
	case StrategyCachedResults:
		recommendations = append(recommendations, "Update cache when primary service is restored")
	case StrategyFallbackData:
		recommendations = append(recommendations, "Update fallback data sources regularly")
	case StrategyPartialResults:
		recommendations = append(recommendations, "Complete the classification when primary service is available")
	case StrategyAlternativeLogic:
		recommendations = append(recommendations, "Re-run classification with full logic when possible")
	case StrategyStaticResponse:
		recommendations = append(recommendations, "Manual classification review required")
	}

	// Error-specific recommendations
	if originalError != nil {
		errorMsg := strings.ToLower(originalError.Error())
		if strings.Contains(errorMsg, "timeout") {
			recommendations = append(recommendations, "Check network connectivity and service response times")
		}
		if strings.Contains(errorMsg, "connection") {
			recommendations = append(recommendations, "Verify database connectivity and service availability")
		}
		if strings.Contains(errorMsg, "authentication") {
			recommendations = append(recommendations, "Check authentication credentials and permissions")
		}
	}

	return recommendations
}

var lastSuccessfulData interface{}

// getLastSuccessfulData returns the last successful data
func (gds *GracefulDegradationService) getLastSuccessfulData() interface{} {
	return lastSuccessfulData
}

// SetLastSuccessfulData sets the last successful data for fallback purposes
func (gds *GracefulDegradationService) SetLastSuccessfulData(data interface{}) {
	lastSuccessfulData = data
}

// initializeStaticData initializes the static fallback data
func (fp *FallbackDataProvider) initializeStaticData() {
	// Common business types for fallback
	fp.staticData["restaurant"] = &IndustryCode{
		ID:          "static-001",
		Code:        "5812",
		Type:        CodeTypeMCC,
		Description: "Eating Places and Restaurants",
		Category:    "Food Service",
		Confidence:  0.7,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	fp.staticData["retail"] = &IndustryCode{
		ID:          "static-002",
		Code:        "5999",
		Type:        CodeTypeMCC,
		Description: "Miscellaneous and Specialty Retail Stores",
		Category:    "Retail",
		Confidence:  0.7,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	fp.staticData["consulting"] = &IndustryCode{
		ID:          "static-003",
		Code:        "8999",
		Type:        CodeTypeMCC,
		Description: "Professional Services",
		Category:    "Professional Services",
		Confidence:  0.7,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	fp.lastUpdated = time.Now()
}

// generateSimpleRules generates simple classification rules for alternative logic
func generateSimpleRules() []SimpleRule {
	return []SimpleRule{
		{
			Keywords:    []string{"food", "restaurant", "dining", "cafe", "bistro"},
			Code:        "5812",
			Type:        CodeTypeMCC,
			Confidence:  0.6,
			Description: "Food Service",
		},
		{
			Keywords:    []string{"retail", "store", "shop", "boutique", "market"},
			Code:        "5999",
			Type:        CodeTypeMCC,
			Confidence:  0.6,
			Description: "Retail",
		},
		{
			Keywords:    []string{"consulting", "advisory", "professional", "services"},
			Code:        "8999",
			Type:        CodeTypeMCC,
			Confidence:  0.6,
			Description: "Professional Services",
		},
		{
			Keywords:    []string{"technology", "software", "IT", "computer"},
			Code:        "7372",
			Type:        CodeTypeMCC,
			Confidence:  0.6,
			Description: "Technology Services",
		},
		{
			Keywords:    []string{"healthcare", "medical", "clinic", "hospital"},
			Code:        "8011",
			Type:        CodeTypeMCC,
			Confidence:  0.6,
			Description: "Healthcare",
		},
	}
}

// scoreWithSimpleRules performs simple rule-based scoring
func (as *AlternativeScorer) scoreWithSimpleRules(query string) *IndustryCode {
	query = strings.ToLower(strings.TrimSpace(query))

	for _, rule := range as.simpleRules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(query, strings.ToLower(keyword)) {
				return &IndustryCode{
					ID:          fmt.Sprintf("alt-%s-%s", rule.Type, rule.Code),
					Code:        rule.Code,
					Type:        rule.Type,
					Description: fmt.Sprintf("%s (Alternative Logic)", rule.Description),
					Category:    rule.Description,
					Confidence:  rule.Confidence,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
			}
		}
	}

	return nil
}

// CacheResult caches a successful result for future fallback use
func (gds *GracefulDegradationService) CacheResult(query string, code *IndustryCode, confidence float64) {
	if gds.config.EnableCachedResults && query != "" && code != nil {
		gds.fallbackData.cachedData[query] = &CachedResult{
			Code:       code,
			Confidence: confidence,
			Timestamp:  time.Now(),
			Source:     "primary_operation",
		}

		gds.logger.Info("cached result for fallback",
			zap.String("query", query),
			zap.Float64("confidence", confidence))
	}
}

// GetDegradationMetrics returns metrics about degradation usage
func (gds *GracefulDegradationService) GetDegradationMetrics() map[string]interface{} {
	return map[string]interface{}{
		"static_data_entries": len(gds.fallbackData.staticData),
		"cached_entries":      len(gds.fallbackData.cachedData),
		"cache_timeout":       gds.config.CacheTimeout.String(),
		"fallback_enabled":    gds.config.EnableFallback,
		"partial_enabled":     gds.config.EnablePartialResults,
		"cached_enabled":      gds.config.EnableCachedResults,
		"last_updated":        gds.fallbackData.lastUpdated,
	}
}

// CleanupExpiredCache removes expired cache entries
func (gds *GracefulDegradationService) CleanupExpiredCache() {
	now := time.Now()
	expired := []string{}

	for key, cached := range gds.fallbackData.cachedData {
		if now.Sub(cached.Timestamp) > gds.config.CacheTimeout {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(gds.fallbackData.cachedData, key)
	}

	if len(expired) > 0 {
		gds.logger.Info("cleaned up expired cache entries", zap.Int("count", len(expired)))
	}
}
