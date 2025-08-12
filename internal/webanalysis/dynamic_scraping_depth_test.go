package webanalysis

import (
	"testing"
	"time"
)

func TestDynamicScrapingDepthManager(t *testing.T) {
	// Create test configuration
	config := ScrapingDepthConfig{
		BaseDepth: ScrapingDepth{
			MaxDepth:          3,
			MaxPages:          20,
			MaxLinksPerPage:   10,
			MaxContentLength:  5000,
			FollowExternal:    false,
			FollowSubdomains:  true,
			RespectRobotsTxt:  true,
			DelayBetweenPages: 1.0,
		},
		QualityThresholds: map[string]float64{
			"high":   0.8,
			"medium": 0.6,
			"low":    0.3,
		},
		RelevanceThresholds: map[string]float64{
			"high":   0.8,
			"medium": 0.6,
			"low":    0.3,
		},
		PageTypeWeights: map[string]float64{
			"about_us": 1.2,
			"services": 1.3,
			"products": 1.3,
			"contact":  1.0,
			"team":     1.0,
			"careers":  0.8,
			"news":     0.7,
			"unknown":  0.5,
		},
		DepthMultipliers: map[string]float64{
			"about_us": 1.2,
			"services": 1.3,
			"products": 1.3,
			"contact":  1.0,
			"team":     1.0,
			"careers":  0.8,
			"news":     0.7,
			"unknown":  0.5,
		},
	}

	// Create test components
	qualityConfig := ContentQualityConfig{
		Weights: map[string]float64{
			"readability":       0.2,
			"structure":         0.2,
			"completeness":      0.2,
			"business_content":  0.2,
			"technical_content": 0.2,
		},
	}
	qualityAssessor := NewPageContentQualityAssessor(qualityConfig)

	pageTypeConfig := PageTypeConfig{
		DetectionWeights: map[string]float64{
			"url":       0.3,
			"content":   0.4,
			"structure": 0.3,
		},
	}
	pageTypeDetector := NewPageTypeDetector(pageTypeConfig)

	relevanceScorer := NewPageRelevanceScorer()

	// Create manager
	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test high priority content
	highPriorityContent := &ScrapedContent{
		URL:   "https://example.com/about-us",
		Title: "About Our Company",
		Text:  "We are a professional company established in 2010, specializing in quality services. Our team of experts provides comprehensive solutions to meet your needs. Contact us at info@example.com or call (555) 123-4567. We are certified and licensed professionals with years of experience in the industry.",
		HTML:  "<html><head><title>About Our Company</title></head><body><h1>About Us</h1><p>We are a professional company established in 2010.</p><h2>Our Services</h2><ul><li>Consulting</li><li>Training</li></ul></body></html>",
	}

	context := &ScoringContext{
		Industry:     "Technology",
		Location:     "United States",
		BusinessType: "Corporation",
		Metadata:     map[string]string{},
	}

	depth := manager.CalculateScrapingDepth(highPriorityContent, "Example Company", context)

	// Test high priority depth results
	if depth.MaxDepth < 2 {
		t.Errorf("Expected MaxDepth >= 2 for high priority content, got %d", depth.MaxDepth)
	}

	if depth.MaxPages < 15 {
		t.Errorf("Expected MaxPages >= 15 for high priority content, got %d", depth.MaxPages)
	}

	if depth.PriorityScore < 0.5 {
		t.Errorf("Expected PriorityScore >= 0.5 for high priority content, got %f", depth.PriorityScore)
	}

	if depth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated")
	}

	if depth.DelayBetweenPages <= 0 {
		t.Errorf("Expected DelayBetweenPages > 0, got %f", depth.DelayBetweenPages)
	}

	// Test low priority content
	lowPriorityContent := &ScrapedContent{
		URL:   "https://example.com/random-page",
		Title: "Random Page",
		Text:  "This is a random page with minimal content.",
		HTML:  "<html><body><p>Random content.</p></body></html>",
	}

	depth = manager.CalculateScrapingDepth(lowPriorityContent, "Example Company", context)

	// Test low priority depth results
	if depth.MaxDepth > 2 {
		t.Errorf("Expected MaxDepth <= 2 for low priority content, got %d", depth.MaxDepth)
	}

	if depth.MaxPages > 15 {
		t.Errorf("Expected MaxPages <= 15 for low priority content, got %d", depth.MaxPages)
	}

	if depth.FollowExternal {
		t.Error("Expected FollowExternal to be false for low priority content")
	}
}

func TestPriorityScoreCalculation(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test with high quality, high relevance, high priority page type
	highQuality := &PageContentQuality{
		OverallQuality: 0.9,
		AssessedAt:     time.Now(),
	}

	highRelevance := &PageRelevanceScore{
		OverallScore: 0.9,
		ScoredAt:     time.Now(),
	}

	highPriorityPageType := &PageTypeDetection{
		Type:       "about_us",
		Confidence: 0.9,
		DetectedAt: time.Now(),
	}

	priorityScore := manager.calculatePriorityScore(highQuality, highPriorityPageType, highRelevance)

	if priorityScore < 0.8 {
		t.Errorf("Expected high priority score >= 0.8, got %f", priorityScore)
	}

	// Test with low quality, low relevance, low priority page type
	lowQuality := &PageContentQuality{
		OverallQuality: 0.2,
		AssessedAt:     time.Now(),
	}

	lowRelevance := &PageRelevanceScore{
		OverallScore: 0.2,
		ScoredAt:     time.Now(),
	}

	lowPriorityPageType := &PageTypeDetection{
		Type:       "news",
		Confidence: 0.5,
		DetectedAt: time.Now(),
	}

	priorityScore = manager.calculatePriorityScore(lowQuality, lowPriorityPageType, lowRelevance)

	if priorityScore > 0.4 {
		t.Errorf("Expected low priority score <= 0.4, got %f", priorityScore)
	}
}

func TestDepthFromPriority(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test high priority depth
	highQuality := &PageContentQuality{OverallQuality: 0.9, AssessedAt: time.Now()}
	highRelevance := &PageRelevanceScore{OverallScore: 0.9, ScoredAt: time.Now()}
	highPriorityPageType := &PageTypeDetection{Type: "about_us", Confidence: 0.9, DetectedAt: time.Now()}

	depth := manager.determineDepthFromPriority(0.9, highQuality, highPriorityPageType, highRelevance)

	if depth.MaxDepth < 4 {
		t.Errorf("Expected high priority MaxDepth >= 4, got %d", depth.MaxDepth)
	}

	if depth.MaxPages < 30 {
		t.Errorf("Expected high priority MaxPages >= 30, got %d", depth.MaxPages)
	}

	if !depth.FollowExternal {
		t.Error("Expected high priority to follow external links")
	}

	// Test low priority depth
	lowQuality := &PageContentQuality{OverallQuality: 0.2, AssessedAt: time.Now()}
	lowRelevance := &PageRelevanceScore{OverallScore: 0.2, ScoredAt: time.Now()}
	lowPriorityPageType := &PageTypeDetection{Type: "news", Confidence: 0.5, DetectedAt: time.Now()}

	depth = manager.determineDepthFromPriority(0.2, lowQuality, lowPriorityPageType, lowRelevance)

	if depth.MaxDepth > 2 {
		t.Errorf("Expected low priority MaxDepth <= 2, got %d", depth.MaxDepth)
	}

	if depth.MaxPages > 10 {
		t.Errorf("Expected low priority MaxPages <= 10, got %d", depth.MaxPages)
	}

	if depth.FollowExternal {
		t.Error("Expected low priority to not follow external links")
	}
}

func TestPageTypeAdjustments(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test about_us page type adjustments
	baseDepth := &ScrapingDepth{
		MaxDepth:         3,
		MaxPages:         20,
		MaxLinksPerPage:  10,
		MaxContentLength: 5000,
		CalculatedAt:     time.Now(),
	}

	aboutUsPageType := &PageTypeDetection{
		Type:       "about_us",
		Confidence: 0.9,
		DetectedAt: time.Now(),
	}

	adjustedDepth := manager.applyPageTypeAdjustments(baseDepth, aboutUsPageType)

	// Check that adjustments were applied (may be limited by enforceDepthLimits)
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for about_us adjustments")
	}

	// Check that the reason mentions about_us
	if !contains(adjustedDepth.DepthReason, "About Us") {
		t.Error("Expected DepthReason to mention About Us page adjustments")
	}

	// Test news page type adjustments (should reduce depth)
	newsPageType := &PageTypeDetection{
		Type:       "news",
		Confidence: 0.8,
		DetectedAt: time.Now(),
	}

	adjustedDepth = manager.applyPageTypeAdjustments(baseDepth, newsPageType)

	// Check that adjustments were applied
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for news adjustments")
	}

	// Check that the reason mentions news
	if !contains(adjustedDepth.DepthReason, "News page") {
		t.Error("Expected DepthReason to mention News page adjustments")
	}
}

func TestQualityAdjustments(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	baseDepth := &ScrapingDepth{
		MaxDepth:         3,
		MaxPages:         20,
		MaxLinksPerPage:  10,
		MaxContentLength: 5000,
		CalculatedAt:     time.Now(),
	}

	// Test high quality adjustments
	highQuality := &PageContentQuality{
		OverallQuality: 0.9,
		CompletenessMetrics: CompletenessMetrics{
			CompletenessScore: 0.9,
		},
		BusinessMetrics: BusinessContentMetrics{
			BusinessScore: 0.9,
		},
		AssessedAt: time.Now(),
	}

	adjustedDepth := manager.applyQualityAdjustments(baseDepth, highQuality)

	// Check that adjustments were applied
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for high quality adjustments")
	}

	// Check that the reason mentions high quality
	if !contains(adjustedDepth.DepthReason, "High quality") {
		t.Error("Expected DepthReason to mention high quality adjustments")
	}

	// Test low quality adjustments
	lowQuality := &PageContentQuality{
		OverallQuality: 0.2,
		CompletenessMetrics: CompletenessMetrics{
			CompletenessScore: 0.2,
		},
		BusinessMetrics: BusinessContentMetrics{
			BusinessScore: 0.2,
		},
		AssessedAt: time.Now(),
	}

	adjustedDepth = manager.applyQualityAdjustments(baseDepth, lowQuality)

	// Check that adjustments were applied
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for low quality adjustments")
	}

	// Check that the reason mentions low quality
	if !contains(adjustedDepth.DepthReason, "Low quality") {
		t.Error("Expected DepthReason to mention low quality adjustments")
	}
}

func TestRelevanceAdjustments(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	baseDepth := &ScrapingDepth{
		MaxDepth:         3,
		MaxPages:         20,
		MaxLinksPerPage:  10,
		MaxContentLength: 5000,
		CalculatedAt:     time.Now(),
	}

	// Test high relevance adjustments
	highRelevance := &PageRelevanceScore{
		OverallScore: 0.9,
		BusinessRelevance: BusinessRelevance{
			BusinessNameMatch: 0.9,
			IndustryRelevance: 0.9,
		},
		ScoredAt: time.Now(),
	}

	adjustedDepth := manager.applyRelevanceAdjustments(baseDepth, highRelevance)

	// Check that adjustments were applied
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for high relevance adjustments")
	}

	// Check that the reason mentions high relevance
	if !contains(adjustedDepth.DepthReason, "High relevance") {
		t.Error("Expected DepthReason to mention high relevance adjustments")
	}

	if !adjustedDepth.FollowExternal {
		t.Error("Expected high relevance to enable external link following")
	}

	// Test low relevance adjustments
	lowRelevance := &PageRelevanceScore{
		OverallScore: 0.2,
		BusinessRelevance: BusinessRelevance{
			BusinessNameMatch: 0.2,
			IndustryRelevance: 0.2,
		},
		ScoredAt: time.Now(),
	}

	adjustedDepth = manager.applyRelevanceAdjustments(baseDepth, lowRelevance)

	// Check that adjustments were applied
	if adjustedDepth.DepthReason == "" {
		t.Error("Expected DepthReason to be populated for low relevance adjustments")
	}

	// Check that the reason mentions low relevance
	if !contains(adjustedDepth.DepthReason, "Low relevance") {
		t.Error("Expected DepthReason to mention low relevance adjustments")
	}

	if adjustedDepth.FollowExternal {
		t.Error("Expected low relevance to disable external link following")
	}
}

func TestDepthLimits(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test excessive depth limits
	excessiveDepth := &ScrapingDepth{
		MaxDepth:         15,
		MaxPages:         150,
		MaxLinksPerPage:  100,
		MaxContentLength: 100000,
		CalculatedAt:     time.Now(),
	}

	limitedDepth := manager.enforceDepthLimits(excessiveDepth)

	if limitedDepth.MaxDepth > 10 {
		t.Errorf("Expected MaxDepth to be capped at 10, got %d", limitedDepth.MaxDepth)
	}

	if limitedDepth.MaxPages > 100 {
		t.Errorf("Expected MaxPages to be capped at 100, got %d", limitedDepth.MaxPages)
	}

	if limitedDepth.MaxLinksPerPage > 50 {
		t.Errorf("Expected MaxLinksPerPage to be capped at 50, got %d", limitedDepth.MaxLinksPerPage)
	}

	if limitedDepth.MaxContentLength > 50000 {
		t.Errorf("Expected MaxContentLength to be capped at 50000, got %d", limitedDepth.MaxContentLength)
	}

	// Test minimum depth limits
	minimalDepth := &ScrapingDepth{
		MaxDepth:         0,
		MaxPages:         0,
		MaxLinksPerPage:  0,
		MaxContentLength: 50,
		CalculatedAt:     time.Now(),
	}

	enforcedDepth := manager.enforceDepthLimits(minimalDepth)

	if enforcedDepth.MaxDepth < 1 {
		t.Errorf("Expected MinDepth to be enforced at 1, got %d", enforcedDepth.MaxDepth)
	}

	if enforcedDepth.MaxPages < 1 {
		t.Errorf("Expected MinPages to be enforced at 1, got %d", enforcedDepth.MaxPages)
	}

	if enforcedDepth.MaxLinksPerPage < 1 {
		t.Errorf("Expected MinLinksPerPage to be enforced at 1, got %d", enforcedDepth.MaxLinksPerPage)
	}

	if enforcedDepth.MaxContentLength < 100 {
		t.Errorf("Expected MinContentLength to be enforced at 100, got %d", enforcedDepth.MaxContentLength)
	}
}

func TestDelayCalculation(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test high priority delay
	highPriorityDepth := &ScrapingDepth{
		MaxDepth:     5,
		MaxPages:     50,
		CalculatedAt: time.Now(),
	}

	delay := manager.calculateDelay(highPriorityDepth, 0.9)

	if delay > 1.0 {
		t.Errorf("Expected high priority delay <= 1.0, got %f", delay)
	}

	// Test low priority delay
	lowPriorityDepth := &ScrapingDepth{
		MaxDepth:     1,
		MaxPages:     5,
		CalculatedAt: time.Now(),
	}

	delay = manager.calculateDelay(lowPriorityDepth, 0.2)

	if delay < 2.0 {
		t.Errorf("Expected low priority delay >= 2.0, got %f", delay)
	}

	// Test minimum delay
	if delay < 0.1 {
		t.Errorf("Expected minimum delay >= 0.1, got %f", delay)
	}
}

func TestScrapingStrategy(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test deep scraping strategy
	deepDepth := &ScrapingDepth{
		MaxDepth:          5,
		MaxPages:          50,
		MaxLinksPerPage:   20,
		MaxContentLength:  10000,
		FollowExternal:    true,
		FollowSubdomains:  true,
		RespectRobotsTxt:  true,
		DelayBetweenPages: 0.5,
		CalculatedAt:      time.Now(),
	}

	strategy := manager.GetScrapingStrategy(deepDepth)

	if len(strategy) == 0 {
		t.Error("Expected scraping strategy to be populated")
	}

	if !contains(strategy, "Deep crawling") {
		t.Error("Expected strategy to mention deep crawling")
	}

	if !contains(strategy, "50 pages") {
		t.Error("Expected strategy to mention page count")
	}

	// Test minimal scraping strategy
	minimalDepth := &ScrapingDepth{
		MaxDepth:          1,
		MaxPages:          5,
		MaxLinksPerPage:   5,
		MaxContentLength:  1000,
		FollowExternal:    false,
		FollowSubdomains:  false,
		RespectRobotsTxt:  true,
		DelayBetweenPages: 3.0,
		CalculatedAt:      time.Now(),
	}

	strategy = manager.GetScrapingStrategy(minimalDepth)

	if !contains(strategy, "Minimal crawling") {
		t.Error("Expected strategy to mention minimal crawling")
	}

	if !contains(strategy, "5 pages") {
		t.Error("Expected strategy to mention page count")
	}
}

func TestDepthClassification(t *testing.T) {
	config := ScrapingDepthConfig{}
	qualityAssessor := NewPageContentQualityAssessor(ContentQualityConfig{})
	pageTypeDetector := NewPageTypeDetector(PageTypeConfig{})
	relevanceScorer := NewPageRelevanceScorer()

	manager := NewDynamicScrapingDepthManager(config, qualityAssessor, pageTypeDetector, relevanceScorer)

	// Test high priority depth
	highPriorityDepth := &ScrapingDepth{
		PriorityScore: 0.8,
		CalculatedAt:  time.Now(),
	}

	if !manager.IsHighPriorityDepth(highPriorityDepth) {
		t.Error("Expected high priority depth to be classified as high priority")
	}

	// Test deep scraping
	deepDepth := &ScrapingDepth{
		MaxDepth:     5,
		MaxPages:     30,
		CalculatedAt: time.Now(),
	}

	if !manager.IsDeepScraping(deepDepth) {
		t.Error("Expected deep depth to be classified as deep scraping")
	}

	// Test minimal scraping
	minimalDepth := &ScrapingDepth{
		MaxDepth:     1,
		MaxPages:     3,
		CalculatedAt: time.Now(),
	}

	if !manager.IsMinimalScraping(minimalDepth) {
		t.Error("Expected minimal depth to be classified as minimal scraping")
	}

	// Test low priority depth
	lowPriorityDepth := &ScrapingDepth{
		PriorityScore: 0.3,
		CalculatedAt:  time.Now(),
	}

	if manager.IsHighPriorityDepth(lowPriorityDepth) {
		t.Error("Expected low priority depth to not be classified as high priority")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
