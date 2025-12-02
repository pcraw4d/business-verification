-- Migration: Add industry-topics mapping table for enhanced topic modeling
-- This table enables database-driven topic-to-industry mapping with relevance scores
-- Created: 2025-01-XX
-- Purpose: Support Phase 1.3 - Enhanced topic modeling with industry mapping

-- Create industry_topics table for topic-industry relationships
CREATE TABLE IF NOT EXISTS industry_topics (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    topic VARCHAR(100) NOT NULL,
    relevance_score DECIMAL(3,2) DEFAULT 0.80 CHECK (relevance_score >= 0.00 AND relevance_score <= 1.00),
    topic_type VARCHAR(50) DEFAULT 'keyword' CHECK (topic_type IN ('keyword', 'phrase', 'concept')),
    usage_count INTEGER DEFAULT 0,
    accuracy_score DECIMAL(3,2) DEFAULT 0.75, -- Historical accuracy for this topic-industry pair
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, topic)
);

-- Indexes for industry_topics table
CREATE INDEX IF NOT EXISTS idx_industry_topics_industry ON industry_topics(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_topics_topic ON industry_topics(topic);
CREATE INDEX IF NOT EXISTS idx_industry_topics_relevance ON industry_topics(relevance_score DESC);
CREATE INDEX IF NOT EXISTS idx_industry_topics_accuracy ON industry_topics(accuracy_score DESC);
CREATE INDEX IF NOT EXISTS idx_industry_topics_topic_type ON industry_topics(topic_type);

-- Full-text search index for topic matching
CREATE INDEX IF NOT EXISTS idx_industry_topics_topic_fts ON industry_topics USING gin(to_tsvector('english', topic));

COMMENT ON TABLE industry_topics IS 
    'Maps topics (keywords/phrases) to industries with relevance and accuracy scores for enhanced topic modeling';

COMMENT ON COLUMN industry_topics.relevance_score IS 
    'Relevance score indicating how well this topic represents the industry (0.0-1.0)';

COMMENT ON COLUMN industry_topics.accuracy_score IS 
    'Historical accuracy score for this topic-industry pair based on classification results';

-- Function to update accuracy scores based on classification feedback
CREATE OR REPLACE FUNCTION update_topic_accuracy(
    p_industry_id INTEGER,
    p_topic VARCHAR,
    p_is_correct BOOLEAN
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    current_accuracy DECIMAL(3,2);
    new_accuracy DECIMAL(3,2);
    current_count INTEGER;
BEGIN
    -- Get current accuracy and usage count
    SELECT accuracy_score, usage_count INTO current_accuracy, current_count
    FROM industry_topics
    WHERE industry_id = p_industry_id AND topic = p_topic;
    
    IF current_accuracy IS NULL THEN
        current_accuracy := 0.75; -- Default accuracy
    END IF;
    
    IF current_count IS NULL THEN
        current_count := 0;
    END IF;
    
    -- Calculate new accuracy using exponential moving average
    -- Weight recent results more heavily
    IF p_is_correct THEN
        new_accuracy := current_accuracy * 0.9 + 1.0 * 0.1;
    ELSE
        new_accuracy := current_accuracy * 0.9 + 0.0 * 0.1;
    END IF;
    
    -- Update accuracy and increment usage count
    UPDATE industry_topics
    SET 
        accuracy_score = new_accuracy,
        usage_count = current_count + 1,
        updated_at = NOW()
    WHERE industry_id = p_industry_id AND topic = p_topic;
END;
$$;

COMMENT ON FUNCTION update_topic_accuracy IS 
    'Updates topic accuracy score based on classification feedback using exponential moving average';

-- Fix any existing rows with invalid relevance_score values (if table already exists)
UPDATE industry_topics 
SET relevance_score = LEAST(relevance_score, 1.0)
WHERE relevance_score > 1.0;

-- Populate industry_topics with initial data from existing industry_keywords
-- This creates topic mappings based on existing keyword data
INSERT INTO industry_topics (industry_id, topic, relevance_score, topic_type, accuracy_score)
SELECT 
    ik.industry_id,
    LOWER(ik.keyword) as topic,
    LEAST(ik.weight, 1.0) as relevance_score, -- Cap at 1.0 to satisfy CHECK constraint
    CASE 
        WHEN ik.keyword LIKE '% %' THEN 'phrase'
        ELSE 'keyword'
    END as topic_type,
    0.75 as accuracy_score -- Default accuracy, will be updated based on usage
FROM industry_keywords ik
WHERE ik.is_active = true
ON CONFLICT (industry_id, topic) DO UPDATE
SET relevance_score = LEAST(EXCLUDED.relevance_score, 1.0); -- Update existing rows and cap value

-- Create view for easy querying of topic-industry relationships
CREATE OR REPLACE VIEW industry_topics_view AS
SELECT 
    it.id,
    it.industry_id,
    i.name as industry_name,
    it.topic,
    it.relevance_score,
    it.topic_type,
    it.accuracy_score,
    it.usage_count,
    it.updated_at
FROM industry_topics it
JOIN industries i ON i.id = it.industry_id
WHERE i.is_active = true;

COMMENT ON VIEW industry_topics_view IS 
    'View providing industry-topics relationships with industry names for easy querying';

