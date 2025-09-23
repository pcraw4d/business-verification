package classification

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetailEcommerceKeywords tests the retail and e-commerce keywords implementation
// for Task 3.2.4: Add retail and e-commerce keywords
func TestRetailEcommerceKeywords(t *testing.T) {
	t.Run("Task 3.2.4.1: Retail Industry Keywords", func(t *testing.T) {
		testRetailIndustryKeywords(t)
	})

	t.Run("Task 3.2.4.2: E-commerce Industry Keywords", func(t *testing.T) {
		testEcommerceIndustryKeywords(t)
	})

	t.Run("Task 3.2.4.3: Wholesale Industry Keywords", func(t *testing.T) {
		testWholesaleIndustryKeywords(t)
	})

	t.Run("Task 3.2.4.4: Consumer Goods Industry Keywords", func(t *testing.T) {
		testConsumerGoodsIndustryKeywords(t)
	})

	t.Run("Task 3.2.4.5: Keyword Relevance and Classification", func(t *testing.T) {
		testRetailEcommerceClassification(t)
	})
}

// testRetailIndustryKeywords tests retail industry keyword implementation
func testRetailIndustryKeywords(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Get retail industry
	retailIndustry, err := repo.GetIndustryByName(ctx, "Retail")
	require.NoError(t, err)
	require.NotNil(t, retailIndustry)
	assert.Equal(t, "Retail", retailIndustry.Name)

	// Get retail keywords
	keywords, err := repo.GetKeywordsByIndustry(ctx, retailIndustry.ID)
	require.NoError(t, err)
	require.NotNil(t, keywords)

	// Validate minimum keyword count (50+)
	assert.GreaterOrEqual(t, len(keywords), 50, "Retail industry should have 50+ keywords")

	// Validate keyword weights are within range (0.5-1.0)
	for _, keyword := range keywords {
		assert.GreaterOrEqual(t, keyword.Weight, 0.5, "Keyword weight should be >= 0.5")
		assert.LessOrEqual(t, keyword.Weight, 1.0, "Keyword weight should be <= 1.0")
	}

	// Validate high-weight keywords (>= 0.8) for classification accuracy
	highWeightCount := 0
	for _, keyword := range keywords {
		if keyword.Weight >= 0.8 {
			highWeightCount++
		}
	}
	assert.GreaterOrEqual(t, highWeightCount, 10, "Retail industry should have 10+ high-weight keywords (>=0.8)")

	// Validate core retail keywords exist
	coreKeywords := []string{"retail", "store", "shop", "shopping", "merchandise", "inventory", "sales", "customer"}
	keywordMap := make(map[string]float64)
	for _, keyword := range keywords {
		keywordMap[keyword.Keyword] = keyword.Weight
	}

	for _, coreKeyword := range coreKeywords {
		weight, exists := keywordMap[coreKeyword]
		assert.True(t, exists, "Core retail keyword '%s' should exist", coreKeyword)
		if exists {
			assert.GreaterOrEqual(t, weight, 0.8, "Core retail keyword '%s' should have high weight", coreKeyword)
		}
	}

	t.Logf("✅ Retail industry validation passed: %d keywords, %d high-weight keywords", len(keywords), highWeightCount)
}

// testEcommerceIndustryKeywords tests e-commerce industry keyword implementation
func testEcommerceIndustryKeywords(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Get e-commerce industry
	ecommerceIndustry, err := repo.GetIndustryByName(ctx, "E-commerce")
	require.NoError(t, err)
	require.NotNil(t, ecommerceIndustry)
	assert.Equal(t, "E-commerce", ecommerceIndustry.Name)

	// Get e-commerce keywords
	keywords, err := repo.GetKeywordsByIndustry(ctx, ecommerceIndustry.ID)
	require.NoError(t, err)
	require.NotNil(t, keywords)

	// Validate minimum keyword count (50+)
	assert.GreaterOrEqual(t, len(keywords), 50, "E-commerce industry should have 50+ keywords")

	// Validate keyword weights are within range (0.5-1.0)
	for _, keyword := range keywords {
		assert.GreaterOrEqual(t, keyword.Weight, 0.5, "Keyword weight should be >= 0.5")
		assert.LessOrEqual(t, keyword.Weight, 1.0, "Keyword weight should be <= 1.0")
	}

	// Validate high-weight keywords (>= 0.8) for classification accuracy
	highWeightCount := 0
	for _, keyword := range keywords {
		if keyword.Weight >= 0.8 {
			highWeightCount++
		}
	}
	assert.GreaterOrEqual(t, highWeightCount, 10, "E-commerce industry should have 10+ high-weight keywords (>=0.8)")

	// Validate core e-commerce keywords exist
	coreKeywords := []string{"ecommerce", "e-commerce", "online store", "online shop", "online retail", "digital commerce", "web store"}
	keywordMap := make(map[string]float64)
	for _, keyword := range keywords {
		keywordMap[keyword.Keyword] = keyword.Weight
	}

	for _, coreKeyword := range coreKeywords {
		weight, exists := keywordMap[coreKeyword]
		assert.True(t, exists, "Core e-commerce keyword '%s' should exist", coreKeyword)
		if exists {
			assert.GreaterOrEqual(t, weight, 0.8, "Core e-commerce keyword '%s' should have high weight", coreKeyword)
		}
	}

	t.Logf("✅ E-commerce industry validation passed: %d keywords, %d high-weight keywords", len(keywords), highWeightCount)
}

// testWholesaleIndustryKeywords tests wholesale industry keyword implementation
func testWholesaleIndustryKeywords(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Get wholesale industry
	wholesaleIndustry, err := repo.GetIndustryByName(ctx, "Wholesale")
	require.NoError(t, err)
	require.NotNil(t, wholesaleIndustry)
	assert.Equal(t, "Wholesale", wholesaleIndustry.Name)

	// Get wholesale keywords
	keywords, err := repo.GetKeywordsByIndustry(ctx, wholesaleIndustry.ID)
	require.NoError(t, err)
	require.NotNil(t, keywords)

	// Validate minimum keyword count (50+)
	assert.GreaterOrEqual(t, len(keywords), 50, "Wholesale industry should have 50+ keywords")

	// Validate keyword weights are within range (0.5-1.0)
	for _, keyword := range keywords {
		assert.GreaterOrEqual(t, keyword.Weight, 0.5, "Keyword weight should be >= 0.5")
		assert.LessOrEqual(t, keyword.Weight, 1.0, "Keyword weight should be <= 1.0")
	}

	// Validate high-weight keywords (>= 0.8) for classification accuracy
	highWeightCount := 0
	for _, keyword := range keywords {
		if keyword.Weight >= 0.8 {
			highWeightCount++
		}
	}
	assert.GreaterOrEqual(t, highWeightCount, 10, "Wholesale industry should have 10+ high-weight keywords (>=0.8)")

	// Validate core wholesale keywords exist
	coreKeywords := []string{"wholesale", "wholesaler", "wholesale trade", "wholesale business", "wholesale distribution", "b2b", "business to business"}
	keywordMap := make(map[string]float64)
	for _, keyword := range keywords {
		keywordMap[keyword.Keyword] = keyword.Weight
	}

	for _, coreKeyword := range coreKeywords {
		weight, exists := keywordMap[coreKeyword]
		assert.True(t, exists, "Core wholesale keyword '%s' should exist", coreKeyword)
		if exists {
			assert.GreaterOrEqual(t, weight, 0.8, "Core wholesale keyword '%s' should have high weight", coreKeyword)
		}
	}

	t.Logf("✅ Wholesale industry validation passed: %d keywords, %d high-weight keywords", len(keywords), highWeightCount)
}

// testConsumerGoodsIndustryKeywords tests consumer goods industry keyword implementation
func testConsumerGoodsIndustryKeywords(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Get consumer goods industry
	consumerGoodsIndustry, err := repo.GetIndustryByName(ctx, "Consumer Goods")
	require.NoError(t, err)
	require.NotNil(t, consumerGoodsIndustry)
	assert.Equal(t, "Consumer Goods", consumerGoodsIndustry.Name)

	// Get consumer goods keywords
	keywords, err := repo.GetKeywordsByIndustry(ctx, consumerGoodsIndustry.ID)
	require.NoError(t, err)
	require.NotNil(t, keywords)

	// Validate minimum keyword count (50+)
	assert.GreaterOrEqual(t, len(keywords), 50, "Consumer Goods industry should have 50+ keywords")

	// Validate keyword weights are within range (0.5-1.0)
	for _, keyword := range keywords {
		assert.GreaterOrEqual(t, keyword.Weight, 0.5, "Keyword weight should be >= 0.5")
		assert.LessOrEqual(t, keyword.Weight, 1.0, "Keyword weight should be <= 1.0")
	}

	// Validate high-weight keywords (>= 0.8) for classification accuracy
	highWeightCount := 0
	for _, keyword := range keywords {
		if keyword.Weight >= 0.8 {
			highWeightCount++
		}
	}
	assert.GreaterOrEqual(t, highWeightCount, 10, "Consumer Goods industry should have 10+ high-weight keywords (>=0.8)")

	// Validate core consumer goods keywords exist
	coreKeywords := []string{"consumer goods", "consumer products", "consumer items", "consumer merchandise", "consumer brands", "manufacturing", "production"}
	keywordMap := make(map[string]float64)
	for _, keyword := range keywords {
		keywordMap[keyword.Keyword] = keyword.Weight
	}

	for _, coreKeyword := range coreKeywords {
		weight, exists := keywordMap[coreKeyword]
		assert.True(t, exists, "Core consumer goods keyword '%s' should exist", coreKeyword)
		if exists {
			assert.GreaterOrEqual(t, weight, 0.8, "Core consumer goods keyword '%s' should have high weight", coreKeyword)
		}
	}

	t.Logf("✅ Consumer Goods industry validation passed: %d keywords, %d high-weight keywords", len(keywords), highWeightCount)
}

// testRetailEcommerceClassification tests classification accuracy with retail and e-commerce keywords
func testRetailEcommerceClassification(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Test cases for retail and e-commerce classification
	testCases := []struct {
		name             string
		businessName     string
		description      string
		websiteURL       string
		expectedIndustry string
		minConfidence    float64
	}{
		// Retail test cases
		{
			name:             "Traditional Retail Store",
			businessName:     "Macy's Department Store",
			description:      "Traditional department store selling clothing, accessories, and home goods",
			websiteURL:       "",
			expectedIndustry: "Retail",
			minConfidence:    0.75,
		},
		{
			name:             "Specialty Retail Boutique",
			businessName:     "Fashion Forward Boutique",
			description:      "Specialty boutique store offering unique fashion items and accessories",
			websiteURL:       "",
			expectedIndustry: "Retail",
			minConfidence:    0.70,
		},
		{
			name:             "Grocery Store",
			businessName:     "Fresh Market Grocery",
			description:      "Local grocery store providing fresh produce, meat, and household items",
			websiteURL:       "",
			expectedIndustry: "Retail",
			minConfidence:    0.70,
		},
		{
			name:             "Convenience Store",
			businessName:     "Quick Stop Convenience",
			description:      "Convenience store offering snacks, beverages, and basic household items",
			websiteURL:       "",
			expectedIndustry: "Retail",
			minConfidence:    0.70,
		},

		// E-commerce test cases
		{
			name:             "Online Fashion Store",
			businessName:     "StyleHub Online",
			description:      "Online fashion store selling trendy clothing and accessories through our e-commerce platform",
			websiteURL:       "https://stylehub.com",
			expectedIndustry: "E-commerce",
			minConfidence:    0.75,
		},
		{
			name:             "Digital Marketplace",
			businessName:     "TechGear Marketplace",
			description:      "Digital marketplace for electronics and technology products with online shopping and digital commerce solutions",
			websiteURL:       "https://techgear.com",
			expectedIndustry: "E-commerce",
			minConfidence:    0.75,
		},
		{
			name:             "Online Bookstore",
			businessName:     "BookWorld Online",
			description:      "Online bookstore offering books, e-books, and educational materials through our web store",
			websiteURL:       "https://bookworld.com",
			expectedIndustry: "E-commerce",
			minConfidence:    0.70,
		},
		{
			name:             "E-commerce Platform",
			businessName:     "ShopEasy Platform",
			description:      "E-commerce platform providing online retail solutions and digital commerce services",
			websiteURL:       "https://shopeasy.com",
			expectedIndustry: "E-commerce",
			minConfidence:    0.80,
		},

		// Wholesale test cases
		{
			name:             "Wholesale Distributor",
			businessName:     "Global Wholesale Distributors",
			description:      "Wholesale distribution company providing B2B sales and supply chain services",
			websiteURL:       "",
			expectedIndustry: "Wholesale",
			minConfidence:    0.75,
		},
		{
			name:             "B2B Trading Company",
			businessName:     "TradeMaster B2B",
			description:      "Business-to-business trading company specializing in wholesale trade and distribution",
			websiteURL:       "",
			expectedIndustry: "Wholesale",
			minConfidence:    0.75,
		},

		// Consumer Goods test cases
		{
			name:             "Consumer Products Manufacturer",
			businessName:     "HomeGoods Manufacturing",
			description:      "Consumer goods manufacturing company producing household products and consumer items",
			websiteURL:       "",
			expectedIndustry: "Consumer Goods",
			minConfidence:    0.75,
		},
		{
			name:             "Beauty Products Company",
			businessName:     "BeautyBrand Corp",
			description:      "Consumer products company manufacturing beauty products, cosmetics, and personal care items",
			websiteURL:       "",
			expectedIndustry: "Consumer Goods",
			minConfidence:    0.70,
		},
	}

	// Run classification tests
	successCount := 0
	totalTests := len(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Classify business
			result, err := repo.ClassifyBusiness(ctx, tc.businessName, tc.description, tc.websiteURL)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Validate classification result
			assert.Equal(t, tc.expectedIndustry, result.Industry.Name,
				"Business '%s' should be classified as '%s'", tc.businessName, tc.expectedIndustry)
			assert.GreaterOrEqual(t, result.Confidence, tc.minConfidence,
				"Confidence for '%s' should be >= %.2f, got %.2f", tc.businessName, tc.minConfidence, result.Confidence)

			// Log successful classification
			if result.Industry.Name == tc.expectedIndustry && result.Confidence >= tc.minConfidence {
				successCount++
				t.Logf("✅ %s: %s (confidence: %.2f)", tc.name, result.Industry.Name, result.Confidence)
			} else {
				t.Logf("❌ %s: Expected %s (confidence >= %.2f), got %s (confidence: %.2f)",
					tc.name, tc.expectedIndustry, tc.minConfidence, result.Industry.Name, result.Confidence)
			}
		})
	}

	// Calculate and validate overall accuracy
	accuracy := float64(successCount) / float64(totalTests)
	t.Logf("📊 Classification Accuracy: %d/%d (%.1f%%)", successCount, totalTests, accuracy*100)

	// Validate minimum accuracy requirement (80% for retail/e-commerce)
	assert.GreaterOrEqual(t, accuracy, 0.80,
		"Retail and e-commerce classification accuracy should be >= 80%%, got %.1f%%", accuracy*100)

	t.Logf("✅ Retail and e-commerce classification testing completed successfully")
}

// setupTestRepository creates a test repository for testing
func setupTestRepository(t *testing.T) repository.KeywordRepository {
	// This would be implemented based on your actual repository setup
	// For now, we'll use a mock or create a test-specific repository
	t.Skip("Repository setup needs to be implemented based on your actual repository structure")
	return nil
}

// TestRetailEcommerceKeywordsPerformance tests the performance of retail and e-commerce keyword classification
func TestRetailEcommerceKeywordsPerformance(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Test performance with various business descriptions
	testCases := []struct {
		name        string
		description string
		maxTime     time.Duration
	}{
		{
			name:        "Short Description",
			description: "Retail store selling clothing and accessories",
			maxTime:     100 * time.Millisecond,
		},
		{
			name:        "Medium Description",
			description: "Online e-commerce platform providing digital commerce solutions for retail businesses with comprehensive online shopping experience and digital marketplace services",
			maxTime:     200 * time.Millisecond,
		},
		{
			name:        "Long Description",
			description: "Comprehensive wholesale distribution company specializing in B2B sales and supply chain management services for consumer goods manufacturers, providing business-to-business trading solutions, wholesale trade services, distribution network management, and supply chain optimization for retail and e-commerce businesses across multiple industries",
			maxTime:     300 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()

			result, err := repo.ClassifyBusiness(ctx, "Test Business", tc.description, "")

			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Less(t, duration, tc.maxTime,
				"Classification should complete within %v, took %v", tc.maxTime, duration)

			t.Logf("✅ %s: Classified as %s in %v", tc.name, result.Industry.Name, duration)
		})
	}
}

// TestRetailEcommerceKeywordsEdgeCases tests edge cases for retail and e-commerce keyword classification
func TestRetailEcommerceKeywordsEdgeCases(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)

	// Test edge cases
	edgeCases := []struct {
		name         string
		businessName string
		description  string
		websiteURL   string
		expectError  bool
	}{
		{
			name:         "Empty Description",
			businessName: "Test Store",
			description:  "",
			websiteURL:   "",
			expectError:  false,
		},
		{
			name:         "Very Long Description",
			businessName: "Test Business",
			description:  "This is a very long description that contains many retail and e-commerce related terms repeated multiple times to test the system's ability to handle large amounts of text while maintaining performance and accuracy in classification. The description includes terms like retail, store, shop, shopping, merchandise, inventory, sales, customer, ecommerce, e-commerce, online store, online shop, online retail, digital commerce, web store, internet store, online marketplace, digital storefront, wholesale, wholesaler, wholesale trade, wholesale business, wholesale distribution, wholesale sales, wholesale operations, wholesale market, wholesale pricing, wholesale supplier, b2b, business to business, b2b sales, b2b trade, b2b commerce, b2b marketplace, b2b platform, b2b services, b2b solutions, b2b network, distribution, distributor, distribution center, distribution network, distribution channel, distribution system, distribution services, distribution operations, distribution management, distribution logistics, trade, trading, trader, trading company, trading house, trading platform, trading network, trading services, trading operations, trading business, supply chain, supply chain management, supply chain services, supply chain solutions, supply chain network, supply chain operations, supply chain logistics, supply chain optimization, supply chain integration, supply chain technology, consumer goods, consumer products, consumer items, consumer merchandise, consumer brands, consumer market, consumer sales, consumer business, consumer industry, consumer sector, household goods, personal care, beauty products, cosmetics, skincare, hair care, body care, health products, wellness products, lifestyle products, manufacturing, manufacturer, manufacturing company, manufacturing business, manufacturing operations, production, producer, production company, production facility, production line, brand, branding, brand management, brand development, brand strategy, brand marketing, brand promotion, brand awareness, brand recognition, brand loyalty, market, marketplace, market research, market analysis, market strategy, market development, market penetration, market share, market position, market leader, sales, sales team, sales management, sales strategy, sales development, sales growth, sales performance, sales targets, sales goals, sales objectives",
			websiteURL:   "",
			expectError:  false,
		},
		{
			name:         "Special Characters",
			businessName: "Café & Bistro",
			description:  "French café serving coffee, pastries, and light meals with retail sales of coffee beans and accessories",
			websiteURL:   "",
			expectError:  false,
		},
		{
			name:         "Mixed Case",
			businessName: "Retail Store",
			description:  "RETAIL STORE selling RETAIL products with RETAIL services and RETAIL operations",
			websiteURL:   "",
			expectError:  false,
		},
		{
			name:         "Numbers and Symbols",
			businessName: "Store #1",
			description:  "Retail store #1 selling products at 50% off with $100 minimum purchase",
			websiteURL:   "",
			expectError:  false,
		},
	}

	for _, ec := range edgeCases {
		t.Run(ec.name, func(t *testing.T) {
			result, err := repo.ClassifyBusiness(ctx, ec.businessName, ec.description, ec.websiteURL)

			if ec.expectError {
				assert.Error(t, err, "Expected error for edge case: %s", ec.name)
			} else {
				require.NoError(t, err, "Unexpected error for edge case: %s", ec.name)
				require.NotNil(t, result, "Result should not be nil for edge case: %s", ec.name)
				assert.Greater(t, result.Confidence, 0.0, "Confidence should be greater than 0 for edge case: %s", ec.name)
				t.Logf("✅ %s: Classified as %s (confidence: %.2f)", ec.name, result.Industry.Name, result.Confidence)
			}
		})
	}
}
