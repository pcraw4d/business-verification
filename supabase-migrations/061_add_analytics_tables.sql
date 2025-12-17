-- Migration: Add analytics and monitoring tables
-- Phase 5, Day 2: Metrics tracking and dashboard for classification monitoring

-- Step 1: Create classification_metrics table
CREATE TABLE classification_metrics (
    id BIGSERIAL PRIMARY KEY,
    
    -- Request info
    request_id VARCHAR(36),
    business_name VARCHAR(255),
    website_url TEXT,
    
    -- Classification result
    primary_industry VARCHAR(255),
    confidence DECIMAL(5,4),
    layer_used VARCHAR(20),
    method VARCHAR(50),
    
    -- Performance
    total_time_ms INTEGER,
    scrape_time_ms INTEGER,
    layer1_time_ms INTEGER,
    layer2_time_ms INTEGER,
    layer3_time_ms INTEGER,
    
    -- Cache
    from_cache BOOLEAN DEFAULT FALSE,
    
    -- Codes
    mcc_codes JSONB,
    sic_codes JSONB,
    naics_codes JSONB,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

-- Step 2: Create indexes
CREATE INDEX idx_metrics_created_at ON classification_metrics(created_at DESC);
CREATE INDEX idx_metrics_layer_used ON classification_metrics(layer_used);
CREATE INDEX idx_metrics_from_cache ON classification_metrics(from_cache);
CREATE INDEX idx_metrics_confidence ON classification_metrics(confidence);
CREATE INDEX idx_metrics_total_time ON classification_metrics(total_time_ms);
CREATE INDEX idx_metrics_request_id ON classification_metrics(request_id);

-- Step 3: Create materialized view for dashboard
CREATE MATERIALIZED VIEW classification_dashboard AS
SELECT
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as total_classifications,
    COUNT(*) FILTER (WHERE from_cache) as cache_hits,
    COUNT(*) FILTER (WHERE NOT from_cache) as cache_misses,
    ROUND(AVG(confidence), 4) as avg_confidence,
    ROUND(AVG(total_time_ms), 0) as avg_total_time_ms,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%') as layer1_count,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer2%') as layer2_count,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%') as layer3_count,
    COUNT(*) FILTER (WHERE confidence >= 0.90) as high_confidence_count,
    COUNT(*) FILTER (WHERE confidence < 0.70) as low_confidence_count
FROM classification_metrics
WHERE created_at >= NOW() - INTERVAL '90 days'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY date DESC;

-- Step 4: Create index on materialized view
CREATE UNIQUE INDEX ON classification_dashboard (date);

-- Step 5: Create function to refresh dashboard
CREATE OR REPLACE FUNCTION refresh_classification_dashboard()
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY classification_dashboard;
END;
$$;

-- Step 6: Create function to get dashboard summary
CREATE OR REPLACE FUNCTION get_dashboard_summary(days INTEGER DEFAULT 30)
RETURNS TABLE (
    metric VARCHAR,
    value NUMERIC,
    description TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM (
        VALUES
            ('total_classifications', 
             (SELECT COUNT(*)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Total classifications in period'),
            
            ('cache_hit_rate',
             (SELECT ROUND(COUNT(*) FILTER (WHERE from_cache)::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage of requests served from cache'),
            
            ('avg_confidence',
             (SELECT ROUND(AVG(confidence), 4)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Average confidence score'),
            
            ('avg_processing_time_ms',
             (SELECT ROUND(AVG(total_time_ms), 0)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL 
              AND NOT from_cache),
             'Average processing time (non-cached)'),
            
            ('layer1_percentage',
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 1'),
            
            ('layer2_percentage',
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer2%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 2'),
            
            ('layer3_percentage',
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 3 (LLM)')
    ) AS stats(metric, value, description);
END;
$$;

-- Step 7: Create function to get time series data
CREATE OR REPLACE FUNCTION get_dashboard_timeseries(days INTEGER DEFAULT 30)
RETURNS TABLE (
    date DATE,
    total_classifications BIGINT,
    cache_hits BIGINT,
    cache_misses BIGINT,
    avg_confidence NUMERIC,
    avg_total_time_ms NUMERIC,
    layer1_count BIGINT,
    layer2_count BIGINT,
    layer3_count BIGINT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        DATE_TRUNC('day', created_at)::DATE as date,
        COUNT(*) as total_classifications,
        COUNT(*) FILTER (WHERE from_cache) as cache_hits,
        COUNT(*) FILTER (WHERE NOT from_cache) as cache_misses,
        ROUND(AVG(confidence), 4) as avg_confidence,
        ROUND(AVG(total_time_ms), 0) as avg_total_time_ms,
        COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%') as layer1_count,
        COUNT(*) FILTER (WHERE layer_used LIKE 'layer2%') as layer2_count,
        COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%') as layer3_count
    FROM classification_metrics
    WHERE created_at >= NOW() - (days || ' days')::INTERVAL
    GROUP BY DATE_TRUNC('day', created_at)
    ORDER BY date DESC;
END;
$$;

-- Step 8: Grant permissions
GRANT SELECT, INSERT ON classification_metrics TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE classification_metrics_id_seq TO authenticated;
GRANT SELECT ON classification_dashboard TO authenticated;
GRANT EXECUTE ON FUNCTION get_dashboard_summary TO authenticated;
GRANT EXECUTE ON FUNCTION get_dashboard_timeseries TO authenticated;
GRANT EXECUTE ON FUNCTION refresh_classification_dashboard TO authenticated;

-- Step 9: Add helpful comments
COMMENT ON TABLE classification_metrics IS 'Detailed metrics for every classification request';
COMMENT ON MATERIALIZED VIEW classification_dashboard IS 'Daily aggregated metrics for monitoring dashboard';
COMMENT ON FUNCTION get_dashboard_summary IS 'Get summary statistics for the dashboard';
COMMENT ON FUNCTION get_dashboard_timeseries IS 'Get time series data for charts';

