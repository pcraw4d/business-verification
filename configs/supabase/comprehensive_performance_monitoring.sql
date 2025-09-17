-- Comprehensive Performance Monitoring for Business Classification System
-- This script provides comprehensive performance monitoring infrastructure

-- 1. Create comprehensive performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(50) NOT NULL,
    metric_category VARCHAR(100) NOT NULL,
    threshold_warning DECIMAL(15,4),
    threshold_critical DECIMAL(15,4),
    status VARCHAR(20) NOT NULL CHECK (status IN ('OK', 'WARNING', 'CRITICAL')),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    details JSONB,
    recommendations TEXT
);

-- 2. Create performance alerts table
CREATE TABLE IF NOT EXISTS performance_alerts (
    id SERIAL PRIMARY KEY,
    alert_id VARCHAR(255) UNIQUE NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    metric_type VARCHAR(100) NOT NULL,
    threshold DECIMAL(15,4) NOT NULL,
    actual_value DECIMAL(15,4) NOT NULL,
    message TEXT NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- 3. Create response time metrics table
CREATE TABLE IF NOT EXISTS response_time_metrics (
    id SERIAL PRIMARY KEY,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_count BIGINT DEFAULT 0,
    total_time_ms DECIMAL(15,3) DEFAULT 0,
    average_time_ms DECIMAL(15,3) DEFAULT 0,
    min_time_ms DECIMAL(15,3) DEFAULT 0,
    max_time_ms DECIMAL(15,3) DEFAULT 0,
    p50_time_ms DECIMAL(15,3) DEFAULT 0,
    p95_time_ms DECIMAL(15,3) DEFAULT 0,
    p99_time_ms DECIMAL(15,3) DEFAULT 0,
    slow_request_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create memory usage metrics table
CREATE TABLE IF NOT EXISTS memory_metrics (
    id SERIAL PRIMARY KEY,
    allocated_mb DECIMAL(15,3) NOT NULL,
    total_allocated_mb DECIMAL(15,3) NOT NULL,
    system_mb DECIMAL(15,3) NOT NULL,
    num_gc INTEGER NOT NULL,
    gc_pause_time_ms DECIMAL(15,3) NOT NULL,
    heap_objects BIGINT NOT NULL,
    stack_in_use_mb DECIMAL(15,3) NOT NULL,
    goroutine_count INTEGER NOT NULL,
    last_gc TIMESTAMP WITH TIME ZONE,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Create database performance metrics table
CREATE TABLE IF NOT EXISTS database_performance_metrics (
    id SERIAL PRIMARY KEY,
    connection_count INTEGER NOT NULL,
    active_connections INTEGER NOT NULL,
    idle_connections INTEGER NOT NULL,
    max_connections INTEGER NOT NULL,
    query_count BIGINT DEFAULT 0,
    slow_query_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    avg_query_time_ms DECIMAL(15,3) DEFAULT 0,
    max_query_time_ms DECIMAL(15,3) DEFAULT 0,
    database_size_bytes BIGINT DEFAULT 0,
    cache_hit_ratio DECIMAL(5,2) DEFAULT 0,
    lock_count INTEGER DEFAULT 0,
    deadlock_count INTEGER DEFAULT 0,
    uptime_seconds BIGINT DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Create security validation metrics table
CREATE TABLE IF NOT EXISTS security_validation_metrics (
    id SERIAL PRIMARY KEY,
    total_validations BIGINT DEFAULT 0,
    trusted_data_source_validations BIGINT DEFAULT 0,
    website_verification_count BIGINT DEFAULT 0,
    average_validation_time_ms DECIMAL(15,3) DEFAULT 0,
    average_website_verification_ms DECIMAL(15,3) DEFAULT 0,
    security_violations BIGINT DEFAULT 0,
    trusted_data_source_rate DECIMAL(5,2) DEFAULT 0,
    website_verification_rate DECIMAL(5,2) DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_performance_metrics_category ON performance_metrics(metric_category);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_recorded_at ON performance_metrics(recorded_at);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_status ON performance_metrics(status);

CREATE INDEX IF NOT EXISTS idx_performance_alerts_severity ON performance_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_resolved ON performance_alerts(resolved);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_created_at ON performance_alerts(created_at);

CREATE INDEX IF NOT EXISTS idx_response_time_endpoint ON response_time_metrics(endpoint);
CREATE INDEX IF NOT EXISTS idx_response_time_last_updated ON response_time_metrics(last_updated);

CREATE INDEX IF NOT EXISTS idx_memory_metrics_recorded_at ON memory_metrics(recorded_at);

CREATE INDEX IF NOT EXISTS idx_database_performance_recorded_at ON database_performance_metrics(recorded_at);

CREATE INDEX IF NOT EXISTS idx_security_validation_recorded_at ON security_validation_metrics(recorded_at);

-- 8. Create function to collect comprehensive performance metrics
CREATE OR REPLACE FUNCTION collect_comprehensive_performance_metrics() 
RETURNS TABLE (
    metric_name TEXT,
    metric_value DECIMAL,
    metric_unit TEXT,
    metric_category TEXT,
    status TEXT,
    details JSONB,
    recommendations TEXT
) AS $$
DECLARE
    db_size DECIMAL;
    db_connections INTEGER;
    slow_queries INTEGER;
    cache_hit_ratio DECIMAL;
    memory_usage DECIMAL;
    response_time_avg DECIMAL;
    security_validation_avg DECIMAL;
    overall_health_score DECIMAL;
BEGIN
    -- Database size metrics
    SELECT pg_database_size(current_database()) / (1024 * 1024) INTO db_size;
    
    -- Database connections
    SELECT count(*) INTO db_connections FROM pg_stat_activity WHERE datname = current_database();
    
    -- Slow queries (queries taking more than 100ms)
    SELECT count(*) INTO slow_queries 
    FROM pg_stat_statements 
    WHERE mean_exec_time > 100 AND dbid = (SELECT oid FROM pg_database WHERE datname = current_database());
    
    -- Cache hit ratio
    SELECT 
        round(100.0 * sum(blks_hit) / (sum(blks_hit) + sum(blks_read)), 2) 
    INTO cache_hit_ratio
    FROM pg_stat_database 
    WHERE datname = current_database();
    
    -- Memory usage (from application metrics)
    SELECT COALESCE(avg(metric_value), 0) INTO memory_usage
    FROM performance_metrics 
    WHERE metric_category = 'memory' 
    AND recorded_at > NOW() - INTERVAL '1 hour';
    
    -- Average response time
    SELECT COALESCE(avg(metric_value), 0) INTO response_time_avg
    FROM performance_metrics 
    WHERE metric_category = 'response_time' 
    AND recorded_at > NOW() - INTERVAL '1 hour';
    
    -- Average security validation time
    SELECT COALESCE(avg(metric_value), 0) INTO security_validation_avg
    FROM performance_metrics 
    WHERE metric_category = 'security' 
    AND recorded_at > NOW() - INTERVAL '1 hour';
    
    -- Calculate overall health score (0-100)
    overall_health_score := 100.0;
    
    -- Deduct points for issues
    IF cache_hit_ratio < 90 THEN
        overall_health_score := overall_health_score - (90 - cache_hit_ratio);
    END IF;
    
    IF slow_queries > 10 THEN
        overall_health_score := overall_health_score - (slow_queries * 2);
    END IF;
    
    IF response_time_avg > 500 THEN
        overall_health_score := overall_health_score - ((response_time_avg - 500) / 10);
    END IF;
    
    IF memory_usage > 512 THEN
        overall_health_score := overall_health_score - ((memory_usage - 512) / 10);
    END IF;
    
    -- Ensure score is between 0 and 100
    overall_health_score := GREATEST(0, LEAST(100, overall_health_score));
    
    -- Return metrics
    RETURN QUERY
    SELECT 
        'database_size_mb'::TEXT,
        db_size,
        'MB'::TEXT,
        'database'::TEXT,
        CASE 
            WHEN db_size > 1000 THEN 'WARNING'
            WHEN db_size > 2000 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('database_name', current_database()),
        CASE 
            WHEN db_size > 1000 THEN 'Consider database cleanup or archiving old data'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'active_connections'::TEXT,
        db_connections,
        'count'::TEXT,
        'database'::TEXT,
        CASE 
            WHEN db_connections > 50 THEN 'WARNING'
            WHEN db_connections > 80 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('max_connections', 100),
        CASE 
            WHEN db_connections > 50 THEN 'High connection count detected'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'slow_queries'::TEXT,
        slow_queries,
        'count'::TEXT,
        'database'::TEXT,
        CASE 
            WHEN slow_queries > 5 THEN 'WARNING'
            WHEN slow_queries > 20 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('threshold_ms', 100),
        CASE 
            WHEN slow_queries > 5 THEN 'Optimize slow queries or add indexes'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'cache_hit_ratio'::TEXT,
        cache_hit_ratio,
        'percent'::TEXT,
        'database'::TEXT,
        CASE 
            WHEN cache_hit_ratio < 90 THEN 'WARNING'
            WHEN cache_hit_ratio < 80 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('target_percent', 95),
        CASE 
            WHEN cache_hit_ratio < 90 THEN 'Increase shared_buffers or optimize queries'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'memory_usage_mb'::TEXT,
        memory_usage,
        'MB'::TEXT,
        'memory'::TEXT,
        CASE 
            WHEN memory_usage > 512 THEN 'WARNING'
            WHEN memory_usage > 1024 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('threshold_mb', 512),
        CASE 
            WHEN memory_usage > 512 THEN 'Monitor memory usage and consider optimization'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'response_time_avg_ms'::TEXT,
        response_time_avg,
        'ms'::TEXT,
        'response_time'::TEXT,
        CASE 
            WHEN response_time_avg > 500 THEN 'WARNING'
            WHEN response_time_avg > 1000 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('threshold_ms', 500),
        CASE 
            WHEN response_time_avg > 500 THEN 'Optimize application performance'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'security_validation_avg_ms'::TEXT,
        security_validation_avg,
        'ms'::TEXT,
        'security'::TEXT,
        CASE 
            WHEN security_validation_avg > 50 THEN 'WARNING'
            WHEN security_validation_avg > 100 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('threshold_ms', 50),
        CASE 
            WHEN security_validation_avg > 50 THEN 'Optimize security validation performance'
            ELSE NULL
        END::TEXT
    
    UNION ALL
    
    SELECT 
        'overall_health_score'::TEXT,
        overall_health_score,
        'score'::TEXT,
        'system'::TEXT,
        CASE 
            WHEN overall_health_score < 70 THEN 'WARNING'
            WHEN overall_health_score < 50 THEN 'CRITICAL'
            ELSE 'OK'
        END::TEXT,
        jsonb_build_object('target_score', 90),
        CASE 
            WHEN overall_health_score < 70 THEN 'System performance needs attention'
            ELSE NULL
        END::TEXT;
END;
$$ LANGUAGE plpgsql;

-- 9. Create function to get performance summary
CREATE OR REPLACE FUNCTION get_performance_summary(
    p_hours INTEGER DEFAULT 1
) RETURNS TABLE (
    metric_category TEXT,
    total_metrics BIGINT,
    avg_value DECIMAL,
    min_value DECIMAL,
    max_value DECIMAL,
    warning_count BIGINT,
    critical_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pm.metric_category,
        count(*) as total_metrics,
        round(avg(pm.metric_value), 2) as avg_value,
        round(min(pm.metric_value), 2) as min_value,
        round(max(pm.metric_value), 2) as max_value,
        count(*) FILTER (WHERE pm.status = 'WARNING') as warning_count,
        count(*) FILTER (WHERE pm.status = 'CRITICAL') as critical_count
    FROM performance_metrics pm
    WHERE pm.recorded_at > NOW() - (p_hours || ' hours')::INTERVAL
    GROUP BY pm.metric_category
    ORDER BY pm.metric_category;
END;
$$ LANGUAGE plpgsql;

-- 10. Create function to clean up old metrics
CREATE OR REPLACE FUNCTION cleanup_old_performance_metrics(
    p_retention_days INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete old performance metrics
    DELETE FROM performance_metrics 
    WHERE recorded_at < NOW() - (p_retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Delete old alerts (keep resolved ones for longer)
    DELETE FROM performance_alerts 
    WHERE resolved = TRUE 
    AND resolved_at < NOW() - (p_retention_days * 2 || ' days')::INTERVAL;
    
    -- Delete old response time metrics
    DELETE FROM response_time_metrics 
    WHERE recorded_at < NOW() - (p_retention_days || ' days')::INTERVAL;
    
    -- Delete old memory metrics
    DELETE FROM memory_metrics 
    WHERE recorded_at < NOW() - (p_retention_days || ' days')::INTERVAL;
    
    -- Delete old database performance metrics
    DELETE FROM database_performance_metrics 
    WHERE recorded_at < NOW() - (p_retention_days || ' days')::INTERVAL;
    
    -- Delete old security validation metrics
    DELETE FROM security_validation_metrics 
    WHERE recorded_at < NOW() - (p_retention_days || ' days')::INTERVAL;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 11. Create view for performance dashboard
CREATE OR REPLACE VIEW performance_dashboard AS
SELECT 
    pm.metric_category,
    pm.metric_name,
    pm.metric_value,
    pm.metric_unit,
    pm.status,
    pm.recorded_at,
    pm.details,
    pm.recommendations,
    CASE 
        WHEN pm.status = 'CRITICAL' THEN 3
        WHEN pm.status = 'WARNING' THEN 2
        ELSE 1
    END as priority
FROM performance_metrics pm
WHERE pm.recorded_at > NOW() - INTERVAL '24 hours'
ORDER BY 
    priority DESC,
    pm.recorded_at DESC;

-- 12. Create view for alert summary
CREATE OR REPLACE VIEW alert_summary AS
SELECT 
    pa.severity,
    pa.alert_type,
    count(*) as alert_count,
    count(*) FILTER (WHERE pa.resolved = FALSE) as active_count,
    count(*) FILTER (WHERE pa.resolved = TRUE) as resolved_count,
    min(pa.created_at) as first_alert,
    max(pa.created_at) as last_alert
FROM performance_alerts pa
WHERE pa.created_at > NOW() - INTERVAL '7 days'
GROUP BY pa.severity, pa.alert_type
ORDER BY 
    CASE pa.severity 
        WHEN 'critical' THEN 4
        WHEN 'high' THEN 3
        WHEN 'medium' THEN 2
        ELSE 1
    END DESC,
    alert_count DESC;

-- 13. Create trigger to automatically collect metrics
CREATE OR REPLACE FUNCTION trigger_performance_metrics_collection()
RETURNS TRIGGER AS $$
BEGIN
    -- Insert collected metrics
    INSERT INTO performance_metrics (metric_name, metric_value, metric_unit, metric_category, status, details, recommendations)
    SELECT * FROM collect_comprehensive_performance_metrics();
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 14. Create scheduled job to collect metrics (if pg_cron is available)
-- This would be set up in the application or via a cron job
-- SELECT cron.schedule('collect-performance-metrics', '*/5 * * * *', 'SELECT trigger_performance_metrics_collection();');

-- 15. Grant permissions (adjust as needed for your setup)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON performance_metrics TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON performance_alerts TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON response_time_metrics TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON memory_metrics TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON database_performance_metrics TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON security_validation_metrics TO your_app_user;
-- GRANT SELECT ON performance_dashboard TO your_app_user;
-- GRANT SELECT ON alert_summary TO your_app_user;
