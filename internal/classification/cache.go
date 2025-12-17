package classification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/external"
)

// ClassificationCache provides database-backed caching for classification results
type ClassificationCache struct {
	repo   repository.KeywordRepository
	logger *log.Logger
}

// NewClassificationCache creates a new classification cache
func NewClassificationCache(repo repository.KeywordRepository, logger *log.Logger) *ClassificationCache {
	if logger == nil {
		logger = log.Default()
	}
	return &ClassificationCache{
		repo:   repo,
		logger: logger,
	}
}

// GenerateCacheKey generates a cache key from website content
// Uses SHA-256 hash of: title|meta_desc|about_text|domain
// Accepts external.ScrapedContent which has the structured fields we need
func (c *ClassificationCache) GenerateCacheKey(scrapedContent interface{}, websiteURL string) string {
	// Extract domain from URL
	domain := extractDomain(websiteURL)
	
	// Handle different content types
	var title, textContent string
	
	// Try to extract from external.ScrapedContent (has Title, MetaDesc, AboutText)
	if extContent, ok := scrapedContent.(*external.ScrapedContent); ok {
		title = extContent.Title
		// Combine high-signal fields
		textContent = fmt.Sprintf("%s|%s|%s",
			extContent.MetaDesc,
			extContent.AboutText,
			strings.Join(extContent.Headings, "|"))
	} else if scrapingResult, ok := scrapedContent.(*ScrapingResult); ok {
		// Fallback to ScrapingResult (has TextContent)
		textContent = scrapingResult.TextContent
		textLen := len(textContent)
		if textLen > 500 {
			textContent = textContent[:500]
		}
	} else {
		// Fallback: use URL only
		textContent = websiteURL
	}
	
	// Create deterministic string from content
	// Prioritize high-signal content: title, meta description, about text
	contentStr := fmt.Sprintf(
		"%s|%s|%s|%s",
		getStringValue(title),
		getStringValue(textContent),
		domain,
		websiteURL,
	)
	
	// Hash it
	hash := sha256.Sum256([]byte(contentStr))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached classification result
func (c *ClassificationCache) Get(
	ctx context.Context,
	cacheKey string,
) (*IndustryDetectionResult, error) {
	c.logger.Printf("🔍 [Phase 5] Checking cache for key: %s", cacheKey[:16]+"...")
	
	// Query database
	cached, err := c.repo.GetCachedClassification(ctx, cacheKey)
	if err != nil {
		c.logger.Printf("⚠️ [Phase 5] Cache lookup error: %v", err)
		return nil, err
	}
	
	if cached == nil {
		c.logger.Printf("ℹ️ [Phase 5] Cache miss for key: %s", cacheKey[:16]+"...")
		return nil, nil
	}
	
	c.logger.Printf("✅ [Phase 5] Cache hit for key: %s", cacheKey[:16]+"...")
	
	// Convert cached result to IndustryDetectionResult
	result := &IndustryDetectionResult{
		IndustryName:   cached.IndustryName,
		Confidence:     cached.Confidence,
		Keywords:       cached.Keywords,
		ProcessingTime: cached.ProcessingTime,
		Method:         cached.Method,
		Reasoning:      cached.Reasoning,
		CreatedAt:      cached.CreatedAt,
	}
	
	// Parse explanation if present
	if len(cached.Explanation) > 0 {
		var explanation ClassificationExplanation
		if err := json.Unmarshal(cached.Explanation, &explanation); err == nil {
			result.Explanation = &explanation
		}
	}
	
	// Set cache metadata
	now := time.Now()
	result.CreatedAt = cached.CreatedAt
	if cached.CachedAt != nil {
		result.CachedAt = cached.CachedAt
	} else {
		result.CachedAt = &now
	}
	result.FromCache = true
	
	return result, nil
}

// Set stores a classification result in cache (async, non-blocking)
func (c *ClassificationCache) Set(
	ctx context.Context,
	cacheKey string,
	businessName string,
	websiteURL string,
	result *IndustryDetectionResult,
) error {
	c.logger.Printf("💾 [Phase 5] Caching result for key: %s", cacheKey[:16]+"...")
	
	// Convert IndustryDetectionResult to CachedClassificationResult
	cachedResult := &repository.CachedClassificationResult{
		IndustryName:     result.IndustryName,
		Confidence:       result.Confidence,
		Keywords:         result.Keywords,
		ProcessingTime:   result.ProcessingTime,
		Method:           result.Method,
		Reasoning:        result.Reasoning,
		CreatedAt:        result.CreatedAt,
		LayerUsed:        extractLayerFromMethod(result.Method),
		ProcessingTimeMs: int(result.ProcessingTime.Milliseconds()),
		FromCache:        false, // This is being cached, not from cache
	}
	
	// Serialize explanation if present
	if result.Explanation != nil {
		explanationJSON, err := json.Marshal(result.Explanation)
		if err == nil {
			cachedResult.Explanation = explanationJSON
		}
	}
	
	// Store in cache (async to avoid blocking response)
	go func() {
		cacheCtx := context.Background()
		if err := c.repo.SetCachedClassification(cacheCtx, cacheKey, businessName, websiteURL, cachedResult); err != nil {
			c.logger.Printf("⚠️ [Phase 5] Failed to cache result: %v", err)
		} else {
			c.logger.Printf("✅ [Phase 5] Result cached successfully for key: %s", cacheKey[:16]+"...")
		}
	}()
	
	return nil
}

// GetStats retrieves cache statistics
func (c *ClassificationCache) GetStats(ctx context.Context) (*repository.CacheStats, error) {
	stats, err := c.repo.GetCacheStats(ctx)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// Helper functions

func extractDomain(urlStr string) string {
	// Simple domain extraction
	// Remove protocol
	if idx := strings.Index(urlStr, "://"); idx != -1 {
		urlStr = urlStr[idx+3:]
	}
	// Remove path
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	// Remove port
	if idx := strings.Index(urlStr, ":"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	return urlStr
}

func getStringValue(s string) string {
	if s == "" {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(s))
}

func extractLayerFromMethod(method string) string {
	// Extract layer from method name
	// Examples: "layer1_high_conf", "layer2_better", "layer3_llm"
	if strings.Contains(method, "layer1") {
		return "layer1"
	}
	if strings.Contains(method, "layer2") {
		return "layer2"
	}
	if strings.Contains(method, "layer3") {
		return "layer3"
	}
	return "layer1" // Default
}

