-- Migration: Fix dashboard summary function return type
-- Fixes: Returned type text does not match expected type character varying

-- Fix the get_dashboard_summary function to use TEXT instead of VARCHAR for metric
CREATE OR REPLACE FUNCTION get_dashboard_summary(days INTEGER DEFAULT 30)
RETURNS TABLE (
    metric TEXT,
    value NUMERIC,
    description TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM (
        VALUES
            ('total_classifications'::TEXT, 
             (SELECT COUNT(*)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Total classifications in period'::TEXT),
            
            ('cache_hit_rate'::TEXT,
             (SELECT ROUND(COUNT(*) FILTER (WHERE from_cache)::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage of requests served from cache'::TEXT),
            
            ('avg_confidence'::TEXT,
             (SELECT ROUND(AVG(confidence), 4)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Average confidence score'::TEXT),
            
            ('avg_processing_time_ms'::TEXT,
             (SELECT ROUND(AVG(total_time_ms), 0)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL 
              AND NOT from_cache),
             'Average processing time (non-cached)'::TEXT),
            
            ('layer1_percentage'::TEXT,
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 1'::TEXT),
            
            ('layer2_percentage'::TEXT,
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer2%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 2'::TEXT),
            
            ('layer3_percentage'::TEXT,
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 3 (LLM)'::TEXT)
    ) AS stats(metric, value, description);
END;
$$;

-- Grant permissions
GRANT EXECUTE ON FUNCTION get_dashboard_summary TO authenticated;

-- Add comment
COMMENT ON FUNCTION get_dashboard_summary IS 'Get summary statistics for the dashboard (fixed return type)';

