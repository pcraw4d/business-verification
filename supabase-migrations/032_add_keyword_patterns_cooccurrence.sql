-- Migration: Add keyword patterns table for co-occurrence analysis
-- This table enables relationship analysis between keywords and entities
-- Created: 2025-01-XX
-- Purpose: Support Phase 1.4 - Enhanced co-occurrence strategy with relationship analysis

-- Create keyword_patterns table for co-occurrence pattern storage
CREATE TABLE IF NOT EXISTS keyword_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword_pair VARCHAR(200) NOT NULL, -- Format: "keyword1|keyword2" (sorted alphabetically)
    keyword1 VARCHAR(100) NOT NULL,
    keyword2 VARCHAR(100) NOT NULL,
    co_occurrence_score DECIMAL(3,2) DEFAULT 0.75 CHECK (co_occurrence_score >= 0.00 AND co_occurrence_score <= 1.00),
    pattern_type VARCHAR(50) DEFAULT 'keyword_keyword' CHECK (pattern_type IN ('keyword_keyword', 'entity_keyword', 'entity_entity')),
    frequency INTEGER DEFAULT 1, -- How often this pattern appears in training data
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword_pair)
);

-- Indexes for keyword_patterns table
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_industry ON keyword_patterns(industry_id);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_pair ON keyword_patterns(keyword_pair);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_keyword1 ON keyword_patterns(keyword1);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_keyword2 ON keyword_patterns(keyword2);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_score ON keyword_patterns(co_occurrence_score DESC);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_type ON keyword_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_frequency ON keyword_patterns(frequency DESC);

-- Composite index for common query pattern
CREATE INDEX IF NOT EXISTS idx_keyword_patterns_industry_pair ON keyword_patterns(industry_id, keyword_pair);

COMMENT ON TABLE keyword_patterns IS 
    'Stores co-occurrence patterns (keyword pairs) for industries to enable relationship analysis';

COMMENT ON COLUMN keyword_patterns.keyword_pair IS 
    'Normalized pair format: "keyword1|keyword2" (sorted alphabetically for consistency)';

COMMENT ON COLUMN keyword_patterns.co_occurrence_score IS 
    'Score indicating how strongly this keyword pair indicates the industry (0.0-1.0)';

COMMENT ON COLUMN keyword_patterns.pattern_type IS 
    'Type of pattern: keyword_keyword, entity_keyword, or entity_entity';

-- Function to normalize keyword pair (ensures consistent ordering)
CREATE OR REPLACE FUNCTION normalize_keyword_pair(kw1 TEXT, kw2 TEXT)
RETURNS TEXT
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
    -- Sort keywords alphabetically to ensure consistent pair format
    IF LOWER(kw1) < LOWER(kw2) THEN
        RETURN LOWER(kw1) || '|' || LOWER(kw2);
    ELSE
        RETURN LOWER(kw2) || '|' || LOWER(kw1);
    END IF;
END;
$$;

COMMENT ON FUNCTION normalize_keyword_pair IS 
    'Normalizes keyword pairs by sorting alphabetically for consistent storage';

-- Function to find industries by keyword patterns
CREATE OR REPLACE FUNCTION find_industries_by_patterns(
    p_patterns TEXT[]
)
RETURNS TABLE (
    industry_id INTEGER,
    industry_name TEXT,
    pattern_matches INTEGER,
    avg_score DECIMAL(3,2),
    matched_patterns TEXT[]
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.id AS industry_id,
        i.name AS industry_name,
        COUNT(DISTINCT kp.keyword_pair)::INTEGER AS pattern_matches,
        AVG(kp.co_occurrence_score)::DECIMAL(3,2) AS avg_score,
        ARRAY_AGG(DISTINCT kp.keyword_pair) AS matched_patterns
    FROM industries i
    JOIN keyword_patterns kp ON kp.industry_id = i.id
    WHERE kp.keyword_pair = ANY(p_patterns)
        AND i.is_active = true
    GROUP BY i.id, i.name
    HAVING COUNT(DISTINCT kp.keyword_pair) >= 2 -- Require at least 2 pattern matches
    ORDER BY pattern_matches DESC, avg_score DESC
    LIMIT 10;
END;
$$;

COMMENT ON FUNCTION find_industries_by_patterns IS 
    'Finds industries matching keyword co-occurrence patterns. Requires at least 2 pattern matches.';

-- Populate initial keyword_patterns from existing keyword_weights
-- Generate pairs from keywords within the same industry
INSERT INTO keyword_patterns (industry_id, keyword_pair, keyword1, keyword2, co_occurrence_score, pattern_type, frequency)
SELECT DISTINCT
    kw1.industry_id,
    normalize_keyword_pair(kw1.keyword, kw2.keyword) AS keyword_pair,
    LEAST(LOWER(kw1.keyword), LOWER(kw2.keyword)) AS keyword1,
    GREATEST(LOWER(kw1.keyword), LOWER(kw2.keyword)) AS keyword2,
    LEAST((kw1.base_weight + kw2.base_weight) / 2.0, 1.0) AS co_occurrence_score,
    'keyword_keyword' AS pattern_type,
    1 AS frequency
FROM keyword_weights kw1
JOIN keyword_weights kw2 ON kw2.industry_id = kw1.industry_id
WHERE kw1.is_active = true
    AND kw2.is_active = true
    AND kw1.keyword < kw2.keyword -- Avoid duplicates and self-pairs
    AND kw1.industry_id = kw2.industry_id
ON CONFLICT (industry_id, keyword_pair) DO UPDATE
SET 
    co_occurrence_score = GREATEST(keyword_patterns.co_occurrence_score, EXCLUDED.co_occurrence_score),
    frequency = keyword_patterns.frequency + 1,
    updated_at = NOW();

-- Create view for easy querying
CREATE OR REPLACE VIEW keyword_patterns_view AS
SELECT 
    kp.id,
    kp.industry_id,
    i.name AS industry_name,
    kp.keyword_pair,
    kp.keyword1,
    kp.keyword2,
    kp.co_occurrence_score,
    kp.pattern_type,
    kp.frequency,
    kp.updated_at
FROM keyword_patterns kp
JOIN industries i ON i.id = kp.industry_id
WHERE i.is_active = true;

COMMENT ON VIEW keyword_patterns_view IS 
    'View providing keyword pattern relationships with industry names for easy querying';

