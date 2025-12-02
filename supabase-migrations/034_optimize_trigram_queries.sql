-- Migration: Optimize queries to leverage trigram indexes
-- This migration creates database functions that use trigram similarity for better performance
-- Created: 2025-01-XX
-- Purpose: Support Phase 4.1 - Database Optimization with trigram indexes

-- Ensure pg_trgm extension is enabled
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Function to search keywords using trigram similarity
-- This replaces ILIKE queries with trigram-based similarity for better performance
CREATE OR REPLACE FUNCTION search_keywords_trigram(
    p_query text,
    p_limit int DEFAULT 50,
    p_similarity_threshold float DEFAULT 0.3
)
RETURNS TABLE (
    id int,
    industry_id int,
    keyword text,
    weight float8,
    is_active boolean,
    similarity_score float8
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ik.id,
        ik.industry_id,
        ik.keyword,
        ik.weight,
        ik.is_active,
        GREATEST(
            similarity(ik.keyword, p_query),
            CASE WHEN ik.keyword ILIKE ('%' || p_query || '%') THEN 0.5 ELSE 0 END
        ) AS similarity_score
    FROM industry_keywords ik
    WHERE 
        ik.is_active = true
        AND (
            -- Use trigram similarity for fuzzy matching
            similarity(ik.keyword, p_query) > p_similarity_threshold
            -- OR ILIKE for substring matching (fallback)
            OR ik.keyword ILIKE ('%' || p_query || '%')
        )
    ORDER BY 
        similarity_score DESC,
        ik.weight DESC,
        ik.keyword ASC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION search_keywords_trigram IS 
    'Searches keywords using trigram similarity for fast fuzzy matching. 
     Leverages trigram indexes for better performance than ILIKE queries.';

-- Function to find classification codes by keywords using trigram similarity
-- This optimizes the code_keywords matching process
CREATE OR REPLACE FUNCTION find_codes_by_keywords_trigram(
    p_keywords text[],
    p_code_type text,
    p_min_relevance float DEFAULT 0.5,
    p_similarity_threshold float DEFAULT 0.3,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    code text,
    code_type text,
    description text,
    industry_id int,
    relevance_score float8,
    match_type text,
    similarity_score float8
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    WITH keyword_matches AS (
        SELECT DISTINCT
            ck.code_id,
            ck.keyword,
            ck.relevance_score,
            ck.match_type,
            GREATEST(
                MAX(similarity(ck.keyword, k.keyword)),
                CASE WHEN ck.keyword = ANY(p_keywords) THEN 1.0 ELSE 0 END
            ) AS similarity_score
        FROM code_keywords ck
        CROSS JOIN (SELECT unnest(p_keywords) AS keyword) k
        WHERE 
            -- Use trigram similarity for fuzzy matching
            similarity(ck.keyword, k.keyword) > p_similarity_threshold
            -- OR exact match
            OR ck.keyword = ANY(p_keywords)
        GROUP BY ck.code_id, ck.keyword, ck.relevance_score, ck.match_type
        HAVING MAX(similarity(ck.keyword, k.keyword)) > p_similarity_threshold
            OR ck.keyword = ANY(p_keywords)
    ),
    code_scores AS (
        SELECT 
            cc.code,
            cc.code_type,
            cc.description,
            cc.industry_id,
            MAX(km.relevance_score * km.similarity_score) AS relevance_score,
            MAX(km.match_type) AS match_type,
            MAX(km.similarity_score) AS similarity_score
        FROM classification_codes cc
        JOIN keyword_matches km ON cc.id = km.code_id
        WHERE cc.code_type = p_code_type
        GROUP BY cc.code, cc.code_type, cc.description, cc.industry_id
        HAVING MAX(km.relevance_score * km.similarity_score) >= p_min_relevance
    )
    SELECT 
        cs.code,
        cs.code_type,
        cs.description,
        cs.industry_id,
        cs.relevance_score,
        cs.match_type,
        cs.similarity_score
    FROM code_scores cs
    ORDER BY 
        cs.relevance_score DESC,
        cs.similarity_score DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION find_codes_by_keywords_trigram IS 
    'Finds classification codes by keywords using trigram similarity for fuzzy matching. 
     Optimizes code_keywords matching with trigram indexes for better performance.';

-- Create trigram indexes if they don't exist
-- Index on industry_keywords.keyword for search_keywords_trigram
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword_trgm 
    ON industry_keywords USING gin (keyword gin_trgm_ops);

COMMENT ON INDEX idx_industry_keywords_keyword_trgm IS 
    'Trigram index on industry_keywords.keyword for fast fuzzy matching in search_keywords_trigram function';

-- Index on code_keywords.keyword for find_codes_by_keywords_trigram
CREATE INDEX IF NOT EXISTS idx_code_keywords_keyword_trgm 
    ON code_keywords USING gin (keyword gin_trgm_ops);

COMMENT ON INDEX idx_code_keywords_keyword_trgm IS 
    'Trigram index on code_keywords.keyword for fast fuzzy matching in find_codes_by_keywords_trigram function';

-- Note: idx_keyword_weights_keyword_trgm was already created in migration 030
-- This migration adds indexes for industry_keywords and code_keywords tables

