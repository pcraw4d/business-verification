package webanalysis

import (
	"fmt"
	"math"
	"time"
)

// ScrapingDepth represents the depth configuration for web scraping
type ScrapingDepth struct {
	MaxDepth          int       `json:"max_depth"`
	MaxPages          int       `json:"max_pages"`
	MaxLinksPerPage   int       `json:"max_links_per_page"`
	MaxContentLength  int       `json:"max_content_length"`
	FollowExternal    bool      `json:"follow_external"`
	FollowSubdomains  bool      `json:"follow_subdomains"`
	RespectRobotsTxt  bool      `json:"respect_robots_txt"`
	DelayBetweenPages float64   `json:"delay_between_pages"`
	PriorityScore     float64   `json:"priority_score"`
	DepthReason       string    `json:"depth_reason"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// DynamicScrapingDepthManager manages dynamic scraping depth calculations
type DynamicScrapingDepthManager struct {
	config           ScrapingDepthConfig
	qualityAssessor  *PageContentQualityAssessor
	pageTypeDetector *PageTypeDetector
	relevanceScorer  *PageRelevanceScorer
}

// ScrapingDepthConfig holds configuration for dynamic scraping depth
type ScrapingDepthConfig struct {
	BaseDepth           ScrapingDepth      `json:"base_depth"`
	QualityThresholds   map[string]float64 `json:"quality_thresholds"`
	RelevanceThresholds map[string]float64 `json:"relevance_thresholds"`
	PageTypeWeights     map[string]float64 `json:"page_type_weights"`
	DepthMultipliers    map[string]float64 `json:"depth_multipliers"`
	MaxDepthLimits      map[string]int     `json:"max_depth_limits"`
}

// NewDynamicScrapingDepthManager creates a new dynamic scraping depth manager
func NewDynamicScrapingDepthManager(config ScrapingDepthConfig, qualityAssessor *PageContentQualityAssessor, pageTypeDetector *PageTypeDetector, relevanceScorer *PageRelevanceScorer) *DynamicScrapingDepthManager {
	return &DynamicScrapingDepthManager{
		config:           config,
		qualityAssessor:  qualityAssessor,
		pageTypeDetector: pageTypeDetector,
		relevanceScorer:  relevanceScorer,
	}
}

// CalculateScrapingDepth calculates the optimal scraping depth for a page
func (dsdm *DynamicScrapingDepthManager) CalculateScrapingDepth(content *ScrapedContent, business string, context *ScoringContext) *ScrapingDepth {
	depth := &ScrapingDepth{
		CalculatedAt: time.Now(),
	}

	// Get page quality assessment
	contentQuality := dsdm.qualityAssessor.AssessContentQuality(content, business)

	// Get page type detection
	pageTypeConfig := PageTypeConfig{
		DetectionWeights: map[string]float64{
			"url":       0.3,
			"content":   0.4,
			"structure": 0.3,
		},
	}
	pageTypeDetector := NewPageTypeDetector(pageTypeConfig)
	pageType := pageTypeDetector.DetectPageType(content)

	// Get page relevance score
	relevanceScore := dsdm.relevanceScorer.ScorePage(content, business, context)

	// Calculate priority score based on multiple factors
	priorityScore := dsdm.calculatePriorityScore(contentQuality, pageType, relevanceScore)

	// Determine scraping depth based on priority score
	depth = dsdm.determineDepthFromPriority(priorityScore, contentQuality, pageType, relevanceScore)

	// Apply page type specific adjustments
	depth = dsdm.applyPageTypeAdjustments(depth, pageType)

	// Apply quality-based adjustments
	depth = dsdm.applyQualityAdjustments(depth, contentQuality)

	// Apply relevance-based adjustments
	depth = dsdm.applyRelevanceAdjustments(depth, relevanceScore)

	// Ensure depth is within reasonable limits
	depth = dsdm.enforceDepthLimits(depth)

	// Set delay based on depth and priority
	depth.DelayBetweenPages = dsdm.calculateDelay(depth, priorityScore)

	return depth
}

// calculatePriorityScore calculates overall priority score for scraping
func (dsdm *DynamicScrapingDepthManager) calculatePriorityScore(quality *PageContentQuality, pageType *PageTypeDetection, relevance *PageRelevanceScore) float64 {
	// Quality weight (30%)
	qualityWeight := 0.3
	qualityScore := quality.OverallQuality

	// Page type weight (25%)
	pageTypeWeight := 0.25
	pageTypeScore := dsdm.pageTypeDetector.GetPageTypePriority(pageType.Type)

	// Relevance weight (45%)
	relevanceWeight := 0.45
	relevanceScore := relevance.OverallScore

	// Calculate weighted priority score
	priorityScore := (qualityScore * qualityWeight) +
		(pageTypeScore * pageTypeWeight) +
		(relevanceScore * relevanceWeight)

	return math.Min(priorityScore, 1.0)
}

// determineDepthFromPriority determines base scraping depth from priority score
func (dsdm *DynamicScrapingDepthManager) determineDepthFromPriority(priorityScore float64, quality *PageContentQuality, pageType *PageTypeDetection, relevance *PageRelevanceScore) *ScrapingDepth {
	depth := &ScrapingDepth{
		CalculatedAt: time.Now(),
	}

	// Base depth calculation
	if priorityScore >= 0.8 {
		// High priority - deep scraping
		depth.MaxDepth = 5
		depth.MaxPages = 50
		depth.MaxLinksPerPage = 20
		depth.MaxContentLength = 10000
		depth.FollowExternal = true
		depth.FollowSubdomains = true
		depth.RespectRobotsTxt = true
		depth.PriorityScore = priorityScore
		depth.DepthReason = "High priority page with excellent quality and relevance"
	} else if priorityScore >= 0.6 {
		// Medium-high priority - moderate scraping
		depth.MaxDepth = 3
		depth.MaxPages = 25
		depth.MaxLinksPerPage = 15
		depth.MaxContentLength = 5000
		depth.FollowExternal = false
		depth.FollowSubdomains = true
		depth.RespectRobotsTxt = true
		depth.PriorityScore = priorityScore
		depth.DepthReason = "Medium-high priority page with good quality and relevance"
	} else if priorityScore >= 0.4 {
		// Medium priority - light scraping
		depth.MaxDepth = 2
		depth.MaxPages = 15
		depth.MaxLinksPerPage = 10
		depth.MaxContentLength = 3000
		depth.FollowExternal = false
		depth.FollowSubdomains = false
		depth.RespectRobotsTxt = true
		depth.PriorityScore = priorityScore
		depth.DepthReason = "Medium priority page with moderate quality and relevance"
	} else {
		// Low priority - minimal scraping
		depth.MaxDepth = 1
		depth.MaxPages = 5
		depth.MaxLinksPerPage = 5
		depth.MaxContentLength = 1000
		depth.FollowExternal = false
		depth.FollowSubdomains = false
		depth.RespectRobotsTxt = true
		depth.PriorityScore = priorityScore
		depth.DepthReason = "Low priority page with limited quality and relevance"
	}

	return depth
}

// applyPageTypeAdjustments applies page type specific adjustments
func (dsdm *DynamicScrapingDepthManager) applyPageTypeAdjustments(depth *ScrapingDepth, pageType *PageTypeDetection) *ScrapingDepth {
	// Get page type weight multiplier
	weightMultiplier := dsdm.config.PageTypeWeights[pageType.Type]
	if weightMultiplier == 0 {
		weightMultiplier = 1.0
	}

	// Get depth multiplier for this page type
	depthMultiplier := dsdm.config.DepthMultipliers[pageType.Type]
	if depthMultiplier == 0 {
		depthMultiplier = 1.0
	}

	// Apply adjustments based on page type
	switch pageType.Type {
	case "about_us":
		// About us pages are very important - increase depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.2)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.3)
		depth.MaxContentLength = int(float64(depth.MaxContentLength) * 1.2)
		depth.DepthReason += " (About Us page - increased depth for comprehensive business information)"

	case "mission":
		// Mission pages are important - moderate increase
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.1)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.2)
		depth.DepthReason += " (Mission page - increased depth for business purpose analysis)"

	case "services":
		// Services pages are very important - increase depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.3)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.4)
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 1.2)
		depth.DepthReason += " (Services page - increased depth for comprehensive service analysis)"

	case "products":
		// Products pages are very important - increase depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.3)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.4)
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 1.2)
		depth.DepthReason += " (Products page - increased depth for comprehensive product analysis)"

	case "contact":
		// Contact pages are moderately important
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.0)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.1)
		depth.DepthReason += " (Contact page - standard depth for contact information)"

	case "team":
		// Team pages are moderately important
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.0)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.1)
		depth.DepthReason += " (Team page - standard depth for team information)"

	case "careers":
		// Careers pages are less important - reduce depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 0.8)
		depth.MaxPages = int(float64(depth.MaxPages) * 0.7)
		depth.DepthReason += " (Careers page - reduced depth for employment information)"

	case "news":
		// News pages are less important - reduce depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 0.7)
		depth.MaxPages = int(float64(depth.MaxPages) * 0.6)
		depth.DepthReason += " (News page - reduced depth for news content)"

	case "unknown":
		// Unknown pages get minimal depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 0.5)
		depth.MaxPages = int(float64(depth.MaxPages) * 0.5)
		depth.DepthReason += " (Unknown page type - minimal depth)"
	}

	return depth
}

// applyQualityAdjustments applies quality-based adjustments
func (dsdm *DynamicScrapingDepthManager) applyQualityAdjustments(depth *ScrapingDepth, quality *PageContentQuality) *ScrapingDepth {
	// Adjust based on overall quality
	if quality.OverallQuality >= 0.8 {
		// High quality content - increase depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.2)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.3)
		depth.MaxContentLength = int(float64(depth.MaxContentLength) * 1.2)
		depth.DepthReason += " (High quality content - increased depth)"
	} else if quality.OverallQuality >= 0.6 {
		// Good quality content - moderate increase
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.1)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.1)
		depth.DepthReason += " (Good quality content - moderate increase)"
	} else if quality.OverallQuality < 0.3 {
		// Low quality content - reduce depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 0.7)
		depth.MaxPages = int(float64(depth.MaxPages) * 0.7)
		depth.DepthReason += " (Low quality content - reduced depth)"
	}

	// Adjust based on content completeness
	if quality.CompletenessMetrics.CompletenessScore >= 0.8 {
		depth.MaxContentLength = int(float64(depth.MaxContentLength) * 1.1)
		depth.DepthReason += " (Complete content - increased content length)"
	} else if quality.CompletenessMetrics.CompletenessScore < 0.4 {
		depth.MaxContentLength = int(float64(depth.MaxContentLength) * 0.8)
		depth.DepthReason += " (Incomplete content - reduced content length)"
	}

	// Adjust based on business content quality
	if quality.BusinessMetrics.BusinessScore >= 0.8 {
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 1.2)
		depth.DepthReason += " (High business content - increased link exploration)"
	} else if quality.BusinessMetrics.BusinessScore < 0.3 {
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 0.8)
		depth.DepthReason += " (Low business content - reduced link exploration)"
	}

	return depth
}

// applyRelevanceAdjustments applies relevance-based adjustments
func (dsdm *DynamicScrapingDepthManager) applyRelevanceAdjustments(depth *ScrapingDepth, relevance *PageRelevanceScore) *ScrapingDepth {
	// Adjust based on overall relevance
	if relevance.OverallScore >= 0.8 {
		// High relevance - increase depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.3)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.4)
		depth.FollowExternal = true
		depth.DepthReason += " (High relevance - increased depth and external links)"
	} else if relevance.OverallScore >= 0.6 {
		// Good relevance - moderate increase
		depth.MaxDepth = int(float64(depth.MaxDepth) * 1.1)
		depth.MaxPages = int(float64(depth.MaxPages) * 1.2)
		depth.DepthReason += " (Good relevance - moderate increase)"
	} else if relevance.OverallScore < 0.3 {
		// Low relevance - reduce depth
		depth.MaxDepth = int(float64(depth.MaxDepth) * 0.6)
		depth.MaxPages = int(float64(depth.MaxPages) * 0.6)
		depth.FollowExternal = false
		depth.DepthReason += " (Low relevance - reduced depth and no external links)"
	}

	// Adjust based on business relevance
	if relevance.BusinessRelevance.BusinessNameMatch >= 0.8 {
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 1.3)
		depth.DepthReason += " (High business name match - increased link exploration)"
	} else if relevance.BusinessRelevance.BusinessNameMatch < 0.3 {
		depth.MaxLinksPerPage = int(float64(depth.MaxLinksPerPage) * 0.7)
		depth.DepthReason += " (Low business name match - reduced link exploration)"
	}

	// Adjust based on industry relevance
	if relevance.BusinessRelevance.IndustryRelevance >= 0.8 {
		depth.MaxContentLength = int(float64(depth.MaxContentLength) * 1.2)
		depth.DepthReason += " (High industry relevance - increased content length)"
	}

	return depth
}

// enforceDepthLimits ensures depth is within reasonable limits
func (dsdm *DynamicScrapingDepthManager) enforceDepthLimits(depth *ScrapingDepth) *ScrapingDepth {
	// Enforce maximum depth limits
	if depth.MaxDepth > 10 {
		depth.MaxDepth = 10
		depth.DepthReason += " (Capped at maximum depth of 10)"
	}
	if depth.MaxDepth < 1 {
		depth.MaxDepth = 1
		depth.DepthReason += " (Minimum depth of 1 enforced)"
	}

	// Enforce maximum pages limits
	if depth.MaxPages > 100 {
		depth.MaxPages = 100
		depth.DepthReason += " (Capped at maximum pages of 100)"
	}
	if depth.MaxPages < 1 {
		depth.MaxPages = 1
		depth.DepthReason += " (Minimum pages of 1 enforced)"
	}

	// Enforce maximum links per page limits
	if depth.MaxLinksPerPage > 50 {
		depth.MaxLinksPerPage = 50
		depth.DepthReason += " (Capped at maximum links per page of 50)"
	}
	if depth.MaxLinksPerPage < 1 {
		depth.MaxLinksPerPage = 1
		depth.DepthReason += " (Minimum links per page of 1 enforced)"
	}

	// Enforce maximum content length limits
	if depth.MaxContentLength > 50000 {
		depth.MaxContentLength = 50000
		depth.DepthReason += " (Capped at maximum content length of 50KB)"
	}
	if depth.MaxContentLength < 100 {
		depth.MaxContentLength = 100
		depth.DepthReason += " (Minimum content length of 100 chars enforced)"
	}

	return depth
}

// calculateDelay calculates appropriate delay between page requests
func (dsdm *DynamicScrapingDepthManager) calculateDelay(depth *ScrapingDepth, priorityScore float64) float64 {
	// Base delay in seconds
	baseDelay := 1.0

	// Adjust delay based on priority score
	if priorityScore >= 0.8 {
		// High priority - faster scraping
		baseDelay = 0.5
	} else if priorityScore >= 0.6 {
		// Medium-high priority - moderate speed
		baseDelay = 1.0
	} else if priorityScore >= 0.4 {
		// Medium priority - slower scraping
		baseDelay = 2.0
	} else {
		// Low priority - very slow scraping
		baseDelay = 3.0
	}

	// Adjust delay based on depth
	if depth.MaxDepth > 5 {
		baseDelay *= 0.8 // Faster for deep scraping
	} else if depth.MaxDepth < 2 {
		baseDelay *= 1.5 // Slower for shallow scraping
	}

	// Ensure minimum delay
	if baseDelay < 0.1 {
		baseDelay = 0.1
	}

	return baseDelay
}

// GetScrapingStrategy returns a human-readable scraping strategy description
func (dsdm *DynamicScrapingDepthManager) GetScrapingStrategy(depth *ScrapingDepth) string {
	strategy := "Dynamic scraping strategy: "

	if depth.MaxDepth >= 5 {
		strategy += "Deep crawling with "
	} else if depth.MaxDepth >= 3 {
		strategy += "Moderate crawling with "
	} else if depth.MaxDepth >= 2 {
		strategy += "Light crawling with "
	} else {
		strategy += "Minimal crawling with "
	}

	strategy += fmt.Sprintf("%d pages, %d links per page, %d chars max content",
		depth.MaxPages, depth.MaxLinksPerPage, depth.MaxContentLength)

	if depth.FollowExternal {
		strategy += ", following external links"
	}
	if depth.FollowSubdomains {
		strategy += ", following subdomains"
	}
	if depth.RespectRobotsTxt {
		strategy += ", respecting robots.txt"
	}

	strategy += fmt.Sprintf(", %.1fs delay between requests", depth.DelayBetweenPages)

	return strategy
}

// IsHighPriorityDepth checks if the scraping depth is for a high priority page
func (dsdm *DynamicScrapingDepthManager) IsHighPriorityDepth(depth *ScrapingDepth) bool {
	return depth.PriorityScore >= 0.7
}

// IsDeepScraping checks if this is a deep scraping configuration
func (dsdm *DynamicScrapingDepthManager) IsDeepScraping(depth *ScrapingDepth) bool {
	return depth.MaxDepth >= 4 && depth.MaxPages >= 20
}

// IsMinimalScraping checks if this is a minimal scraping configuration
func (dsdm *DynamicScrapingDepthManager) IsMinimalScraping(depth *ScrapingDepth) bool {
	return depth.MaxDepth <= 1 && depth.MaxPages <= 5
}
