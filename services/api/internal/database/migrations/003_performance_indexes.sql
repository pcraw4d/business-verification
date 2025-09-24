-- Performance indexes and extensions
-- Migration: 003_performance_indexes.sql

-- Enable pg_trgm for fast ILIKE searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Businesses: accelerate ILIKE searches on name, legal_name, and registration_number
CREATE INDEX IF NOT EXISTS idx_businesses_name_trgm ON businesses USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_businesses_legal_name_trgm ON businesses USING gin (legal_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_businesses_reg_number_trgm ON businesses USING gin (registration_number gin_trgm_ops);

-- Business classifications: frequent lookups by business and recent-first ordering
CREATE INDEX IF NOT EXISTS idx_business_classifications_business_id_created_at
    ON business_classifications (business_id, created_at DESC);

-- Users: accelerate case-insensitive email lookups (if queries use ILIKE)
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);

-- API keys: accelerate active keys by role (common admin queries)
CREATE INDEX IF NOT EXISTS idx_api_keys_role_status ON api_keys (role, status);


