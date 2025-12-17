-- Migration: Add classification caching infrastructure
-- Phase 5, Day 1: 30-day database-backed cache for classification results

-- Step 1: Create classification_cache table
CREATE TABLE classification_cache (
    id BIGSERIAL PRIMARY KEY,
    
    -- Cache key (content hash)
    content_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hash of website content
    
    -- Input data (for debugging)
    business_name VARCHAR(255),
    website_url TEXT,
    
    -- Cached result
    classification_result JSONB NOT NULL,
    
    -- Metadata
    layer_used VARCHAR(20),  -- layer1_high_conf, layer2_better, layer3_high_conf
    confidence DECIMAL(5,4),
    processing_time_ms INTEGER,
    
    -- Cache management
    created_at TIMESTAMPTZ DEFAULT NOW(),
    accessed_at TIMESTAMPTZ DEFAULT NOW(),
    access_count INTEGER DEFAULT 1,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '30 days'),
    
    -- Constraints
    CONSTRAINT valid_confidence CHECK (confidence >= 0 AND confidence <= 1)
);

-- Step 2: Create indexes for fast lookups
CREATE INDEX idx_cache_content_hash ON classification_cache(content_hash);
CREATE INDEX idx_cache_expires_at ON classification_cache(expires_at);
CREATE INDEX idx_cache_accessed_at ON classification_cache(accessed_at);
CREATE INDEX idx_cache_created_at ON classification_cache(created_at);

-- Step 3: Create function to get cached result
CREATE OR REPLACE FUNCTION get_cached_classification(
    p_content_hash VARCHAR(64)
)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    cached_result JSONB;
BEGIN
    -- Get result if not expired
    SELECT classification_result INTO cached_result
    FROM classification_cache
    WHERE content_hash = p_content_hash
        AND expires_at > NOW()
    LIMIT 1;
    
    -- Update access stats if found
    IF cached_result IS NOT NULL THEN
        UPDATE classification_cache
        SET accessed_at = NOW(),
            access_count = access_count + 1
        WHERE content_hash = p_content_hash;
    END IF;
    
    RETURN cached_result;
END;
$$;

-- Step 4: Create function to set cached result
CREATE OR REPLACE FUNCTION set_cached_classification(
    p_content_hash VARCHAR(64),
    p_business_name VARCHAR(255),
    p_website_url TEXT,
    p_result JSONB,
    p_layer_used VARCHAR(20),
    p_confidence DECIMAL(5,4),
    p_processing_time_ms INTEGER
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO classification_cache (
        content_hash,
        business_name,
        website_url,
        classification_result,
        layer_used,
        confidence,
        processing_time_ms
    )
    VALUES (
        p_content_hash,
        p_business_name,
        p_website_url,
        p_result,
        p_layer_used,
        p_confidence,
        p_processing_time_ms
    )
    ON CONFLICT (content_hash) DO UPDATE
    SET
        classification_result = p_result,
        layer_used = p_layer_used,
        confidence = p_confidence,
        processing_time_ms = p_processing_time_ms,
        accessed_at = NOW(),
        access_count = classification_cache.access_count + 1,
        expires_at = NOW() + INTERVAL '30 days';
END;
$$;

-- Step 5: Create cleanup function for expired entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM classification_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$;

-- Step 6: Grant permissions
GRANT SELECT, INSERT, UPDATE ON classification_cache TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE classification_cache_id_seq TO authenticated;
GRANT EXECUTE ON FUNCTION get_cached_classification TO authenticated;
GRANT EXECUTE ON FUNCTION set_cached_classification TO authenticated;
GRANT EXECUTE ON FUNCTION cleanup_expired_cache TO authenticated;

-- Step 7: Add helpful comments
COMMENT ON TABLE classification_cache IS '30-day cache for classification results to avoid repeated processing';
COMMENT ON COLUMN classification_cache.content_hash IS 'SHA-256 hash of website content for cache key';
COMMENT ON COLUMN classification_cache.expires_at IS 'Cache entries expire after 30 days';
COMMENT ON FUNCTION get_cached_classification IS 'Retrieve cached classification if not expired';
COMMENT ON FUNCTION set_cached_classification IS 'Store or update cached classification';
COMMENT ON FUNCTION cleanup_expired_cache IS 'Remove expired cache entries';

