package datasource

import (
	"context"

	"kyb-platform/internal/database"
)

// DBSource enriches data using our primary application database
type DBSource struct {
	db database.Database
}

func NewDBSource(db database.Database) *DBSource {
	return &DBSource{db: db}
}

func (s *DBSource) Name() string { return "db_source" }

func (s *DBSource) HealthCheck(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DBSource) Enrich(ctx context.Context, req EnrichmentRequest) (EnrichmentResult, error) {
	// Try by registration number first
	if req.RegistrationNumber != "" {
		if b, err := s.db.GetBusinessByRegistrationNumber(ctx, req.RegistrationNumber); err == nil && b != nil {
			return mapBusinessToEnrichment(b), nil
		}
	}
	// Fallback by name
	if req.BusinessName != "" {
		// Basic search and pick the first match
		if list, err := s.db.SearchBusinesses(ctx, req.BusinessName, 1, 0); err == nil && len(list) > 0 {
			return mapBusinessToEnrichment(list[0]), nil
		}
	}
	return EnrichmentResult{}, nil
}

func mapBusinessToEnrichment(b *database.Business) EnrichmentResult {
	res := EnrichmentResult{
		CleanBusinessName: b.LegalName,
		Industry:          b.Industry,
		Description:       "",
	}
	// Derive simple keywords from known fields
	keys := []string{}
	if b.Industry != "" {
		keys = append(keys, b.Industry)
	}
	if b.BusinessType != "" {
		keys = append(keys, b.BusinessType)
	}
	res.Keywords = keys
	return res
}
