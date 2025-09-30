package supabase

import (
	"context"
	"fmt"
	"time"

	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"

	"kyb-platform/services/classification-service/internal/config"
)

// Client wraps the Supabase client with classification-specific functionality
type Client struct {
	client *supabase.Client
	config *config.SupabaseConfig
	logger *zap.Logger
}

// NewClient creates a new Supabase client for the Classification Service
func NewClient(cfg *config.SupabaseConfig, logger *zap.Logger) (*Client, error) {
	// Initialize Supabase client with correct parameters
	client, err := supabase.NewClient(
		cfg.URL,
		cfg.APIKey,
		&supabase.ClientOptions{
			Headers: map[string]string{
				"apikey":        cfg.APIKey,
				"Authorization": "Bearer " + cfg.APIKey,
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

	logger.Info("✅ Classification Service Supabase client initialized",
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

	// Try to query the classifications table to verify connection
	var result []map[string]interface{}
	_, err := c.client.From("classifications").
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

// GetClassificationData retrieves classification-related data from Supabase
func (c *Client) GetClassificationData(ctx context.Context) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get counts for key tables
	tables := []string{"classifications", "risk_keywords", "industry_code_crosswalks"}
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
