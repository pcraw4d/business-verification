package database

import (
	"context"
	"fmt"
	"log"

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
	postgrestClient := postgrest.NewClient(cfg.URL+"/rest/v1", cfg.APIKey, nil)

	if logger == nil {
		logger = log.Default()
	}

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
	// Test connection with a simple query to check if tables exist
	// This will help verify the database schema is accessible
	_, _, err := s.postgrest.From("industries").Select("*", "", false).Execute()
	if err != nil {
		return fmt.Errorf("ping failed - database schema may not be initialized: %w", err)
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
