package web_search_analysis

import (
	"testing"
	"time"

	"kyb-platform/internal/architecture"
	"github.com/stretchr/testify/assert"
)

func TestNewWebSearchAnalysisModule(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	assert.NotNil(t, module)
	assert.Equal(t, "web_search_analysis_module", module.ID())
	assert.False(t, module.IsRunning())
}

func TestWebSearchAnalysisModuleMetadata(t *testing.T) {
	module := NewWebSearchAnalysisModule()
	metadata := module.Metadata()

	assert.Equal(t, "Web Search Analysis Module", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Performs comprehensive web search analysis and result processing", metadata.Description)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityClassification)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityWebAnalysis)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityDataExtraction)
	assert.Equal(t, architecture.PriorityHigh, metadata.Priority)
}

func TestWebSearchAnalysisModuleCanHandle(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test supported request type
	req := architecture.ModuleRequest{
		Type: "analyze_web_search",
	}
	assert.True(t, module.CanHandle(req))

	// Test unsupported request type
	req.Type = "unsupported_type"
	assert.False(t, module.CanHandle(req))
}

func TestWebSearchAnalysisModuleHealth(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Health when not running
	health := module.Health()
	assert.Equal(t, architecture.ModuleStatusStopped, health.Status)
	assert.Contains(t, health.Message, "Web search analysis module")

	// Health when running
	module.running = true
	health = module.Health()
	assert.Equal(t, architecture.ModuleStatusRunning, health.Status)
}

func TestWebSearchAnalysisModuleFactory(t *testing.T) {
	// Create factory
	factory := NewWebSearchAnalysisFactory(nil, nil, nil, nil, nil)

	assert.NotNil(t, factory)
	assert.Equal(t, "web_search_analysis", factory.GetModuleType())

	// Create module through factory
	moduleConfig := architecture.ModuleConfig{
		Enabled: true,
	}

	module, err := factory.CreateModule(moduleConfig)
	assert.NoError(t, err)
	assert.NotNil(t, module)
	assert.Equal(t, "web_search_analysis_module", module.ID())

	// Verify module implements Module interface
	_, ok := module.(architecture.Module)
	assert.True(t, ok, "Module should implement architecture.Module interface")
}

func TestWebSearchAnalysisModuleConfiguration(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test search configuration
	assert.Equal(t, 10, module.searchConfig.MaxResultsPerEngine)
	assert.Equal(t, 30*time.Second, module.searchConfig.SearchTimeout)
	assert.Equal(t, 3, module.searchConfig.RetryAttempts)
	assert.Equal(t, 1*time.Second, module.searchConfig.RateLimitDelay)
	assert.True(t, module.searchConfig.EnableMultiSource)
	assert.True(t, module.searchConfig.EnableQueryOptimization)
	assert.True(t, module.searchConfig.EnableResultAnalysis)
	assert.True(t, module.searchConfig.EnableResultRanking)
	assert.True(t, module.searchConfig.EnableBusinessExtraction)

	// Test analysis configuration
	assert.Equal(t, 0.3, module.analysisConfig.MinRelevanceScore)
	assert.Equal(t, 20, module.analysisConfig.MaxResultsToAnalyze)
	assert.True(t, module.analysisConfig.EnableSpamDetection)
	assert.True(t, module.analysisConfig.EnableDuplicateDetection)
	assert.True(t, module.analysisConfig.EnableContentAnalysis)
}

func TestWebSearchAnalysisModuleCaching(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test cache initialization
	assert.NotNil(t, module.resultCache)
	assert.Equal(t, 1*time.Hour, module.cacheTTL)

	// Test cache key generation
	req := &SearchRequest{
		BusinessName:  "Test Business",
		SearchQuery:   "test business company",
		BusinessType:  "LLC",
		Industry:      "Technology",
		MaxResults:    10,
		SearchEngines: []string{"google", "bing"},
	}

	cacheKey := module.generateCacheKey(req)
	assert.NotEmpty(t, cacheKey)
	assert.Len(t, cacheKey, 64) // SHA256 hash length
}

func TestWebSearchAnalysisModulePerformanceTracking(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test performance tracking initialization
	assert.NotNil(t, module.searchTimes)
	assert.NotNil(t, module.successRates)

	// Test performance metrics update
	module.updatePerformanceMetrics("test query", 1500*time.Millisecond)

	module.metricsMutex.RLock()
	duration, exists := module.searchTimes["test query"]
	module.metricsMutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, 1500*time.Millisecond, duration)
}

func TestWebSearchAnalysisModuleQueryOptimization(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test query optimization
	originalQuery := "the Digital Health Solutions business company"
	optimizedQuery := module.optimizeQuery(originalQuery)

	// Should remove stop words and add quotes
	assert.Contains(t, optimizedQuery, "Digital Health Solutions")
	assert.NotContains(t, optimizedQuery, "the")
	assert.NotContains(t, optimizedQuery, "business company")
}

func TestWebSearchAnalysisModuleSearchQueryBuilding(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test search query building
	payload := map[string]interface{}{
		"business_name": "Digital Health Solutions",
		"business_type": "LLC",
		"industry":      "Healthcare",
		"address":       "123 Main St, New York, NY",
	}

	query := module.buildSearchQuery(payload)
	assert.Contains(t, query, "Digital Health Solutions")
	assert.Contains(t, query, "LLC")
	assert.Contains(t, query, "Healthcare")
	assert.Contains(t, query, "123 Main St, New York, NY")
	assert.Contains(t, query, "business company")
}

func TestWebSearchAnalysisModuleDuplicateRemoval(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test duplicate removal
	results := []SearchResult{
		{URL: "https://example.com/1", Title: "Result 1"},
		{URL: "https://example.com/2", Title: "Result 2"},
		{URL: "https://example.com/1", Title: "Duplicate Result 1"}, // Duplicate URL
		{URL: "https://example.com/3", Title: "Result 3"},
	}

	uniqueResults := module.removeDuplicates(results)
	assert.Len(t, uniqueResults, 3)
	assert.Equal(t, "https://example.com/1", uniqueResults[0].URL)
	assert.Equal(t, "https://example.com/2", uniqueResults[1].URL)
	assert.Equal(t, "https://example.com/3", uniqueResults[2].URL)
}

func TestWebSearchAnalysisModuleSpamDetection(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test spam detection
	results := []SearchResult{
		{Title: "Legitimate Business", Description: "A legitimate business website"},
		{Title: "Buy Now!", Description: "Click here to buy now and make money fast"},
		{Title: "Another Business", Description: "Another legitimate business"},
		{Title: "Work From Home", Description: "Earn money working from home"},
	}

	spamCount := module.detectSpam(results)
	assert.Equal(t, 2, spamCount) // Should detect 2 spam results
}

func TestWebSearchAnalysisModuleStopWordDetection(t *testing.T) {
	module := NewWebSearchAnalysisModule()

	// Test stop word detection
	assert.True(t, module.isStopWord("the"))
	assert.True(t, module.isStopWord("and"))
	assert.True(t, module.isStopWord("in"))
	assert.False(t, module.isStopWord("business"))
	assert.False(t, module.isStopWord("technology"))
	assert.False(t, module.isStopWord("healthcare"))
}
