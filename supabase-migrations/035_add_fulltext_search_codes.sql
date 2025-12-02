-- Migration: Add full-text search for classification codes
-- This migration enables PostgreSQL full-text search for better semantic matching of code descriptions
-- Created: 2025-01-XX
-- Purpose: Support Phase 4.2 - Leverage full-text search for code matching

-- Create full-text search function for classification codes
CREATE OR REPLACE FUNCTION find_codes_by_fulltext_search(
    p_search_text text,
    p_code_type text,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    id int,
    industry_id int,
    code_type text,
    code text,
    description text,
    is_active boolean,
    relevance float8
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_tsquery tsquery;
    v_search_normalized text;
BEGIN
    -- Normalize search text: remove special characters, lowercase
    v_search_normalized := lower(trim(p_search_text));
    
    -- Use plainto_tsquery for better handling of user input
    -- This automatically handles multiple words, punctuation, and special characters
    -- It converts the search text to a tsquery that matches all words (AND logic)
    v_tsquery := plainto_tsquery('english', p_search_text);
    
    -- If plainto_tsquery returns empty (e.g., only stop words), fall back to simple matching
    IF v_tsquery IS NULL OR v_tsquery = ''::tsquery THEN
        -- Fallback: use ILIKE for simple substring matching
        RETURN QUERY
        SELECT 
            cc.id,
            cc.industry_id,
            cc.code_type,
            cc.code,
            cc.description,
            cc.is_active,
            0.5::float8 AS relevance
        FROM classification_codes cc
        WHERE 
            cc.code_type = p_code_type
            AND cc.is_active = true
            AND cc.description ILIKE '%' || p_search_text || '%'
        ORDER BY 
            cc.code ASC
        LIMIT p_limit;
        RETURN;
    END IF;
    
    RETURN QUERY
    SELECT 
        cc.id,
        cc.industry_id,
        cc.code_type,
        cc.code,
        cc.description,
        cc.is_active,
        ts_rank(
            to_tsvector('english', cc.description),
            v_tsquery
        ) AS relevance
    FROM classification_codes cc
    WHERE 
        cc.code_type = p_code_type
        AND cc.is_active = true
        AND to_tsvector('english', cc.description) @@ v_tsquery
    ORDER BY 
        relevance DESC,
        cc.code ASC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION find_codes_by_fulltext_search IS 
    'Finds classification codes using PostgreSQL full-text search for semantic matching of descriptions. 
     Uses ts_rank for relevance scoring and supports phrase matching.';

-- Create full-text search index on classification_codes.description
-- This GIN index will significantly speed up full-text search queries
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_fts 
    ON classification_codes USING gin (to_tsvector('english', description))
    WHERE is_active = true;

COMMENT ON INDEX idx_classification_codes_description_fts IS 
    'Full-text search GIN index on classification_codes.description for fast semantic matching.';

-- Optional: Create a materialized column for tsvector if performance is critical
-- This pre-computes the tsvector and can be faster for large datasets
-- ALTER TABLE classification_codes ADD COLUMN IF NOT EXISTS description_tsvector tsvector;
-- CREATE INDEX IF NOT EXISTS idx_classification_codes_description_tsvector 
--     ON classification_codes USING gin (description_tsvector)
--     WHERE is_active = true;
-- 
-- -- Trigger to keep tsvector column updated
-- CREATE OR REPLACE FUNCTION update_classification_codes_tsvector()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     NEW.description_tsvector := to_tsvector('english', NEW.description);
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
-- 
-- CREATE TRIGGER classification_codes_tsvector_update
--     BEFORE INSERT OR UPDATE ON classification_codes
--     FOR EACH ROW
--     EXECUTE FUNCTION update_classification_codes_tsvector();
-- 
-- -- Update existing rows
-- UPDATE classification_codes 
-- SET description_tsvector = to_tsvector('english', description)
-- WHERE description_tsvector IS NULL;

