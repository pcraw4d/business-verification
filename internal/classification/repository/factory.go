package repository

import (
	"log"

	"github.com/pcraw4d/business-verification/internal/database"
)

// NewRepository creates a new keyword repository with the real Supabase client
func NewRepository(supabaseClient *database.SupabaseClient, logger *log.Logger) KeywordRepository {
	adapter := NewSupabaseClientAdapter(supabaseClient)
	return NewSupabaseKeywordRepository(adapter, logger)
}

// NewRepositoryWithDefaultLogger creates a new keyword repository with default logger
func NewRepositoryWithDefaultLogger(supabaseClient *database.SupabaseClient) KeywordRepository {
	adapter := NewSupabaseClientAdapter(supabaseClient)
	return NewSupabaseKeywordRepository(adapter, log.Default())
}

// Note: Mock repository creation is handled in the test files
// Use NewRepository() with a real Supabase client for production
