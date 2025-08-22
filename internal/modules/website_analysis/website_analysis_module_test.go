package website_analysis

import (
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsiteAnalysisModule(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	assert.NotNil(t, module)
	assert.Equal(t, "website_analysis_module", module.ID())
	assert.False(t, module.IsRunning())
}

func TestWebsiteAnalysisModuleMetadata(t *testing.T) {
	module := NewWebsiteAnalysisModule()
	metadata := module.Metadata()

	assert.Equal(t, "Website Analysis Module", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Performs comprehensive website analysis and content extraction", metadata.Description)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityClassification)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityWebAnalysis)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityDataExtraction)
	assert.Equal(t, architecture.PriorityHigh, metadata.Priority)
}

func TestWebsiteAnalysisModuleCanHandle(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	// Test supported request type
	req := architecture.ModuleRequest{
		Type: "analyze_website",
	}
	assert.True(t, module.CanHandle(req))

	// Test unsupported request type
	req.Type = "unsupported_type"
	assert.False(t, module.CanHandle(req))
}

func TestWebsiteAnalysisModuleHealth(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	// Health when not running
	health := module.Health()
	assert.Equal(t, architecture.ModuleStatusStopped, health.Status)
	assert.Contains(t, health.Message, "Website analysis module")

	// Health when running
	module.running = true
	health = module.Health()
	assert.Equal(t, architecture.ModuleStatusRunning, health.Status)
}

func TestWebsiteAnalysisModuleFactory(t *testing.T) {
	// Create factory
	factory := NewWebsiteAnalysisFactory(nil, nil, nil, nil, nil)

	assert.NotNil(t, factory)
	assert.Equal(t, "website_analysis", factory.GetModuleType())

	// Create module through factory
	moduleConfig := architecture.ModuleConfig{
		Enabled: true,
	}

	module, err := factory.CreateModule(moduleConfig)
	assert.NoError(t, err)
	assert.NotNil(t, module)
	assert.Equal(t, "website_analysis_module", module.ID())

	// Verify module implements Module interface
	_, ok := module.(architecture.Module)
	assert.True(t, ok, "Module should implement architecture.Module interface")
}

func TestWebsiteAnalysisModuleConfiguration(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	// Test scraping configuration
	assert.Equal(t, 30, int(module.scrapingConfig.Timeout.Seconds()))
	assert.Equal(t, 3, module.scrapingConfig.MaxRetries)
	assert.Equal(t, 5, module.scrapingConfig.MaxConcurrent)
	assert.Equal(t, 2, module.scrapingConfig.RateLimitPerSec)
	assert.Len(t, module.scrapingConfig.UserAgents, 3)

	// Test analysis configuration
	assert.Equal(t, 5, module.analysisConfig.MaxPages)
	assert.Equal(t, 100, module.analysisConfig.ContentMinLength)
	assert.Equal(t, 0.6, module.analysisConfig.QualityThreshold)
	assert.True(t, module.analysisConfig.EnableMetaTags)
	assert.True(t, module.analysisConfig.EnableStructuredData)
	assert.True(t, module.analysisConfig.EnableSemanticAnalysis)
}

func TestWebsiteAnalysisModuleCaching(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	// Test cache initialization
	assert.NotNil(t, module.resultCache)
	assert.Equal(t, 2*time.Hour, module.cacheTTL)

	// Test cache key generation
	req := &AnalysisRequest{
		BusinessName:          "Test Business",
		WebsiteURL:            "https://example.com",
		MaxPages:              3,
		IncludeMeta:           true,
		IncludeStructuredData: false,
	}

	cacheKey := module.generateCacheKey(req)
	assert.NotEmpty(t, cacheKey)
	assert.Len(t, cacheKey, 64) // SHA256 hash length
}

func TestWebsiteAnalysisModulePerformanceTracking(t *testing.T) {
	module := NewWebsiteAnalysisModule()

	// Test performance tracking initialization
	assert.NotNil(t, module.analysisTimes)
	assert.NotNil(t, module.successRates)

	// Test performance metrics update
	module.updatePerformanceMetrics("https://example.com", 1500*time.Millisecond)

	module.metricsMutex.RLock()
	duration, exists := module.analysisTimes["https://example.com"]
	module.metricsMutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, 1500*time.Millisecond, duration)
}
