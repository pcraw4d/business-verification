-- Migration: Create get_codes_by_keywords function
-- This migration creates the database function for keyword-based code matching
-- Created: 2025-12-22
-- Purpose: Fix keyword matching (GetCodesByKeywords RPC function missing)

-- Function to get classification codes by keywords from code_keywords table
-- This function is used by GetCodesByKeywords in supabase_repository.go
-- Note: code_keywords table has code_id (FK to classification_codes.id), not code/code_type
-- Note: code_keywords has relevance_score, not weight
CREATE OR REPLACE FUNCTION get_codes_by_keywords(
    p_code_type text,
    p_keywords text[],
    p_limit int DEFAULT 10
)
RETURNS TABLE (
    code text,
    description text,
    max_weight float8
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        cc.code::text,  -- Cast VARCHAR(20) to text to match function return type
        cc.description,
        MAX(ck.relevance_score)::double precision as max_weight  -- Cast DECIMAL(3,2) to double precision
    FROM code_keywords ck
    JOIN classification_codes cc ON cc.id = ck.code_id
    WHERE cc.code_type = p_code_type
        AND ck.keyword = ANY(p_keywords)
        AND cc.is_active = true
    GROUP BY cc.code, cc.description
    ORDER BY max_weight DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_codes_by_keywords IS 
    'Returns classification codes (MCC, SIC, NAICS) matching keywords from code_keywords table. 
     Used by GetCodesByKeywords for keyword-based code matching.';

-- Ensure indexes exist for performance
-- Note: code_keywords doesn't have code_type column, so we index on keyword only
-- The code_type filter is done via JOIN with classification_codes
CREATE INDEX IF NOT EXISTS idx_code_keywords_keyword_lookup 
    ON code_keywords(keyword) 
    WHERE keyword IS NOT NULL;

COMMENT ON INDEX idx_code_keywords_keyword_lookup IS 
    'Index on code_keywords(keyword) for fast keyword matching in get_codes_by_keywords function';

