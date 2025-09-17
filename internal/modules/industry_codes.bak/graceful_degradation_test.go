package industry_codes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewGracefulDegradationService(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}

	t.Run("creates service with default config", func(t *testing.T) {
		service := NewGracefulDegradationService(db, logger, nil)

		require.NotNil(t, service)
		require.NotNil(t, service.config)
		assert.True(t, service.config.EnableFallback)
		assert.True(t, service.config.EnablePartialResults)
		assert.True(t, service.config.EnableCachedResults)
		assert.Equal(t, 0.6, service.config.PartialResultThreshold)
		assert.Equal(t, 0.3, service.config.MinimalResultThreshold)
		assert.Equal(t, 5*time.Second, service.config.FallbackTimeout)
	})

	t.Run("creates service with custom config", func(t *testing.T) {
		config := &DegradationConfig{
			EnableFallback:         false,
			EnablePartialResults:   true,
			PartialResultThreshold: 0.8,
			FallbackTimeout:        10 * time.Second,
		}

		service := NewGracefulDegradationService(db, logger, config)

		require.NotNil(t, service)
		assert.False(t, service.config.EnableFallback)
		assert.True(t, service.config.EnablePartialResults)
		assert.Equal(t, 0.8, service.config.PartialResultThreshold)
		assert.Equal(t, 10*time.Second, service.config.FallbackTimeout)
	})

	t.Run("initializes fallback data provider", func(t *testing.T) {
		service := NewGracefulDegradationService(db, logger, nil)

		require.NotNil(t, service.fallbackData)
		assert.NotEmpty(t, service.fallbackData.staticData)
		assert.Contains(t, service.fallbackData.staticData, "restaurant")
		assert.Contains(t, service.fallbackData.staticData, "retail")
		assert.Contains(t, service.fallbackData.staticData, "consulting")
	})

	t.Run("initializes alternative scorer", func(t *testing.T) {
		service := NewGracefulDegradationService(db, logger, nil)

		require.NotNil(t, service.alternativeScorer)
		assert.NotEmpty(t, service.alternativeScorer.simpleRules)
	})
}

func TestExecuteWithDegradation_SuccessfulPrimaryOperation(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	expectedCode := &IndustryCode{
		ID:          "test-001",
		Code:        "1234",
		Type:        CodeTypeMCC,
		Description: "Test Industry",
		Confidence:  0.95,
	}

	operation := func() (*IndustryCode, error) {
		return expectedCode, nil
	}

	ctx := context.Background()
	result := service.ExecuteWithDegradation(ctx, operation, "test query")

	assert.True(t, result.Success)
	assert.Equal(t, LevelNone, result.DegradationLevel)
	assert.Equal(t, DegradationStrategy("primary"), result.Strategy)
	assert.Equal(t, 1.0, result.Confidence)
	assert.Equal(t, expectedCode, result.Data)
	assert.Empty(t, result.Fallbacks)
	assert.Empty(t, result.Warnings)
	assert.Greater(t, result.QualityScore, 0.0)
}

func TestExecuteWithDegradation_FallbackStrategies(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}

	tests := []struct {
		name                 string
		config               *DegradationConfig
		fallbackData         interface{}
		expectedStrategy     DegradationStrategy
		expectedSuccess      bool
		expectedLevel        DegradationLevel
		expectCachedStrategy bool
	}{
		{
			name: "cached results strategy",
			config: &DegradationConfig{
				EnableCachedResults:    true,
				EnableFallback:         true,
				EnablePartialResults:   true,
				PartialResultThreshold: 0.6,
				MinimalResultThreshold: 0.3,
				CacheTimeout:           1 * time.Hour,
			},
			fallbackData:         "restaurant",
			expectedStrategy:     StrategyCachedResults,
			expectedSuccess:      true,
			expectedLevel:        LevelPartial,
			expectCachedStrategy: true,
		},
		{
			name: "fallback data strategy",
			config: &DegradationConfig{
				EnableCachedResults:    false,
				EnableFallback:         true,
				EnablePartialResults:   true,
				PartialResultThreshold: 0.6,
				MinimalResultThreshold: 0.3,
			},
			fallbackData:     "restaurant",
			expectedStrategy: StrategyFallbackData,
			expectedSuccess:  true,
			expectedLevel:    LevelPartial,
		},
		{
			name: "partial results strategy",
			config: &DegradationConfig{
				EnableCachedResults:    false,
				EnableFallback:         false,
				EnablePartialResults:   true,
				PartialResultThreshold: 0.6,
				MinimalResultThreshold: 0.3,
			},
			fallbackData:     "consulting services",
			expectedStrategy: StrategyPartialResults,
			expectedSuccess:  true,
			expectedLevel:    LevelMinimal,
		},
		{
			name: "alternative logic strategy",
			config: &DegradationConfig{
				EnableCachedResults:    false,
				EnableFallback:         false,
				EnablePartialResults:   false,
				MinimalResultThreshold: 0.3,
			},
			fallbackData:     "food restaurant",
			expectedStrategy: StrategyAlternativeLogic,
			expectedSuccess:  true,
			expectedLevel:    LevelMinimal,
		},
		{
			name: "static response strategy",
			config: &DegradationConfig{
				EnableCachedResults:    false,
				EnableFallback:         false,
				EnablePartialResults:   false,
				MinimalResultThreshold: 0.9, // High threshold to skip alternative logic
			},
			fallbackData:     "unknown business",
			expectedStrategy: StrategyStaticResponse,
			expectedSuccess:  true,
			expectedLevel:    LevelFallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGracefulDegradationService(db, logger, tt.config)

			// Pre-populate cache if testing cached strategy
			if tt.expectCachedStrategy {
				service.CacheResult("restaurant", &IndustryCode{
					ID:         "cached-001",
					Code:       "5812",
					Type:       CodeTypeMCC,
					Confidence: 0.8,
				}, 0.8)
			}

			// Operation that always fails
			operation := func() (*IndustryCode, error) {
				return nil, errors.New("primary operation failed")
			}

			ctx := context.Background()
			result := service.ExecuteWithDegradation(ctx, operation, tt.fallbackData)

			assert.Equal(t, tt.expectedSuccess, result.Success, "Success mismatch")
			if tt.expectedSuccess {
				assert.Equal(t, tt.expectedStrategy, result.Strategy, "Strategy mismatch")
				assert.Equal(t, tt.expectedLevel, result.DegradationLevel, "Degradation level mismatch")
				assert.Greater(t, result.Confidence, 0.0, "Confidence should be positive")
				assert.NotEmpty(t, result.Fallbacks, "Should have fallback attempts")
				assert.NotEmpty(t, result.Warnings, "Should have warnings")
				assert.NotEmpty(t, result.Recommendations, "Should have recommendations")
			}
		})
	}
}

func TestExecuteWithDegradation_AllStrategiesFail(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}

	// Disable all strategies - even static response should fail with impossible threshold
	config := &DegradationConfig{
		EnableCachedResults:    false,
		EnableFallback:         false,
		EnablePartialResults:   false,
		MinimalResultThreshold: 2.0, // Impossible threshold (>1.0)
	}

	service := NewGracefulDegradationService(db, logger, config)

	operation := func() (*IndustryCode, error) {
		return nil, errors.New("primary operation failed")
	}

	ctx := context.Background()
	result := service.ExecuteWithDegradation(ctx, operation, "test")

	assert.False(t, result.Success)
	assert.Equal(t, LevelCritical, result.DegradationLevel)
	assert.Equal(t, DegradationStrategy("none"), result.Strategy)
	assert.Equal(t, 0.0, result.Confidence)
	assert.NotEmpty(t, result.Fallbacks)
	assert.Contains(t, result.Warnings, "All degradation strategies failed")
	assert.NotEmpty(t, result.Recommendations)
}

func TestTryCachedResults(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{
		EnableCachedResults: true,
		CacheTimeout:        1 * time.Hour,
	}
	service := NewGracefulDegradationService(db, logger, config)

	t.Run("cache hit with valid data", func(t *testing.T) {
		// Add data to cache
		service.CacheResult("test query", &IndustryCode{
			ID:         "cached-001",
			Confidence: 0.8,
		}, 0.8)

		success, quality := service.tryCachedResults(context.Background(), "test query")
		assert.True(t, success)
		assert.Equal(t, 0.8, quality)
	})

	t.Run("cache miss", func(t *testing.T) {
		success, quality := service.tryCachedResults(context.Background(), "nonexistent query")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})

	t.Run("expired cache entry", func(t *testing.T) {
		// Add expired entry
		service.fallbackData.cachedData["expired"] = &CachedResult{
			Code:       &IndustryCode{ID: "expired"},
			Confidence: 0.7,
			Timestamp:  time.Now().Add(-2 * time.Hour), // Expired
		}

		success, quality := service.tryCachedResults(context.Background(), "expired")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)

		// Verify entry was removed
		_, exists := service.fallbackData.cachedData["expired"]
		assert.False(t, exists)
	})

	t.Run("caching disabled", func(t *testing.T) {
		disabledConfig := &DegradationConfig{EnableCachedResults: false}
		disabledService := NewGracefulDegradationService(db, logger, disabledConfig)

		success, quality := disabledService.tryCachedResults(context.Background(), "test")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})
}

func TestTryFallbackData(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{EnableFallback: true}
	service := NewGracefulDegradationService(db, logger, config)

	t.Run("exact match", func(t *testing.T) {
		success, quality := service.tryFallbackData(context.Background(), "restaurant")
		assert.True(t, success)
		assert.Equal(t, 0.7, quality) // Static confidence from initializeStaticData
	})

	t.Run("partial match", func(t *testing.T) {
		success, quality := service.tryFallbackData(context.Background(), "restaurant business")
		assert.True(t, success)
		assert.InDelta(t, 0.56, quality, 0.001) // 0.7 * 0.8 for partial match, with delta for floating point precision
	})

	t.Run("no match", func(t *testing.T) {
		success, quality := service.tryFallbackData(context.Background(), "unknown business type")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})

	t.Run("fallback disabled", func(t *testing.T) {
		disabledConfig := &DegradationConfig{EnableFallback: false}
		disabledService := NewGracefulDegradationService(db, logger, disabledConfig)

		success, quality := disabledService.tryFallbackData(context.Background(), "restaurant")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})
}

func TestTryPartialResults(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{
		EnablePartialResults:   true,
		PartialResultThreshold: 0.6,
		MinimalResultThreshold: 0.3,
	}
	service := NewGracefulDegradationService(db, logger, config)

	t.Run("partial results available", func(t *testing.T) {
		success, quality := service.tryPartialResults(context.Background(), "restaurant business")
		assert.True(t, success)
		assert.Equal(t, 0.6, quality) // Should return partial threshold
	})

	t.Run("threshold too low", func(t *testing.T) {
		lowConfig := &DegradationConfig{
			EnablePartialResults:   true,
			PartialResultThreshold: 0.2, // Below minimal threshold
			MinimalResultThreshold: 0.3,
		}
		lowService := NewGracefulDegradationService(db, logger, lowConfig)

		success, quality := lowService.tryPartialResults(context.Background(), "test")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})

	t.Run("partial results disabled", func(t *testing.T) {
		disabledConfig := &DegradationConfig{EnablePartialResults: false}
		disabledService := NewGracefulDegradationService(db, logger, disabledConfig)

		success, quality := disabledService.tryPartialResults(context.Background(), "restaurant")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})
}

func TestTryAlternativeLogic(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{MinimalResultThreshold: 0.3}
	service := NewGracefulDegradationService(db, logger, config)

	t.Run("keyword match found", func(t *testing.T) {
		success, quality := service.tryAlternativeLogic(context.Background(), "food restaurant")
		assert.True(t, success)
		assert.Equal(t, 0.6, quality) // From simple rules
	})

	t.Run("no keyword match", func(t *testing.T) {
		success, quality := service.tryAlternativeLogic(context.Background(), "unknown business")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})

	t.Run("confidence below threshold", func(t *testing.T) {
		highConfig := &DegradationConfig{MinimalResultThreshold: 0.9}
		highService := NewGracefulDegradationService(db, logger, highConfig)

		success, quality := highService.tryAlternativeLogic(context.Background(), "food")
		assert.False(t, success)
		assert.Equal(t, 0.0, quality)
	})
}

func TestTryStaticResponse(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{MinimalResultThreshold: 0.3}
	service := NewGracefulDegradationService(db, logger, config)

	success, quality := service.tryStaticResponse(context.Background(), "anything")
	assert.True(t, success)
	assert.Equal(t, 0.3, quality) // Should return minimal threshold
}

func TestExtractQueryFromFallbackData(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	tests := []struct {
		name         string
		fallbackData interface{}
		expected     string
	}{
		{
			name:         "string input",
			fallbackData: "test query",
			expected:     "test query",
		},
		{
			name:         "map with query key",
			fallbackData: map[string]interface{}{"query": "map query"},
			expected:     "map query",
		},
		{
			name:         "map with name key",
			fallbackData: map[string]interface{}{"name": "business name"},
			expected:     "business name",
		},
		{
			name:         "map with business_name key",
			fallbackData: map[string]interface{}{"business_name": "company name"},
			expected:     "company name",
		},
		{
			name:         "BusinessData struct",
			fallbackData: integrations.BusinessData{CompanyName: "struct business"},
			expected:     "struct business",
		},
		{
			name:         "nil input",
			fallbackData: nil,
			expected:     "",
		},
		{
			name:         "empty map",
			fallbackData: map[string]interface{}{},
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractQueryFromFallbackData(tt.fallbackData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePartialCode(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	tests := []struct {
		query            string
		expectedCode     string
		expectedCategory string
	}{
		{
			query:            "restaurant food service",
			expectedCode:     "5812",
			expectedCategory: "Food Service",
		},
		{
			query:            "retail store shopping",
			expectedCode:     "5999",
			expectedCategory: "Retail",
		},
		{
			query:            "consulting professional services",
			expectedCode:     "541611",
			expectedCategory: "Professional Services",
		},
		{
			query:            "unknown business type",
			expectedCode:     "9999",
			expectedCategory: "General Business",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			code := service.generatePartialCode(tt.query)
			require.NotNil(t, code)
			assert.Equal(t, tt.expectedCode, code.Code)
			assert.Equal(t, tt.expectedCategory, code.Category)
			assert.Greater(t, code.Confidence, 0.0)
			assert.NotEmpty(t, code.ID)
			assert.NotEmpty(t, code.Description)
		})
	}
}

func TestCalculateQualityScore(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	t.Run("nil code", func(t *testing.T) {
		score := service.calculateQualityScore(nil)
		assert.Equal(t, 0.0, score)
	})

	t.Run("complete code", func(t *testing.T) {
		code := &IndustryCode{
			Code:        "1234",
			Description: "Test Description",
			Category:    "Test Category",
			Type:        CodeTypeMCC,
			Confidence:  0.8,
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour), // 10 days ago
		}

		score := service.calculateQualityScore(code)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("incomplete code", func(t *testing.T) {
		code := &IndustryCode{
			Code:       "1234",
			Confidence: 0.5,
		}

		score := service.calculateQualityScore(code)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})
}

func TestCalculateConfidence(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	tests := []struct {
		strategy    DegradationStrategy
		dataQuality float64
		expected    float64
	}{
		{StrategyCachedResults, 1.0, 0.9},
		{StrategyFallbackData, 1.0, 0.8},
		{StrategyPartialResults, 1.0, 0.7},
		{StrategyAlternativeLogic, 1.0, 0.6},
		{StrategyStaticResponse, 1.0, 0.4},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			confidence := service.calculateConfidence(tt.strategy, tt.dataQuality)
			assert.Equal(t, tt.expected, confidence)
		})
	}
}

func TestGetDegradationLevel(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	tests := []struct {
		strategy DegradationStrategy
		expected DegradationLevel
	}{
		{StrategyCachedResults, LevelPartial},
		{StrategyFallbackData, LevelPartial},
		{StrategyPartialResults, LevelMinimal},
		{StrategyAlternativeLogic, LevelMinimal},
		{StrategyStaticResponse, LevelFallback},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			level := service.getDegradationLevel(tt.strategy)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestCacheResult(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{EnableCachedResults: true}
	service := NewGracefulDegradationService(db, logger, config)

	code := &IndustryCode{
		ID:         "test-001",
		Code:       "1234",
		Confidence: 0.8,
	}

	service.CacheResult("test query", code, 0.9)

	// Verify cache entry
	cached, exists := service.fallbackData.cachedData["test query"]
	require.True(t, exists)
	assert.Equal(t, code, cached.Code)
	assert.Equal(t, 0.9, cached.Confidence)
	assert.Equal(t, "primary_operation", cached.Source)
}

func TestGetDegradationMetrics(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	service := NewGracefulDegradationService(db, logger, nil)

	// Add some cache entries
	service.CacheResult("query1", &IndustryCode{ID: "1"}, 0.8)
	service.CacheResult("query2", &IndustryCode{ID: "2"}, 0.9)

	metrics := service.GetDegradationMetrics()

	assert.Equal(t, 3, metrics["static_data_entries"]) // restaurant, retail, consulting
	assert.Equal(t, 2, metrics["cached_entries"])
	assert.True(t, metrics["fallback_enabled"].(bool))
	assert.True(t, metrics["partial_enabled"].(bool))
	assert.True(t, metrics["cached_enabled"].(bool))
	assert.NotNil(t, metrics["last_updated"])
}

func TestCleanupExpiredCache(t *testing.T) {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{}
	config := &DegradationConfig{
		EnableCachedResults: true,
		CacheTimeout:        1 * time.Hour,
	}
	service := NewGracefulDegradationService(db, logger, config)

	// Add fresh entry
	service.CacheResult("fresh", &IndustryCode{ID: "fresh"}, 0.8)

	// Add expired entry manually
	service.fallbackData.cachedData["expired"] = &CachedResult{
		Code:       &IndustryCode{ID: "expired"},
		Confidence: 0.7,
		Timestamp:  time.Now().Add(-2 * time.Hour), // Expired
	}

	// Verify both entries exist
	assert.Len(t, service.fallbackData.cachedData, 2)

	// Cleanup expired entries
	service.CleanupExpiredCache()

	// Verify only fresh entry remains
	assert.Len(t, service.fallbackData.cachedData, 1)
	_, exists := service.fallbackData.cachedData["fresh"]
	assert.True(t, exists)
	_, exists = service.fallbackData.cachedData["expired"]
	assert.False(t, exists)
}

func TestAlternativeScorer_scoreWithSimpleRules(t *testing.T) {
	scorer := &AlternativeScorer{
		simpleRules: generateSimpleRules(),
		logger:      zap.NewNop(),
	}

	tests := []struct {
		query        string
		expectMatch  bool
		expectedCode string
	}{
		{
			query:        "restaurant food service",
			expectMatch:  true,
			expectedCode: "5812",
		},
		{
			query:        "technology software company",
			expectMatch:  true,
			expectedCode: "7372",
		},
		{
			query:        "healthcare medical clinic",
			expectMatch:  true,
			expectedCode: "8011",
		},
		{
			query:        "unknown business type",
			expectMatch:  false,
			expectedCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			code := scorer.scoreWithSimpleRules(tt.query)

			if tt.expectMatch {
				require.NotNil(t, code)
				assert.Equal(t, tt.expectedCode, code.Code)
				assert.Greater(t, code.Confidence, 0.0)
				assert.Contains(t, code.Description, "Alternative Logic")
			} else {
				assert.Nil(t, code)
			}
		})
	}
}

func TestFallbackDataProvider_initializeStaticData(t *testing.T) {
	provider := &FallbackDataProvider{
		staticData:   make(map[string]*IndustryCode),
		cachedData:   make(map[string]*CachedResult),
		logger:       zap.NewNop(),
		cacheTimeout: 1 * time.Hour,
	}

	provider.initializeStaticData()

	// Verify static data was initialized
	assert.Len(t, provider.staticData, 3)
	assert.Contains(t, provider.staticData, "restaurant")
	assert.Contains(t, provider.staticData, "retail")
	assert.Contains(t, provider.staticData, "consulting")

	// Verify data quality
	restaurant := provider.staticData["restaurant"]
	require.NotNil(t, restaurant)
	assert.Equal(t, "5812", restaurant.Code)
	assert.Equal(t, CodeTypeMCC, restaurant.Type)
	assert.Equal(t, "Food Service", restaurant.Category)
	assert.Equal(t, 0.7, restaurant.Confidence)
}

func TestGenerateSimpleRules(t *testing.T) {
	rules := generateSimpleRules()

	assert.NotEmpty(t, rules)

	// Verify each rule has required fields
	for _, rule := range rules {
		assert.NotEmpty(t, rule.Keywords)
		assert.NotEmpty(t, rule.Code)
		assert.NotEmpty(t, rule.Type)
		assert.Greater(t, rule.Confidence, 0.0)
		assert.NotEmpty(t, rule.Description)
	}

	// Verify specific rules exist
	foundFood := false
	foundTech := false

	for _, rule := range rules {
		if containsSlice(rule.Keywords, "food") {
			foundFood = true
		}
		if containsSlice(rule.Keywords, "technology") {
			foundTech = true
		}
	}

	assert.True(t, foundFood, "Should have food-related rule")
	assert.True(t, foundTech, "Should have technology-related rule")
}

// Helper function to check if slice contains string
func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
