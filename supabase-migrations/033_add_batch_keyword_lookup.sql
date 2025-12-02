-- Migration: Add batch keyword lookup function for performance optimization
-- This function enables efficient batch queries instead of N individual queries
-- Created: 2025-01-XX
-- Purpose: Support Phase 2.2 - Query batching for keyword lookups

-- Function to batch find keywords with industry matches
-- Returns all keyword matches in a single query instead of N queries
CREATE OR REPLACE FUNCTION batch_find_keywords(
    p_keywords text[]
)
RETURNS TABLE (
    keyword text,
    industry_id int,
    industry_name text,
    base_weight float8,
    similarity_score float8
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        k.keyword as keyword,
        kw.industry_id,
        i.name as industry_name,
        kw.base_weight,
        GREATEST(
            CASE WHEN kw.keyword = k.keyword THEN 1.0 ELSE 0.0 END,
            similarity(kw.keyword, k.keyword)
        ) as similarity_score
    FROM unnest(p_keywords) AS k(keyword)
    JOIN keyword_weights kw ON 
        kw.keyword = k.keyword 
        OR similarity(kw.keyword, k.keyword) > 0.3
    JOIN industries i ON i.id = kw.industry_id
    WHERE kw.is_active = true
        AND i.is_active = true
    ORDER BY k.keyword, similarity_score DESC;
END;
$$;

COMMENT ON FUNCTION batch_find_keywords IS 
    'Batch lookup for multiple keywords in a single query. Returns all matches with similarity scores.';

-- Function to batch find industry topics for multiple keywords
CREATE OR REPLACE FUNCTION batch_find_industry_topics(
    p_keywords text[]
)
RETURNS TABLE (
    keyword text,
    industry_id int,
    industry_name text,
    relevance_score float8,
    accuracy_score float8
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        k.keyword as keyword,
        it.industry_id,
        i.name as industry_name,
        it.relevance_score,
        it.accuracy_score
    FROM unnest(p_keywords) AS k(keyword)
    JOIN industry_topics it ON 
        it.topic ILIKE '%' || k.keyword || '%'
    JOIN industries i ON i.id = it.industry_id
    WHERE i.is_active = true
    ORDER BY k.keyword, it.relevance_score DESC;
END;
$$;

COMMENT ON FUNCTION batch_find_industry_topics IS 
    'Batch lookup for industry topics matching multiple keywords in a single query.';

