package database

import (
	"context"
	"fmt"

	"github.com/pcraw4d/business-verification/internal/config"
)

// NewDatabase creates a new database instance based on configuration
func NewDatabase(cfg *config.DatabaseConfig) (Database, error) {
	dbConfig := NewDatabaseConfig(cfg)

	switch cfg.Driver {
	case "postgres":
		return NewPostgresDB(dbConfig), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// NewDatabaseWithConnection creates a new database instance and connects to it
func NewDatabaseWithConnection(ctx context.Context, cfg *config.DatabaseConfig) (Database, error) {
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	if err := db.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
