-- Migration: Enable pgvector and create embeddings infrastructure
-- Phase 3: Add Layer 2 (Embeddings) for semantic similarity search
-- This migration enables vector similarity search for classification codes

-- Step 1: Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Step 2: Create code_embeddings table
CREATE TABLE code_embeddings (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL,
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    extended_description TEXT,
    industry_context TEXT,
    embedding vector(384), -- all-MiniLM-L6-v2 produces 384-dim vectors
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_code_embedding UNIQUE(code_type, code)
);

-- Step 3: Create indexes for fast similarity search
-- IVFFlat index for approximate nearest neighbor search
CREATE INDEX idx_code_embeddings_vector ON code_embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Additional indexes for filtering
CREATE INDEX idx_code_embeddings_type ON code_embeddings(code_type);
CREATE INDEX idx_code_embeddings_code ON code_embeddings(code);
CREATE INDEX idx_code_embeddings_updated ON code_embeddings(updated_at);

-- Step 4: Create function for similarity search
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

-- Step 5: Create function to search across all code types
CREATE OR REPLACE FUNCTION match_code_embeddings_all_types(
    query_embedding vector(384),
    match_threshold float DEFAULT 0.7,
    match_count_per_type int DEFAULT 5
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
    (
        SELECT * FROM match_code_embeddings(query_embedding, 'MCC', match_threshold, match_count_per_type)
        UNION ALL
        SELECT * FROM match_code_embeddings(query_embedding, 'SIC', match_threshold, match_count_per_type)
        UNION ALL
        SELECT * FROM match_code_embeddings(query_embedding, 'NAICS', match_threshold, match_count_per_type)
    )
    ORDER BY similarity DESC;
END;
$$;

-- Step 6: Grant permissions
GRANT SELECT ON code_embeddings TO authenticated;
GRANT SELECT ON code_embeddings TO anon;
GRANT EXECUTE ON FUNCTION match_code_embeddings TO authenticated;
GRANT EXECUTE ON FUNCTION match_code_embeddings TO anon;
GRANT EXECUTE ON FUNCTION match_code_embeddings_all_types TO authenticated;
GRANT EXECUTE ON FUNCTION match_code_embeddings_all_types TO anon;

-- Step 7: Add helpful comments
COMMENT ON TABLE code_embeddings IS 'Pre-computed embeddings for all industry codes (MCC/SIC/NAICS) using all-MiniLM-L6-v2';
COMMENT ON COLUMN code_embeddings.embedding IS '384-dimensional embedding vector from sentence-transformers/all-MiniLM-L6-v2';
COMMENT ON FUNCTION match_code_embeddings IS 'Find similar codes using vector similarity search (cosine distance)';

-- Step 8: Create trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_code_embeddings_updated_at
    BEFORE UPDATE ON code_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

