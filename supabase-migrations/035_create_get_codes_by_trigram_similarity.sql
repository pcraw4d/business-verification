-- Migration: Create get_codes_by_trigram_similarity function
-- This migration creates the database function for trigram-based code similarity matching
-- Created: 2025-12-21
-- Purpose: Fix NAICS/SIC code generation (Track 4.2)

-- Ensure pg_trgm extension is enabled
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Function to get classification codes by trigram similarity to industry name
-- This function is used by GetCodesByTrigramSimilarity in supabase_repository.go
CREATE OR REPLACE FUNCTION get_codes_by_trigram_similarity(
    p_code_type text,
    p_industry_name text,
    p_threshold float DEFAULT 0.3,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    code text,
    description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cc.code,
        cc.description,
        similarity(cc.description, p_industry_name) as similarity
    FROM classification_codes cc
    WHERE 
        cc.code_type = p_code_type
        AND cc.is_active = true
        AND similarity(cc.description, p_industry_name) >= p_threshold
    ORDER BY 
        similarity DESC,
        cc.code ASC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_codes_by_trigram_similarity IS 
    'Returns classification codes (MCC, SIC, NAICS) with similarity scores using trigram matching against industry name. 
     Used for fuzzy matching when direct industry lookup fails.';

-- Create trigram index on classification_codes.description if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_trgm 
    ON classification_codes USING gin (description gin_trgm_ops);

COMMENT ON INDEX idx_classification_codes_description_trgm IS 
    'Trigram index on classification_codes.description for fast similarity matching in get_codes_by_trigram_similarity function';

