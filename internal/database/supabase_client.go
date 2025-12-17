package database

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	postgrest "github.com/supabase-community/postgrest-go"
	supa "github.com/supabase-community/supabase-go"
)

// SupabaseConfig holds Supabase-specific configuration
type SupabaseConfig struct {
	URL            string `json:"url" yaml:"url"`
	APIKey         string `json:"api_key" yaml:"api_key"`
	ServiceRoleKey string `json:"service_role_key" yaml:"service_role_key"`
	JWTSecret      string `json:"jwt_secret" yaml:"jwt_secret"`
}

// SupabaseClient represents a Supabase database client
type SupabaseClient struct {
	client     *supa.Client
	postgrest  *postgrest.Client
	url        string
	apiKey     string
	serviceKey string
	logger     *log.Logger
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(cfg *SupabaseConfig, logger *log.Logger) (*SupabaseClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("SUPABASE_API_KEY is required")
	}
	if cfg.ServiceRoleKey == "" {
		return nil, fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is required")
	}

	// Create Supabase client
	client, err := supa.NewClient(cfg.URL, cfg.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	// Create PostgREST client for direct database operations
	// Use service role key for better access to all tables
	// Note: Supabase uses HTTP/REST, not direct PostgreSQL connections
	// Connection pooling is handled by the HTTP client (Go's default HTTP client)
	postgrestClient := postgrest.NewClient(
		cfg.URL+"/rest/v1",
		"public",
		map[string]string{
			"apikey":        cfg.ServiceRoleKey,
			"Authorization": "Bearer " + cfg.ServiceRoleKey,
		},
	)

	if logger == nil {
		logger = log.Default()
	}

	// Phase 5: Log HTTP client configuration note
	// Note: Supabase uses HTTP/REST, not direct PostgreSQL connections
	// Connection pooling is handled by Go's HTTP client (default transport)
	// The PostgREST library uses Go's default HTTP client which has connection pooling built-in
	// Optimization: HTTP client connection pooling is optimized at the application level
	// (see embedding_classifier.go and llm_classifier.go for HTTP client optimizations)
	logger.Printf("ℹ️ [Phase 5] [CONNECTION-POOL] Supabase client uses HTTP/REST (connection pooling handled by Go HTTP client with optimized transport)")

	return &SupabaseClient{
		client:     client,
		postgrest:  postgrestClient,
		url:        cfg.URL,
		apiKey:     cfg.APIKey,
		serviceKey: cfg.ServiceRoleKey,
		logger:     logger,
	}, nil
}

// Connect establishes a connection to Supabase
func (s *SupabaseClient) Connect(ctx context.Context) error {
	s.logger.Printf("🔌 Connecting to Supabase at %s", s.url)

	// Test the connection by making a simple query
	// We'll test with a simple health check query
	err := s.Ping(ctx)
	if err != nil {
		s.logger.Printf("❌ Failed to connect to Supabase: %v", err)
		return fmt.Errorf("failed to connect to Supabase: %w", err)
	}

	s.logger.Printf("✅ Successfully connected to Supabase")
	return nil
}

// Close closes the Supabase connection
func (s *SupabaseClient) Close() error {
	s.logger.Printf("🔌 Closing Supabase connection")
	// Supabase client doesn't require explicit closing
	return nil
}

// Ping checks the connection to Supabase
func (s *SupabaseClient) Ping(ctx context.Context) error {
	// Test connection with a simple health check
	// Try to access the Supabase API directly without requiring specific tables
	_, _, err := s.postgrest.From("_health").Select("*", "", false).Execute()
	if err != nil {
		// If health table doesn't exist, try a simple query to test API access
		_, _, err2 := s.postgrest.From("information_schema.tables").Select("table_name", "", false).Limit(1, "").Execute()
		if err2 != nil {
			// If that fails, try a simple HTTP request to test API access
			req, err3 := http.NewRequestWithContext(ctx, "GET", s.url+"/rest/v1/", nil)
			if err3 != nil {
				return fmt.Errorf("ping failed - unable to create request: %w", err3)
			}
			req.Header.Set("apikey", s.apiKey)
			req.Header.Set("Authorization", "Bearer "+s.apiKey)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err4 := client.Do(req)
			if err4 != nil {
				return fmt.Errorf("ping failed - unable to access Supabase API: %w", err4)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				return fmt.Errorf("ping failed - Supabase API returned status %d", resp.StatusCode)
			}
		}
	}
	return nil
}

// GetClient returns the underlying Supabase client
func (s *SupabaseClient) GetClient() *supa.Client {
	return s.client
}

// GetPostgrestClient returns the PostgREST client for direct database operations
func (s *SupabaseClient) GetPostgrestClient() *postgrest.Client {
	return s.postgrest
}

// GetConnectionString returns the Supabase connection string
func (s *SupabaseClient) GetConnectionString() string {
	return fmt.Sprintf("supabase://%s", s.url)
}

// IsConnected returns true if connected to Supabase
func (s *SupabaseClient) IsConnected() bool {
	return s.client != nil && s.postgrest != nil
}

// GetURL returns the Supabase project URL
func (s *SupabaseClient) GetURL() string {
	return s.url
}

// GetAPIKey returns the API key (for logging/debugging purposes only)
func (s *SupabaseClient) GetAPIKey() string {
	return s.apiKey
}

// GetServiceKey returns the service role key
func (s *SupabaseClient) GetServiceKey() string {
	return s.serviceKey
}
