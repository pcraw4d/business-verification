package supabase

import (
	"context"
	"fmt"
	"time"

	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
)

// Client wraps the Supabase client with additional functionality
type Client struct {
	client *supabase.Client
	config *config.SupabaseConfig
	logger *zap.Logger
}

// NewClient creates a new Supabase client
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

	logger.Info("✅ Supabase client initialized",
		zap.String("url", cfg.URL))

	return sc, nil
}

// GetClient returns the underlying Supabase client
func (c *Client) GetClient() *supabase.Client {
	return c.client
}

// GetAuth returns the auth client
func (c *Client) GetAuth() interface{} {
	return nil // Simplified for now
}

// GetRealtime returns the realtime client
func (c *Client) GetRealtime() interface{} {
	return nil // Simplified for now
}

// GetStorage returns the storage client
func (c *Client) GetStorage() interface{} {
	return nil // Simplified for now
}

// ValidateToken validates a JWT token
func (c *Client) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	// Simplified token validation for now
	// In a real implementation, you would validate the JWT token
	return map[string]interface{}{
		"id":    "user_123",
		"email": "user@example.com",
	}, nil
}

// SubscribeToChanges subscribes to real-time changes in a table
func (c *Client) SubscribeToChanges(table string, callback func(interface{})) error {
	// Simplified for now - real-time subscriptions would be implemented here
	return nil
}

// HealthCheck performs a health check on the Supabase connection
func (c *Client) HealthCheck(ctx context.Context) error {
	// Create a timeout context
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to query a simple table to verify connection
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
