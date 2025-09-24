package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewMarketCoverageExtractor(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	assert.NotNil(t, extractor)
	assert.NotNil(t, extractor.config)
	assert.Equal(t, 0.3, extractor.config.MinConfidenceScore)
	assert.Equal(t, 10, extractor.config.MaxServiceAreas)
	assert.NotEmpty(t, extractor.config.ServiceAreaPatterns)
	assert.NotEmpty(t, extractor.config.MarketCoverageIndicators)
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_LocalService(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We provide local delivery services within 25 miles of our location.
		Serving the greater New York area with fast and reliable service.
		Our local service area covers Manhattan, Brooklyn, and Queens.
		Available in the tri-state area for all your needs.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find local service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1)
	assert.Equal(t, "local", result.CoverageType)

	// Check for radius information
	foundRadius := false
	for _, area := range result.ServiceAreas {
		if area.Radius != nil && *area.Radius == 25 {
			foundRadius = true
			break
		}
	}
	assert.True(t, foundRadius, "Should find 25-mile radius")

	// Check for geographic information - the extractor should find service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1, "Should find at least one service area")

	// Check that we have some geographic information
	foundGeographicInfo := false
	for _, area := range result.ServiceAreas {
		if area.Name != "" || area.Description != "" {
			foundGeographicInfo = true
			break
		}
	}
	assert.True(t, foundGeographicInfo, "Should find geographic information in service areas")
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_RegionalService(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We serve the entire Northeast region including New York, New Jersey, and Connecticut.
		Regional coverage throughout the Mid-Atlantic states.
		Operating across multiple states in the Northeast corridor.
		Available throughout the New England region.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find regional service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1)
	assert.Equal(t, "regional", result.CoverageType)

	// Check for regional indicators
	foundRegional := false
	for _, area := range result.ServiceAreas {
		if area.Type == "regional" {
			foundRegional = true
			break
		}
	}
	assert.True(t, foundRegional, "Should find regional service areas")

	// Check for regional service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1, "Should find at least one service area")

	// Check that we have some service area information
	foundServiceAreaInfo := false
	for _, area := range result.ServiceAreas {
		if area.Name != "" || area.Description != "" {
			foundServiceAreaInfo = true
			break
		}
	}
	assert.True(t, foundServiceAreaInfo, "Should find service area information")
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_NationalService(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We provide nationwide service across all 50 states.
		National coverage with offices in every major city.
		Serving the entire country with our comprehensive solutions.
		Available nationwide for all your business needs.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find national service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1)
	assert.Equal(t, "national", result.CoverageType)

	// Check for national service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1, "Should find at least one service area")

	// Check that we have some service area information
	foundServiceAreaInfo := false
	for _, area := range result.ServiceAreas {
		if area.Name != "" || area.Description != "" {
			foundServiceAreaInfo = true
			break
		}
	}
	assert.True(t, foundServiceAreaInfo, "Should find service area information")
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_InternationalService(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We operate internationally with offices in 25 countries.
		Global coverage serving customers worldwide.
		International service available in Europe, Asia, and North America.
		Worldwide presence with local expertise.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find international service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1)
	assert.Equal(t, "international", result.CoverageType)

	// Check for international service areas
	assert.GreaterOrEqual(t, len(result.ServiceAreas), 1, "Should find at least one service area")

	// Check that we have some service area information
	foundServiceAreaInfo := false
	for _, area := range result.ServiceAreas {
		if area.Name != "" || area.Description != "" {
			foundServiceAreaInfo = true
			break
		}
	}
	assert.True(t, foundServiceAreaInfo, "Should find service area information")
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_TargetMarkets(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We specialize in serving small businesses and enterprise clients.
		Our B2B solutions are designed for mid-market companies.
		Serving the healthcare and education sectors.
		Available for retail and manufacturing industries.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find target markets
	assert.NotEmpty(t, result.TargetMarkets)
	assert.GreaterOrEqual(t, len(result.TargetMarkets), 3)

	// Check for specific target markets
	targetMarkets := make(map[string]bool)
	for _, market := range result.TargetMarkets {
		targetMarkets[market] = true
	}

	assert.True(t, targetMarkets["small business"], "Should find small business target market")
	assert.True(t, targetMarkets["enterprise"], "Should find enterprise target market")
	assert.True(t, targetMarkets["b2b"], "Should find B2B target market")
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_EmptyContent(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	result, err := extractor.ExtractMarketCoverage(context.Background(), "")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should handle empty content gracefully
	assert.Empty(t, result.ServiceAreas)
	assert.Equal(t, 0.0, result.ConfidenceScore)
	assert.Equal(t, "unknown", result.CoverageType)
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_MaxServiceAreas(t *testing.T) {
	logger := zap.NewNop()
	config := &MarketCoverageExtractorConfig{
		MinConfidenceScore: 0.1,
		MaxServiceAreas:    3,
	}
	extractor := NewMarketCoverageExtractor(logger, config)

	content := `
		We serve local areas within 10 miles. Regional coverage in the Northeast.
		National service available. International operations in Europe.
		Global presence worldwide. Local delivery in every city.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should respect max service areas limit
	assert.LessOrEqual(t, len(result.ServiceAreas), 3)
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_ConfidenceScoring(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We provide nationwide service across all 50 states with comprehensive coverage.
		Serving small businesses and enterprise clients throughout the country.
		Available in major cities including New York, Los Angeles, and Chicago.
	`

	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should have reasonable confidence scores
	assert.Greater(t, result.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.ConfidenceScore, 1.0)

	// Individual service areas should have confidence scores
	for _, area := range result.ServiceAreas {
		assert.Greater(t, area.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, area.ConfidenceScore, 1.0)
	}

	// Market coverage should have confidence score
	if result.MarketCoverage != nil {
		assert.Greater(t, result.MarketCoverage.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, result.MarketCoverage.ConfidenceScore, 1.0)
	}
}

func TestMarketCoverageExtractor_ExtractMarketCoverage_ProcessingTime(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := `
		We provide comprehensive nationwide service with local offices in every state.
		Serving small businesses, enterprise clients, and government agencies.
		Available in major metropolitan areas and rural communities.
	`

	startTime := time.Now()
	result, err := extractor.ExtractMarketCoverage(context.Background(), content)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should complete within reasonable time
	assert.Less(t, duration, 100*time.Millisecond)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
}

func TestMarketCoverageExtractor_extractServiceAreas(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := "We serve within 50 miles of our location and provide regional coverage."

	serviceAreas, err := extractor.extractServiceAreas(context.Background(), content)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceAreas)

	// Should find service areas with radius
	foundRadius := false
	for _, area := range serviceAreas {
		if area.Radius != nil && *area.Radius == 50 {
			foundRadius = true
			assert.Equal(t, "miles", area.RadiusUnit)
			break
		}
	}
	assert.True(t, foundRadius, "Should find 50-mile radius")
}

func TestMarketCoverageExtractor_extractMarketCoverage(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := "We provide nationwide service for small businesses and enterprise clients."

	marketCoverage, err := extractor.extractMarketCoverage(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, marketCoverage)

	assert.Equal(t, "national", marketCoverage.Type)
	assert.NotEmpty(t, marketCoverage.TargetMarkets)
	assert.Greater(t, marketCoverage.ConfidenceScore, 0.0)
}

func TestMarketCoverageExtractor_determineMarketCoverageType(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	tests := []struct {
		content  string
		expected string
	}{
		{"We provide global service worldwide", "international"},
		{"National coverage across the country", "national"},
		{"Regional service in the Northeast", "regional"},
		{"Local delivery within 10 miles", "local"},
		{"We provide excellent service", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			result := extractor.determineMarketCoverageType(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarketCoverageExtractor_extractTargetMarkets(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := "We serve small businesses, enterprise clients, and government agencies."

	targetMarkets := extractor.extractTargetMarkets(content)
	assert.NotEmpty(t, targetMarkets)
	assert.Contains(t, targetMarkets, "small business")
	assert.Contains(t, targetMarkets, "enterprise")
	assert.Contains(t, targetMarkets, "government")
}

func TestMarketCoverageExtractor_deduplicateServiceAreas(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	areas := []ServiceArea{
		{Type: "local", Name: "Area 1"},
		{Type: "local", Name: "Area 1"}, // Duplicate
		{Type: "regional", Name: "Area 2"},
		{Type: "regional", Name: "Area 2"}, // Duplicate
	}

	unique := extractor.deduplicateServiceAreas(areas)
	assert.Equal(t, 2, len(unique))
	assert.Equal(t, "local", unique[0].Type)
	assert.Equal(t, "regional", unique[1].Type)
}

func TestMarketCoverageExtractor_validateServiceAreas(t *testing.T) {
	logger := zap.NewNop()
	config := &MarketCoverageExtractorConfig{
		MinConfidenceScore: 0.5,
	}
	extractor := NewMarketCoverageExtractor(logger, config)

	areas := []ServiceArea{
		{ConfidenceScore: 0.8}, // Should pass
		{ConfidenceScore: 0.3}, // Should fail
		{ConfidenceScore: 0.9}, // Should pass
	}

	valid := extractor.validateServiceAreas(areas)
	assert.Equal(t, 2, len(valid))
	assert.Equal(t, 0.8, valid[0].ConfidenceScore)
	assert.Equal(t, 0.9, valid[1].ConfidenceScore)
}

func TestMarketCoverageExtractor_calculateOverallConfidence(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	result := &MarketCoverageResult{
		ServiceAreas: []ServiceArea{
			{ConfidenceScore: 0.8},
			{ConfidenceScore: 0.9},
		},
		MarketCoverage: &MarketCoverage{
			ConfidenceScore: 0.7,
		},
		GeographicScope: "national",
		TargetMarkets:   []string{"small business", "enterprise"},
	}

	confidence := extractor.calculateOverallConfidence(result)
	assert.Greater(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestMarketCoverageExtractor_collectEvidence(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewMarketCoverageExtractor(logger, nil)

	content := "We provide service area coverage throughout the region."
	serviceAreas := []ServiceArea{
		{Name: "Local Area", Type: "local"},
	}
	marketCoverage := &MarketCoverage{
		Description: "Regional coverage",
	}

	evidence := extractor.collectEvidence(content, serviceAreas, marketCoverage)
	assert.NotEmpty(t, evidence)
	assert.Contains(t, evidence[0], "Service area: Local Area")
	assert.Contains(t, evidence[1], "Market coverage: Regional coverage")
}
