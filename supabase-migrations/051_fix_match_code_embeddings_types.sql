-- Migration: Fix type mismatch in match_code_embeddings function
-- The function returns text but table columns are VARCHAR, causing type errors

-- Fix the main function
CREATE OR REPLACE FUNCTION match_code_embeddings(
    query_embedding vector(384),
    code_type_filter text,
    match_threshold float DEFAULT 0.7,
    match_count int DEFAULT 5
)
RETURNS TABLE (
    code text,
    code_type text,
    description text,
    extended_description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ce.code::text,
        ce.code_type::text,
        ce.description::text,
        ce.extended_description::text,
        1 - (ce.embedding <=> query_embedding) as similarity
    FROM code_embeddings ce
    WHERE ce.code_type = code_type_filter
        AND 1 - (ce.embedding <=> query_embedding) > match_threshold
    ORDER BY ce.embedding <=> query_embedding
    LIMIT match_count;
END;
$$;

