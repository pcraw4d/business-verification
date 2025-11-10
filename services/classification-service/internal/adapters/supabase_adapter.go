package adapters

import (
	"log"

	"kyb-platform/internal/database"
	"kyb-platform/services/classification-service/internal/config"
)

// CreateDatabaseClient creates a database.SupabaseClient from classification service config
// This adapter allows the classification service to use the internal classification repository
func CreateDatabaseClient(cfg *config.SupabaseConfig, logger *log.Logger) (*database.SupabaseClient, error) {
	dbConfig := &database.SupabaseConfig{
		URL:            cfg.URL,
		APIKey:         cfg.APIKey,
		ServiceRoleKey: cfg.ServiceRoleKey,
		JWTSecret:      cfg.JWTSecret,
	}

	return database.NewSupabaseClient(dbConfig, logger)
}

