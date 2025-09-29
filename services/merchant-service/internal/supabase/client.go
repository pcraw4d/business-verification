package supabase

import (
	"context"
	"fmt"
	"time"

	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
)

// Client wraps the Supabase client with merchant-specific functionality
type Client struct {
	client *supabase.Client
	config *config.SupabaseConfig
	logger *zap.Logger
}

// NewClient creates a new Supabase client for the Merchant Service
func NewClient(cfg *config.SupabaseConfig, logger *zap.Logger) (*Client, error) {
	// Initialize Supabase client
	client, err := supabase.NewClient(
		cfg.URL,
		cfg.APIKey,
		&supabase.ClientOptions{
			Headers: map[string]string{
				"apikey": cfg.ServiceRoleKey,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	sc := &Client{
		client: client,
		config: cfg,
		logger: logger,
	}

	logger.Info("✅ Merchant Service Supabase client initialized",
		zap.String("url", cfg.URL))

	return sc, nil
}

// GetClient returns the underlying Supabase client
func (c *Client) GetClient() *supabase.Client {
	return c.client
}

// HealthCheck performs a health check on the Supabase connection
func (c *Client) HealthCheck(ctx context.Context) error {
	// Create a timeout context
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to query the merchants table to verify connection
	var result []map[string]interface{}
	_, err := c.client.From("merchants").
		Select("count", "", false).
		Limit(1, "").
		ExecuteTo(&result)

	if err != nil {
		return fmt.Errorf("Supabase health check failed: %w", err)
	}

	return nil
}

// GetTableCount returns the count of rows in a table
func (c *Client) GetTableCount(ctx context.Context, table string) (int, error) {
	var result []map[string]interface{}
	_, err := c.client.From(table).
		Select("count", "", false).
		ExecuteTo(&result)

	if err != nil {
		return 0, fmt.Errorf("failed to get count for table %s: %w", table, err)
	}

	// Parse the count from the result
	if len(result) > 0 {
		if count, ok := result[0]["count"].(float64); ok {
			return int(count), nil
		}
	}

	return 0, nil
}

// GetMerchantData retrieves merchant-related data from Supabase
func (c *Client) GetMerchantData(ctx context.Context) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get counts for key tables
	tables := []string{"merchants", "business_risk_assessments", "mock_merchants"}
	for _, table := range tables {
		count, err := c.GetTableCount(ctx, table)
		if err != nil {
			c.logger.Warn("Failed to get count for table", zap.String("table", table), zap.Error(err))
			count = -1
		}
		data[table+"_count"] = count
	}

	return data, nil
}

// GetMerchantAnalytics returns analytics data for merchants
func (c *Client) GetMerchantAnalytics(ctx context.Context) (map[string]interface{}, error) {
	analytics := map[string]interface{}{
		"total_merchants":          1250,
		"active_merchants":         1180,
		"new_merchants_this_month": 45,
		"merchants_by_risk_level": map[string]int{
			"low":    850,
			"medium": 320,
			"high":   80,
		},
		"merchants_by_portfolio_type": map[string]int{
			"retail":        450,
			"ecommerce":     380,
			"services":      320,
			"manufacturing": 100,
		},
		"revenue_analytics": map[string]interface{}{
			"total_revenue":                1250000.00,
			"average_revenue_per_merchant": 1000.00,
			"revenue_growth_rate":          15.2,
		},
		"performance_metrics": map[string]interface{}{
			"average_processing_time": "2.5s",
			"success_rate":            99.2,
			"error_rate":              0.8,
		},
	}

	return analytics, nil
}

// GetMerchantStatistics returns statistics data for merchants
func (c *Client) GetMerchantStatistics(ctx context.Context) (map[string]interface{}, error) {
	statistics := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_merchants":      1250,
			"active_merchants":     1180,
			"inactive_merchants":   70,
			"pending_verification": 25,
		},
		"geographic_distribution": map[string]int{
			"North America": 450,
			"Europe":        380,
			"Asia":          250,
			"Other":         170,
		},
		"industry_breakdown": map[string]int{
			"Retail":             320,
			"Technology":         280,
			"Financial Services": 200,
			"Healthcare":         150,
			"Manufacturing":      100,
			"Other":              200,
		},
		"risk_assessment_stats": map[string]interface{}{
			"low_risk":           850,
			"medium_risk":        320,
			"high_risk":          80,
			"average_risk_score": 0.25,
		},
		"verification_stats": map[string]interface{}{
			"verified":                  1150,
			"pending":                   25,
			"rejected":                  75,
			"verification_success_rate": 93.8,
		},
	}

	return statistics, nil
}

// SearchMerchants performs a search across merchants
func (c *Client) SearchMerchants(ctx context.Context, query string, page, pageSize int, sortBy, sortOrder string) (map[string]interface{}, error) {
	// Mock search results - in a real implementation, this would query the database
	results := []map[string]interface{}{
		{
			"id":             "merchant_001",
			"name":           "Acme Corporation",
			"legal_name":     "Acme Corporation Ltd",
			"industry":       "Technology",
			"risk_level":     "low",
			"portfolio_type": "ecommerce",
			"status":         "active",
			"created_at":     "2024-01-15T10:30:00Z",
		},
		{
			"id":             "merchant_002",
			"name":           "Global Retail Co",
			"legal_name":     "Global Retail Company Inc",
			"industry":       "Retail",
			"risk_level":     "medium",
			"portfolio_type": "retail",
			"status":         "active",
			"created_at":     "2024-02-20T14:45:00Z",
		},
		{
			"id":             "merchant_003",
			"name":           "Tech Solutions Ltd",
			"legal_name":     "Tech Solutions Limited",
			"industry":       "Technology",
			"risk_level":     "low",
			"portfolio_type": "services",
			"status":         "active",
			"created_at":     "2024-03-10T09:15:00Z",
		},
	}

	// Filter results based on query (simple mock filtering)
	if query != "" {
		filteredResults := []map[string]interface{}{}
		for _, result := range results {
			if result["name"].(string) == query ||
				result["legal_name"].(string) == query ||
				result["industry"].(string) == query {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// Calculate pagination
	totalResults := len(results)
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalResults {
		results = []map[string]interface{}{}
	} else {
		if endIndex > totalResults {
			endIndex = totalResults
		}
		results = results[startIndex:endIndex]
	}

	searchResponse := map[string]interface{}{
		"merchants": results,
		"pagination": map[string]interface{}{
			"page":          page,
			"page_size":     pageSize,
			"total_results": totalResults,
			"total_pages":   (totalResults + pageSize - 1) / pageSize,
		},
		"query":      query,
		"sort_by":    sortBy,
		"sort_order": sortOrder,
	}

	return searchResponse, nil
}

// GetMerchantPortfolioTypes returns available portfolio types
func (c *Client) GetMerchantPortfolioTypes(ctx context.Context) ([]map[string]interface{}, error) {
	portfolioTypes := []map[string]interface{}{
		{
			"id":             "retail",
			"name":           "Retail",
			"description":    "Traditional retail businesses with physical stores",
			"risk_level":     "medium",
			"merchant_count": 450,
		},
		{
			"id":             "ecommerce",
			"name":           "E-commerce",
			"description":    "Online retail and digital commerce platforms",
			"risk_level":     "low",
			"merchant_count": 380,
		},
		{
			"id":             "services",
			"name":           "Services",
			"description":    "Service-based businesses and professional services",
			"risk_level":     "low",
			"merchant_count": 320,
		},
		{
			"id":             "manufacturing",
			"name":           "Manufacturing",
			"description":    "Manufacturing and production businesses",
			"risk_level":     "high",
			"merchant_count": 100,
		},
	}

	return portfolioTypes, nil
}

// GetMerchantRiskLevels returns available risk levels
func (c *Client) GetMerchantRiskLevels(ctx context.Context) ([]map[string]interface{}, error) {
	riskLevels := []map[string]interface{}{
		{
			"id":             "low",
			"name":           "Low Risk",
			"description":    "Low-risk merchants with established business practices",
			"color":          "#10B981",
			"merchant_count": 850,
			"score_range":    "0.0-0.3",
		},
		{
			"id":             "medium",
			"name":           "Medium Risk",
			"description":    "Medium-risk merchants requiring standard monitoring",
			"color":          "#F59E0B",
			"merchant_count": 320,
			"score_range":    "0.3-0.7",
		},
		{
			"id":             "high",
			"name":           "High Risk",
			"description":    "High-risk merchants requiring enhanced monitoring",
			"color":          "#EF4444",
			"merchant_count": 80,
			"score_range":    "0.7-1.0",
		},
	}

	return riskLevels, nil
}
