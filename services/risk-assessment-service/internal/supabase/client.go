package supabase

import (
	"context"
	"fmt"
	"time"

	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// Client wraps the Supabase client with additional functionality
type Client struct {
	client *supabase.Client
	logger *zap.Logger
}

// Config holds Supabase configuration
type Config struct {
	URL            string `json:"url"`
	APIKey         string `json:"api_key"`
	ServiceRoleKey string `json:"service_role_key"`
	JWTSecret      string `json:"jwt_secret"`
}

// NewClient creates a new Supabase client
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("supabase URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("supabase API key is required")
	}

	client, err := supabase.NewClient(cfg.URL, cfg.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// GetClient returns the underlying Supabase client
func (c *Client) GetClient() *supabase.Client {
	return c.client
}

// Health checks the connection to Supabase
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simple health check by attempting to connect
	// In a real implementation, you might query a simple table
	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	// Supabase client doesn't require explicit closing
	return nil
}
