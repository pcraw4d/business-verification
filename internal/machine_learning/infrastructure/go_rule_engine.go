package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// GoRuleEngine represents the Go rule engine for rule-based systems
type GoRuleEngine struct {
	// Service configuration
	endpoint string
	config   GoRuleEngineConfig

	// Rule systems
	keywordMatcher   *KeywordMatcher
	mccCodeLookup    *MCCCodeLookup
	blacklistChecker *BlacklistChecker

	// Caching
	cache *RuleEngineCache

	// Performance tracking
	metrics *ServiceMetrics

	// Health status
	healthStatus *HealthStatus

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// GoRuleEngineConfig holds configuration for the Go rule engine
type GoRuleEngineConfig struct {
	// Service configuration
	Host string `json:"host"`
	Port int    `json:"port"`

	// Rule system configuration
	KeywordMatchingEnabled bool `json:"keyword_matching_enabled"`
	MCCCodeLookupEnabled   bool `json:"mcc_code_lookup_enabled"`
	BlacklistCheckEnabled  bool `json:"blacklist_check_enabled"`

	// Performance configuration
	MaxConcurrentRules int           `json:"max_concurrent_rules"`
	RuleTimeout        time.Duration `json:"rule_timeout"`
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheSize          int           `json:"cache_size"`
	CacheTTL           time.Duration `json:"cache_ttl"`

	// Rule data sources
	KeywordDatabasePath   string `json:"keyword_database_path"`
	MCCCodeDatabasePath   string `json:"mcc_code_database_path"`
	BlacklistDatabasePath string `json:"blacklist_database_path"`

	// Performance targets
	TargetResponseTime time.Duration `json:"target_response_time"` // <10ms
	TargetAccuracy     float64       `json:"target_accuracy"`      // >90%
}

// Types KeywordMatcher, MCCCodeLookup, BlacklistChecker, and RuleEngineCache
// are defined in their respective files (keyword_matcher.go, mcc_code_lookup.go,
// blacklist_checker.go, rule_engine_cache.go) to avoid redeclaration

// Types MCCCodeInfo, BlacklistEntry, CachedClassificationResult, CachedRiskResult,
// CacheConfig, RuleEngineClassificationRequest, RuleEngineRiskRequest,
// RuleEngineClassificationResponse, and RuleEngineRiskResponse are defined in types.go
// to avoid redeclaration

// NewGoRuleEngine creates a new Go rule engine
func NewGoRuleEngine(endpoint string, logger *log.Logger) *GoRuleEngine {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &GoRuleEngine{
		endpoint: endpoint,
		config: GoRuleEngineConfig{
			KeywordMatchingEnabled: true,
			MCCCodeLookupEnabled:   true,
			BlacklistCheckEnabled:  true,
			MaxConcurrentRules:     100,
			RuleTimeout:            5 * time.Second,
			CacheEnabled:           true,
			CacheSize:              1000,
			CacheTTL:               1 * time.Hour,
			TargetResponseTime:     10 * time.Millisecond,
			TargetAccuracy:         0.90,
		},
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the Go rule engine
func (gre *GoRuleEngine) Initialize(ctx context.Context) error {
	gre.mu.Lock()
	defer gre.mu.Unlock()

	gre.logger.Printf("🔧 Initializing Go Rule Engine")

	// Initialize metrics
	gre.metrics = &ServiceMetrics{
		RequestCount:   0,
		SuccessCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		P95Latency:     0,
		P99Latency:     0,
		Throughput:     0,
		ErrorRate:      0,
		LastUpdated:    time.Now(),
	}

	// Initialize health status
	gre.healthStatus = &HealthStatus{
		Status:    "unknown",
		LastCheck: time.Now(),
		Checks:    make(map[string]HealthCheck),
	}

	// Initialize keyword matcher
	if gre.config.KeywordMatchingEnabled {
		gre.logger.Printf("🔍 Initializing Keyword Matcher")
		gre.keywordMatcher = NewKeywordMatcher(gre.logger)
		if err := gre.keywordMatcher.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize keyword matcher: %w", err)
		}
	}

	// Initialize MCC code lookup
	if gre.config.MCCCodeLookupEnabled {
		gre.logger.Printf("📋 Initializing MCC Code Lookup")
		gre.mccCodeLookup = NewMCCCodeLookup(gre.logger)
		if err := gre.mccCodeLookup.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize MCC code lookup: %w", err)
		}
	}

	// Initialize blacklist checker
	if gre.config.BlacklistCheckEnabled {
		gre.logger.Printf("🚫 Initializing Blacklist Checker")
		gre.blacklistChecker = NewBlacklistChecker(gre.logger)
		if err := gre.blacklistChecker.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize blacklist checker: %w", err)
		}
	}

	// Initialize cache
	if gre.config.CacheEnabled {
		gre.logger.Printf("💾 Initializing Rule Engine Cache")
		gre.cache = NewRuleEngineCache(gre.logger)
		if err := gre.cache.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
	}

	gre.logger.Printf("✅ Go Rule Engine initialized successfully")
	return nil
}

// Start starts the Go rule engine
func (gre *GoRuleEngine) Start(ctx context.Context) error {
	gre.mu.Lock()
	defer gre.mu.Unlock()

	gre.logger.Printf("🚀 Starting Go Rule Engine")

	// Start cache cleanup if enabled
	if gre.cache != nil {
		go gre.cache.StartCleanup(ctx)
	}

	gre.logger.Printf("✅ Go Rule Engine started successfully")
	return nil
}

// Stop stops the Go rule engine
func (gre *GoRuleEngine) Stop() {
	gre.mu.Lock()
	defer gre.mu.Unlock()

	gre.logger.Printf("🛑 Stopping Go Rule Engine")

	// Cancel context
	gre.cancel()

	gre.logger.Printf("✅ Go Rule Engine stopped successfully")
}

// Classify performs business classification using rule-based systems
func (gre *GoRuleEngine) Classify(ctx context.Context, req *RuleEngineClassificationRequest) (*RuleEngineClassificationResponse, error) {
	start := time.Now()

	gre.mu.Lock()
	gre.metrics.RequestCount++
	gre.mu.Unlock()

	// Check cache first
	if gre.cache != nil {
		cacheKey := gre.generateCacheKey("classification", req.BusinessName, req.Description, req.WebsiteURL)
		if cached, found := gre.cache.GetClassification(cacheKey); found {
			gre.mu.Lock()
			gre.metrics.SuccessCount++
			gre.mu.Unlock()
			return cached.Result, nil
		}
	}

	// Perform classification using available rule systems
	var classifications []ClassificationPrediction
	var confidence float64
	var method string

	// Try keyword matching first
	if gre.keywordMatcher != nil {
		keywordResults, err := gre.keywordMatcher.ClassifyByKeywords(ctx, req.BusinessName, req.Description)
		if err == nil && len(keywordResults) > 0 {
			classifications = keywordResults
			confidence = 0.85 // High confidence for keyword matching
			method = "keyword_matching"
		}
	}

	// If no keyword results, try MCC code lookup
	if len(classifications) == 0 && gre.mccCodeLookup != nil {
		mccResults, err := gre.mccCodeLookup.ClassifyByMCC(ctx, req.BusinessName, req.Description)
		if err == nil && len(mccResults) > 0 {
			classifications = mccResults
			confidence = 0.80 // Good confidence for MCC lookup
			method = "mcc_lookup"
		}
	}

	// Create response
	response := &RuleEngineClassificationResponse{
		RequestID:       fmt.Sprintf("rule_%d", time.Now().UnixNano()),
		Classifications: classifications,
		Confidence:      confidence,
		ProcessingTime:  time.Since(start),
		Timestamp:       time.Now(),
		Success:         len(classifications) > 0,
		Method:          method,
	}

	if len(classifications) == 0 {
		response.Error = "no classification results found"
		gre.mu.Lock()
		gre.metrics.ErrorCount++
		gre.mu.Unlock()
	} else {
		gre.mu.Lock()
		gre.metrics.SuccessCount++
		gre.updateLatencyMetrics(time.Since(start))
		gre.mu.Unlock()

		// Cache the result
		if gre.cache != nil {
			cacheKey := gre.generateCacheKey("classification", req.BusinessName, req.Description, req.WebsiteURL)
			gre.cache.SetClassification(cacheKey, response, gre.config.CacheTTL)
		}
	}

	return response, nil
}

// DetectRisk performs risk detection using rule-based systems
func (gre *GoRuleEngine) DetectRisk(ctx context.Context, req *RuleEngineRiskRequest) (*RuleEngineRiskResponse, error) {
	start := time.Now()

	gre.mu.Lock()
	gre.metrics.RequestCount++
	gre.mu.Unlock()

	// Check cache first
	if gre.cache != nil {
		cacheKey := gre.generateCacheKey("risk", req.BusinessName, req.Description, req.WebsiteURL)
		if cached, found := gre.cache.GetRisk(cacheKey); found {
			gre.mu.Lock()
			gre.metrics.SuccessCount++
			gre.mu.Unlock()
			return cached.Result, nil
		}
	}

	var detectedRisks []DetectedRisk
	var riskScore float64
	var riskLevel string
	var method string

	// Check blacklist first (fastest and most reliable)
	if gre.blacklistChecker != nil {
		blacklistRisks, err := gre.blacklistChecker.CheckBlacklist(ctx, req.BusinessName, req.WebsiteURL)
		if err == nil && len(blacklistRisks) > 0 {
			detectedRisks = append(detectedRisks, blacklistRisks...)
			riskScore = 1.0 // Maximum risk for blacklisted entities
			riskLevel = "critical"
			method = "blacklist_check"
		}
	}

	// If not blacklisted, check risk keywords
	if len(detectedRisks) == 0 && gre.keywordMatcher != nil {
		keywordRisks, err := gre.keywordMatcher.DetectRiskKeywords(ctx, req.BusinessName, req.Description, req.WebsiteContent)
		if err == nil && len(keywordRisks) > 0 {
			detectedRisks = append(detectedRisks, keywordRisks...)
			riskScore = gre.calculateRiskScore(keywordRisks)
			riskLevel = gre.determineRiskLevel(riskScore)
			method = "keyword_matching"
		}
	}

	// Check MCC code restrictions
	if gre.mccCodeLookup != nil {
		mccRisks, err := gre.mccCodeLookup.CheckMCCRestrictions(ctx, req.BusinessName, req.Description)
		if err == nil && len(mccRisks) > 0 {
			detectedRisks = append(detectedRisks, mccRisks...)
			if riskScore < 0.8 { // Don't override high risk scores
				riskScore = gre.calculateRiskScore(mccRisks)
				riskLevel = gre.determineRiskLevel(riskScore)
			}
			if method == "" {
				method = "mcc_lookup"
			} else {
				method += ",mcc_lookup"
			}
		}
	}

	// Create response
	response := &RuleEngineRiskResponse{
		RequestID:      fmt.Sprintf("risk_%d", time.Now().UnixNano()),
		RiskScore:      riskScore,
		RiskLevel:      riskLevel,
		DetectedRisks:  detectedRisks,
		ProcessingTime: time.Since(start),
		Timestamp:      time.Now(),
		Success:        true,
		Method:         method,
	}

	gre.mu.Lock()
	gre.metrics.SuccessCount++
	gre.updateLatencyMetrics(time.Since(start))
	gre.mu.Unlock()

	// Cache the result
	if gre.cache != nil {
		cacheKey := gre.generateCacheKey("risk", req.BusinessName, req.Description, req.WebsiteURL)
		gre.cache.SetRisk(cacheKey, response, gre.config.CacheTTL)
	}

	return response, nil
}

// HealthCheck performs a health check on the Go rule engine
func (gre *GoRuleEngine) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	start := time.Now()

	// Check keyword matcher
	keywordMatcherHealthy := true
	if gre.keywordMatcher != nil {
		if err := gre.keywordMatcher.HealthCheck(ctx); err != nil {
			keywordMatcherHealthy = false
		}
	}

	// Check MCC code lookup
	mccLookupHealthy := true
	if gre.mccCodeLookup != nil {
		if err := gre.mccCodeLookup.HealthCheck(ctx); err != nil {
			mccLookupHealthy = false
		}
	}

	// Check blacklist checker
	blacklistCheckerHealthy := true
	if gre.blacklistChecker != nil {
		if err := gre.blacklistChecker.HealthCheck(ctx); err != nil {
			blacklistCheckerHealthy = false
		}
	}

	// Check cache
	cacheHealthy := true
	if gre.cache != nil {
		if err := gre.cache.HealthCheck(ctx); err != nil {
			cacheHealthy = false
		}
	}

	// Determine overall health
	status := "pass"
	message := "All components healthy"
	if !keywordMatcherHealthy || !mccLookupHealthy || !blacklistCheckerHealthy || !cacheHealthy {
		status = "fail"
		message = "One or more components unhealthy"
	}

	return &HealthCheck{
		Name:      "go_rule_engine",
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}, nil
}

// GetMetrics returns service metrics
func (gre *GoRuleEngine) GetMetrics(ctx context.Context) (*ServiceMetrics, error) {
	gre.mu.RLock()
	defer gre.mu.RUnlock()

	// Return a copy of metrics
	metrics := *gre.metrics
	return &metrics, nil
}

// generateCacheKey generates a cache key for the given parameters
func (gre *GoRuleEngine) generateCacheKey(operation, businessName, description, websiteURL string) string {
	// Create a simple hash-based cache key
	key := fmt.Sprintf("%s:%s:%s:%s", operation, businessName, description, websiteURL)
	return fmt.Sprintf("%x", []byte(key))
}

// calculateRiskScore calculates the overall risk score from detected risks
func (gre *GoRuleEngine) calculateRiskScore(risks []DetectedRisk) float64 {
	if len(risks) == 0 {
		return 0.0
	}

	var totalScore float64
	for _, risk := range risks {
		switch risk.Severity {
		case "low":
			totalScore += 0.2
		case "medium":
			totalScore += 0.5
		case "high":
			totalScore += 0.8
		case "critical":
			totalScore += 1.0
		}
	}

	// Average the scores and cap at 1.0
	avgScore := totalScore / float64(len(risks))
	if avgScore > 1.0 {
		avgScore = 1.0
	}

	return avgScore
}

// determineRiskLevel determines the risk level based on the risk score
func (gre *GoRuleEngine) determineRiskLevel(riskScore float64) string {
	switch {
	case riskScore >= 0.8:
		return "critical"
	case riskScore >= 0.6:
		return "high"
	case riskScore >= 0.3:
		return "medium"
	default:
		return "low"
	}
}

// updateLatencyMetrics updates latency metrics
func (gre *GoRuleEngine) updateLatencyMetrics(latency time.Duration) {
	// Simple moving average for average latency
	if gre.metrics.AverageLatency == 0 {
		gre.metrics.AverageLatency = latency
	} else {
		gre.metrics.AverageLatency = (gre.metrics.AverageLatency + latency) / 2
	}

	// Update P95 and P99 (simplified implementation)
	if latency > gre.metrics.P95Latency {
		gre.metrics.P95Latency = latency
	}
	if latency > gre.metrics.P99Latency {
		gre.metrics.P99Latency = latency
	}

	gre.metrics.LastUpdated = time.Now()
}
