package database

import (
	"context"
	"fmt"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
	supa "github.com/supabase-community/supabase-go"
)

// SupabaseClient represents a Supabase database client
type SupabaseClient struct {
	client *supa.Client
	url    string
	key    string
	logger *observability.Logger
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(cfg *config.SupabaseConfig, logger *observability.Logger) (*SupabaseClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("SUPABASE_ANON_KEY is required")
	}
	if cfg.ServiceRoleKey == "" {
		return nil, fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is required")
	}

	client, err := supa.NewClient(cfg.URL, cfg.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseClient{
		client: client,
		url:    cfg.URL,
		key:    cfg.APIKey,
		logger: logger,
	}, nil
}

// Connect establishes a connection to Supabase
func (s *SupabaseClient) Connect(ctx context.Context) error {
	s.logger.Info("Connecting to Supabase", "url", s.url)

	// Test the connection by making a simple query
	_, err := s.client.DB.From("business_classifications").Select("count", false).Execute("")
	if err != nil {
		s.logger.Error("Failed to connect to Supabase", "error", err)
		return fmt.Errorf("failed to connect to Supabase: %w", err)
	}

	s.logger.Info("Successfully connected to Supabase")
	return nil
}

// Close closes the Supabase connection
func (s *SupabaseClient) Close() error {
	s.logger.Info("Closing Supabase connection")
	// Supabase client doesn't require explicit closing
	return nil
}

// Ping checks the connection to Supabase
func (s *SupabaseClient) Ping(ctx context.Context) error {
	// Test connection with a simple query
	_, err := s.client.DB.From("business_classifications").Select("count", false).Execute("")
	return err
}

// ExecuteQuery executes a query on Supabase
func (s *SupabaseClient) ExecuteQuery(ctx context.Context, query string, args ...interface{}) error {
	s.logger.Debug("Executing Supabase query", "query", query)

	// For Supabase, we use the REST API rather than raw SQL
	// This is a simplified implementation - in practice, you'd use the PostgREST client
	_, err := s.client.DB.Rpc("execute_sql", map[string]interface{}{
		"query": query,
		"args":  args,
	}).Execute("")

	return err
}

// GetConnectionString returns the Supabase connection string
func (s *SupabaseClient) GetConnectionString() string {
	// For Supabase, we use the REST API rather than direct PostgreSQL connection
	return fmt.Sprintf("supabase://%s", s.url)
}

// IsConnected returns true if connected to Supabase
func (s *SupabaseClient) IsConnected() bool {
	return s.client != nil
}

// GetClientInfo returns information about the Supabase client
func (s *SupabaseClient) GetClientInfo() map[string]interface{} {
	return map[string]interface{}{
		"provider":  "supabase",
		"url":       s.url,
		"connected": s.IsConnected(),
	}
}

// BeginTx starts a new transaction (placeholder for Supabase)
func (s *SupabaseClient) BeginTx(ctx context.Context) (interface{}, error) {
	s.logger.Debug("BeginTx called on Supabase client")
	// Supabase handles transactions differently via REST API
	// For now, return a mock transaction
	return &SupabaseTransaction{client: s.client}, nil
}

// Commit commits a transaction (placeholder for Supabase)
func (s *SupabaseClient) Commit(tx interface{}) error {
	s.logger.Debug("Commit called on Supabase client")
	return nil
}

// Rollback rolls back a transaction (placeholder for Supabase)
func (s *SupabaseClient) Rollback(tx interface{}) error {
	s.logger.Debug("Rollback called on Supabase client")
	return nil
}

// SaveClassification saves a business classification to Supabase
func (s *SupabaseClient) SaveClassification(ctx context.Context, classification map[string]interface{}) error {
	s.logger.Debug("Saving classification to Supabase", "business_id", classification["business_id"])

	_, err := s.client.DB.From("business_classifications").Insert(classification).Execute("")
	if err != nil {
		s.logger.Error("Failed to save classification to Supabase", "error", err)
		return fmt.Errorf("failed to save classification: %w", err)
	}

	s.logger.Info("Successfully saved classification to Supabase", "business_id", classification["business_id"])
	return nil
}

// GetClassification retrieves a business classification from Supabase
func (s *SupabaseClient) GetClassification(ctx context.Context, businessID string) (map[string]interface{}, error) {
	s.logger.Debug("Getting classification from Supabase", "business_id", businessID)

	result, err := s.client.DB.From("business_classifications").
		Select("*").
		Eq("business_id", businessID).
		Single().
		Execute("")

	if err != nil {
		s.logger.Error("Failed to get classification from Supabase", "error", err)
		return nil, fmt.Errorf("failed to get classification: %w", err)
	}

	var classification map[string]interface{}
	if err := result.Unmarshal(&classification); err != nil {
		return nil, fmt.Errorf("failed to unmarshal classification: %w", err)
	}

	return classification, nil
}

// GetClassifications retrieves all classifications for a user from Supabase
func (s *SupabaseClient) GetClassifications(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	s.logger.Debug("Getting classifications from Supabase", "user_id", userID)

	result, err := s.client.DB.From("business_classifications").
		Select("*").
		Eq("user_id", userID).
		Order("created_at", &map[string]string{"ascending": "false"}).
		Execute("")

	if err != nil {
		s.logger.Error("Failed to get classifications from Supabase", "error", err)
		return nil, fmt.Errorf("failed to get classifications: %w", err)
	}

	var classifications []map[string]interface{}
	if err := result.Unmarshal(&classifications); err != nil {
		return nil, fmt.Errorf("failed to unmarshal classifications: %w", err)
	}

	return classifications, nil
}

// SaveUser saves a user to Supabase
func (s *SupabaseClient) SaveUser(ctx context.Context, user map[string]interface{}) error {
	s.logger.Debug("Saving user to Supabase", "email", user["email"])

	_, err := s.client.DB.From("users").Insert(user).Execute("")
	if err != nil {
		s.logger.Error("Failed to save user to Supabase", "error", err)
		return fmt.Errorf("failed to save user: %w", err)
	}

	s.logger.Info("Successfully saved user to Supabase", "email", user["email"])
	return nil
}

// GetUser retrieves a user from Supabase
func (s *SupabaseClient) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	s.logger.Debug("Getting user from Supabase", "user_id", userID)

	result, err := s.client.DB.From("users").
		Select("*").
		Eq("id", userID).
		Single().
		Execute("")

	if err != nil {
		s.logger.Error("Failed to get user from Supabase", "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user map[string]interface{}
	if err := result.Unmarshal(&user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

// SaveAuditLog saves an audit log entry to Supabase
func (s *SupabaseClient) SaveAuditLog(ctx context.Context, auditLog map[string]interface{}) error {
	s.logger.Debug("Saving audit log to Supabase", "event_type", auditLog["event_type"])

	_, err := s.client.DB.From("audit_logs").Insert(auditLog).Execute("")
	if err != nil {
		s.logger.Error("Failed to save audit log to Supabase", "error", err)
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs from Supabase
func (s *SupabaseClient) GetAuditLogs(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	s.logger.Debug("Getting audit logs from Supabase", "filters", filters)

	query := s.client.DB.From("audit_logs").Select("*")

	// Apply filters
	for key, value := range filters {
		query = query.Eq(key, value)
	}

	result, err := query.Order("created_at", &map[string]string{"ascending": "false"}).Execute("")
	if err != nil {
		s.logger.Error("Failed to get audit logs from Supabase", "error", err)
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	var auditLogs []map[string]interface{}
	if err := result.Unmarshal(&auditLogs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit logs: %w", err)
	}

	return auditLogs, nil
}

// SupabaseTransaction represents a Supabase transaction
type SupabaseTransaction struct {
	client *supa.Client
}

// Execute executes a query within the transaction
func (tx *SupabaseTransaction) Execute(query string, args ...interface{}) error {
	// In a real implementation, you would use Supabase's transaction support
	// For now, we'll just execute the query normally
	_, err := tx.client.DB.Rpc("execute_sql", map[string]interface{}{
		"query": query,
		"args":  args,
	}).Execute("")
	return err
}
