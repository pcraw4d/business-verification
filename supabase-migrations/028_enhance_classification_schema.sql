-- =====================================================
-- Migration: Enhance Classification Schema
-- Purpose: Add missing columns, indexes, and full-text search capabilities
-- Date: 2025-01-XX
-- =====================================================

-- Step 1: Ensure is_active column exists in classification_codes table
-- (This was added in migration 026, but we verify it exists here)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'classification_codes' 
        AND column_name = 'is_active'
    ) THEN
        ALTER TABLE classification_codes ADD COLUMN is_active BOOLEAN DEFAULT true;
        RAISE NOTICE 'Added is_active column to classification_codes';
    ELSE
        RAISE NOTICE 'is_active column already exists in classification_codes';
    END IF;
END $$;

-- Step 2: Ensure is_active index exists
CREATE INDEX IF NOT EXISTS idx_classification_codes_active 
    ON classification_codes(is_active) 
    WHERE is_active = true; -- Partial index for better performance

-- Step 3: Enable pg_trgm extension for trigram indexes (if not already enabled)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Step 4: Add trigram index on description for fuzzy matching
-- This enables fast similarity searches on code descriptions
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_trgm 
    ON classification_codes USING gin(description gin_trgm_ops);

-- Step 5: Add composite indexes for common query patterns

-- Index for filtering by code_type and is_active (most common query)
CREATE INDEX IF NOT EXISTS idx_classification_codes_type_active 
    ON classification_codes(code_type, is_active) 
    WHERE is_active = true;

-- Index for filtering by industry_id and code_type
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_type 
    ON classification_codes(industry_id, code_type);

-- Index for filtering by industry_id, code_type, and is_active
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_type_active 
    ON classification_codes(industry_id, code_type, is_active) 
    WHERE is_active = true;

-- Index for filtering by is_primary and is_active
CREATE INDEX IF NOT EXISTS idx_classification_codes_primary_active 
    ON classification_codes(is_primary, is_active) 
    WHERE is_primary = true AND is_active = true;

-- Step 6: Add full-text search index on description
-- This enables full-text search capabilities using to_tsvector
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_fts 
    ON classification_codes USING gin(to_tsvector('english', description));

-- Step 7: Add index on code for faster lookups (if not exists)
-- This should already exist from migration 001, but we ensure it exists
CREATE INDEX IF NOT EXISTS idx_classification_codes_code 
    ON classification_codes(code);

-- Step 8: Add composite unique index for (code_type, code, is_active) to prevent duplicates
-- Note: The existing UNIQUE constraint on (code_type, code) already prevents duplicates
-- This is just for documentation/clarity

-- Step 9: Add index on updated_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_classification_codes_updated_at 
    ON classification_codes(updated_at DESC);

-- Step 10: Add index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_classification_codes_created_at 
    ON classification_codes(created_at DESC);

-- Step 11: Verify all indexes were created successfully
DO $$
DECLARE
    index_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes
    WHERE tablename = 'classification_codes'
    AND schemaname = 'public';
    
    RAISE NOTICE 'Total indexes on classification_codes: %', index_count;
END $$;

-- =====================================================
-- Comments for documentation
-- =====================================================

COMMENT ON INDEX idx_classification_codes_active IS 
    'Partial index on is_active=true for faster filtering of active codes';

COMMENT ON INDEX idx_classification_codes_description_trgm IS 
    'Trigram index for fuzzy matching and similarity searches on code descriptions';

COMMENT ON INDEX idx_classification_codes_type_active IS 
    'Composite index for filtering by code_type and is_active (most common query pattern)';

COMMENT ON INDEX idx_classification_codes_industry_type IS 
    'Composite index for filtering by industry_id and code_type';

COMMENT ON INDEX idx_classification_codes_industry_type_active IS 
    'Composite index for filtering by industry_id, code_type, and is_active';

COMMENT ON INDEX idx_classification_codes_primary_active IS 
    'Partial index for filtering primary codes that are active';

COMMENT ON INDEX idx_classification_codes_description_fts IS 
    'Full-text search index on code descriptions using English language';

COMMENT ON INDEX idx_classification_codes_updated_at IS 
    'Index on updated_at for time-based queries and sorting';

COMMENT ON INDEX idx_classification_codes_created_at IS 
    'Index on created_at for time-based queries and sorting';

-- =====================================================
-- Verification queries (for manual checking)
-- =====================================================

-- To verify indexes exist:
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'classification_codes';

-- To verify is_active column exists:
-- SELECT column_name, data_type, column_default 
-- FROM information_schema.columns 
-- WHERE table_name = 'classification_codes' AND column_name = 'is_active';

-- To check index sizes:
-- SELECT 
--     schemaname,
--     tablename,
--     indexname,
--     pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
-- FROM pg_indexes
-- WHERE tablename = 'classification_codes'
-- ORDER BY pg_relation_size(indexrelid) DESC;

