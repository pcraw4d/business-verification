package repository

import (
	"log"

	"github.com/pcraw4d/business-verification/internal/database"
)

// NewRepository creates a new keyword repository with the real Supabase client
func NewRepository(supabaseClient *database.SupabaseClient, logger *log.Logger) KeywordRepository {
	if supabaseClient == nil {
		logger.Printf("⚠️ Supabase client is nil - classification will fail")
		return nil
	}
	return NewSupabaseKeywordRepository(supabaseClient, logger)
}

// NewRepositoryWithDefaultLogger creates a new keyword repository with default logger
func NewRepositoryWithDefaultLogger(supabaseClient *database.SupabaseClient) KeywordRepository {
	return NewSupabaseKeywordRepository(supabaseClient, log.Default())
}

// Note: Mock repository creation is handled in the test files
// Use NewRepository() with a real Supabase client for production
