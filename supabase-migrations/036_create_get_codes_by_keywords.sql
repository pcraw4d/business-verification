-- Migration: Create get_codes_by_keywords function
-- This migration creates the database function for keyword-based code matching
-- Created: 2025-12-22
-- Purpose: Fix keyword matching (GetCodesByKeywords RPC function missing)

-- Function to get classification codes by keywords from code_keywords table
-- This function is used by GetCodesByKeywords in supabase_repository.go
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
        ck.code,
        cc.description,
        MAX(ck.weight) as max_weight
    FROM code_keywords ck
    JOIN classification_codes cc ON cc.code = ck.code AND cc.code_type = ck.code_type
    WHERE ck.code_type = p_code_type
        AND ck.keyword = ANY(p_keywords)
        AND cc.is_active = true
    GROUP BY ck.code, cc.description
    ORDER BY max_weight DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_codes_by_keywords IS 
    'Returns classification codes (MCC, SIC, NAICS) matching keywords from code_keywords table. 
     Used by GetCodesByKeywords for keyword-based code matching.';

-- Ensure indexes exist for performance
CREATE INDEX IF NOT EXISTS idx_code_keywords_code_type_keyword 
    ON code_keywords(code_type, keyword);

COMMENT ON INDEX idx_code_keywords_code_type_keyword IS 
    'Index on code_keywords(code_type, keyword) for fast keyword matching in get_codes_by_keywords function';

