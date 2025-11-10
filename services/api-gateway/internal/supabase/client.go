package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// RegisterUser registers a new user with Supabase Auth
func (c *Client) RegisterUser(ctx context.Context, email, password string, userMetadata map[string]interface{}) (map[string]interface{}, error) {
	// Use Supabase Auth API to register user
	// The supabase-go library v0.0.1 may not have direct Auth methods,
	// so we'll use HTTP calls to the Auth API directly
	
	authURL := fmt.Sprintf("%s/auth/v1/signup", c.config.URL)
	
	// Prepare request payload
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	
	// Add user metadata if provided
	if userMetadata != nil {
		payload["data"] = userMetadata
	}
	
	// Make HTTP request to Supabase Auth API
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.config.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Supabase Auth API: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				return nil, fmt.Errorf("registration failed: %s", msg)
			}
		}
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	c.logger.Info("User registered successfully",
		zap.String("email", email),
		zap.Int("status_code", resp.StatusCode))
	
	return result, nil
}
