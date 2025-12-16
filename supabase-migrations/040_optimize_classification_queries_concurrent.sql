-- Optional: Production-safe version with CONCURRENTLY
-- Run this script manually outside of a transaction for production deployments
-- This allows index creation without locking tables

-- 1. Add composite indexes for common query patterns (CONCURRENTLY)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_keywords_composite
ON code_keywords (code_type, keyword, weight DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_codes_type_description
ON classification_codes (code_type, description);

-- 2. Add index for crosswalk queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_metadata_code_type_code
ON code_metadata (code_type, code)
WHERE is_active = true;

-- 3. Optimize trigram queries with GIN index (if pg_trgm extension is enabled)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_codes_description_trgm
        ON classification_codes USING gin (description gin_trgm_ops);
    ELSE
        RAISE NOTICE 'pg_trgm extension not found, skipping trigram index';
    END IF;
END $$;

-- 4. Add covering index for keyword queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_keywords_covering
ON code_keywords (code_type, keyword) INCLUDE (code, weight);

-- 5. Add index for classification_codes lookups by code and type
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classification_codes_code_type_code
ON classification_codes (code_type, code)
WHERE is_active = true;

-- Update refresh function to use CONCURRENTLY (run after migration)
CREATE OR REPLACE FUNCTION refresh_code_search_cache()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY code_search_cache;
END;
$$;
