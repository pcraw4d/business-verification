-- Migration: Optimize classification queries for Phase 2
-- This migration adds indexes and materialized views to improve query performance
-- Target: <50ms average query time for keyword and trigram lookups

-- 1. Add composite indexes for common query patterns
-- Note: CONCURRENTLY removed to allow running inside transaction block
-- For production with large tables, consider running these with CONCURRENTLY in a maintenance window
-- Note: code_keywords table has: code_id, keyword, relevance_score (not code_type, code, weight)
-- We need to join with classification_codes to get code_type and code

-- Index on code_keywords for keyword lookups (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_code_keywords_keyword_relevance
ON code_keywords (keyword, relevance_score DESC);

-- Index on code_keywords for code_id lookups (for joins)
CREATE INDEX IF NOT EXISTS idx_code_keywords_code_id
ON code_keywords (code_id);

-- Index on classification_codes for type and description (for trigram searches)
CREATE INDEX IF NOT EXISTS idx_codes_type_description
ON classification_codes (code_type, description);

-- 2. Add index for crosswalk queries (if industry_code_crosswalks table exists)
-- Note: Crosswalks may be in code_metadata table instead
-- This index is for the code_metadata crosswalk_data JSONB field
CREATE INDEX IF NOT EXISTS idx_code_metadata_code_type_code
ON code_metadata (code_type, code)
WHERE is_active = true;

-- 3. Optimize trigram queries with GIN index (if pg_trgm extension is enabled)
-- Check if extension exists before creating index
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        CREATE INDEX IF NOT EXISTS idx_codes_description_trgm
        ON classification_codes USING gin (description gin_trgm_ops);
    ELSE
        RAISE NOTICE 'pg_trgm extension not found, skipping trigram index';
    END IF;
END $$;

-- 4. Composite index for keyword queries with code_type filter
-- This supports queries that filter by code_type and keyword
CREATE INDEX IF NOT EXISTS idx_codes_type_code_active
ON classification_codes (code_type, code)
WHERE is_active = true;

-- 5. Add index for classification_codes lookups by code and type
CREATE INDEX IF NOT EXISTS idx_classification_codes_code_type_code
ON classification_codes (code_type, code)
WHERE is_active = true;

-- 6. Refresh statistics for query planner
ANALYZE code_keywords;
ANALYZE classification_codes;
ANALYZE code_metadata;

-- 7. Create materialized view for frequently accessed code metadata
-- This pre-computes code-keyword relationships for faster lookups
-- Note: code_keywords uses code_id (FK to classification_codes.id), not code/code_type
CREATE MATERIALIZED VIEW IF NOT EXISTS code_search_cache AS
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    array_agg(DISTINCT ck.keyword) FILTER (WHERE ck.keyword IS NOT NULL) as keywords,
    MAX(ck.relevance_score) as max_keyword_weight
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code_id = cc.id
WHERE cc.is_active = true
GROUP BY cc.code_type, cc.code, cc.description;

-- 8. Create indexes on materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_code_search_cache_unique
ON code_search_cache (code_type, code);

-- GIN index for keyword array searches
CREATE INDEX IF NOT EXISTS idx_code_search_cache_keywords
ON code_search_cache USING gin (keywords);

-- Index for code type lookups
CREATE INDEX IF NOT EXISTS idx_code_search_cache_code_type
ON code_search_cache (code_type);

-- 9. Create function to refresh materialized view (can be called periodically)
-- Note: CONCURRENTLY removed from function to allow running inside transaction
-- For production, you can manually refresh with CONCURRENTLY outside transactions
CREATE OR REPLACE FUNCTION refresh_code_search_cache()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW code_search_cache;
END;
$$;

-- 10. Initial refresh of materialized view
REFRESH MATERIALIZED VIEW code_search_cache;

-- 11. Add comment documenting the migration
COMMENT ON MATERIALIZED VIEW code_search_cache IS 
'Phase 2: Materialized view for fast code-keyword lookups. Refresh periodically with refresh_code_search_cache() function.';

COMMENT ON INDEX idx_code_keywords_keyword_relevance IS 
'Phase 2: Composite index for code_keywords queries by keyword and relevance_score';

COMMENT ON INDEX idx_codes_description_trgm IS 
'Phase 2: GIN trigram index for fuzzy matching on code descriptions';
