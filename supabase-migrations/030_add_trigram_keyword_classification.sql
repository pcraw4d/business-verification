-- Migration: Add trigram-based keyword classification function
-- This function enables fuzzy keyword matching using PostgreSQL's pg_trgm extension
-- Created: 2025-01-XX
-- Purpose: Support Phase 1.1 - Enhanced keyword strategy with trigram similarity

-- Ensure pg_trgm extension is enabled (should already be enabled from migration 028)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Function to classify business by keywords using trigram similarity
-- This function leverages trigram indexes for fast fuzzy matching
CREATE OR REPLACE FUNCTION classify_business_by_keywords_trigram(
    p_keywords text[],
    p_business_name text DEFAULT '',
    p_similarity_threshold float DEFAULT 0.3
)
RETURNS TABLE (
    industry_id int,
    industry_name text,
    score float,
    match_count int,
    matched_keywords text[]
) 
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.id as industry_id,
        i.name as industry_name,
        SUM(kw.base_weight * GREATEST(
            similarity(kw.keyword, k.keyword),
            CASE WHEN kw.keyword = k.keyword THEN 1.0 ELSE 0.0 END
        )) as score,
        COUNT(DISTINCT kw.keyword)::int as match_count,
        array_agg(DISTINCT kw.keyword) as matched_keywords
    FROM industries i
    JOIN keyword_weights kw ON kw.industry_id = i.id
    CROSS JOIN (SELECT unnest(p_keywords) as keyword) k
    WHERE 
        -- Use trigram similarity for fuzzy matching
        similarity(kw.keyword, k.keyword) > p_similarity_threshold
        -- OR exact match
        OR kw.keyword = ANY(p_keywords)
        -- Only active keywords
        AND kw.is_active = true
    GROUP BY i.id, i.name
    HAVING COUNT(DISTINCT kw.keyword) >= 1
    ORDER BY score DESC, match_count DESC
    LIMIT 10;
END;
$$;

-- Add comment
COMMENT ON FUNCTION classify_business_by_keywords_trigram IS 
    'Classifies business by keywords using trigram similarity for fuzzy matching. 
     Leverages pg_trgm extension and trigram indexes for fast performance.';

-- Create index on keyword_weights.keyword for trigram similarity if not exists
-- This index will be used by the similarity() function
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword_trgm 
    ON keyword_weights USING gin (keyword gin_trgm_ops);

COMMENT ON INDEX idx_keyword_weights_keyword_trgm IS 
    'Trigram index on keyword_weights.keyword for fast fuzzy matching with similarity() function';

