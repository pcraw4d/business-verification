-- Supabase Free Tier Usage Monitoring and Limits
-- This script provides comprehensive monitoring for Supabase free tier usage and limits

-- 1. Create a usage monitoring table
CREATE TABLE IF NOT EXISTS usage_monitoring (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(15,2) NOT NULL,
    metric_unit VARCHAR(50) NOT NULL,
    limit_value DECIMAL(15,2) NOT NULL,
    usage_percentage DECIMAL(5,2) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('OK', 'WARNING', 'CRITICAL')),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notes TEXT
);

-- 2. Create a function to check database storage usage
CREATE OR REPLACE FUNCTION check_database_storage_usage() 
RETURNS TABLE (
    database_name TEXT,
    total_size_mb DECIMAL,
    limit_mb DECIMAL,
    usage_percentage DECIMAL,
    status TEXT
) AS $$
DECLARE
    total_size DECIMAL;
    limit_size DECIMAL := 500; -- 500MB free tier limit
    usage_pct DECIMAL;
    status_text TEXT;
BEGIN
    -- Get total database size
    SELECT 
        ROUND(
            SUM(pg_database_size(datname)) / (1024 * 1024), 2
        ) INTO total_size
    FROM pg_database 
    WHERE datname = current_database();
    
    -- Calculate usage percentage
    usage_pct := ROUND((total_size / limit_size) * 100, 2);
    
    -- Determine status
    IF usage_pct >= 90 THEN
        status_text := 'CRITICAL';
    ELSIF usage_pct >= 75 THEN
        status_text := 'WARNING';
    ELSE
        status_text := 'OK';
    END IF;
    
    RETURN QUERY SELECT 
        current_database()::TEXT,
        total_size,
        limit_size,
        usage_pct,
        status_text;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to check table sizes
CREATE OR REPLACE FUNCTION check_table_sizes() 
RETURNS TABLE (
    table_name TEXT,
    size_mb DECIMAL,
    row_count BIGINT,
    index_size_mb DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname||'.'||tablename as table_name,
        ROUND(pg_total_relation_size(schemaname||'.'||tablename) / (1024 * 1024), 2) as size_mb,
        (SELECT COUNT(*) FROM information_schema.tables t2 WHERE t2.table_name = t1.tablename) as row_count,
        ROUND(pg_indexes_size(schemaname||'.'||tablename) / (1024 * 1024), 2) as index_size_mb
    FROM pg_tables t1
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to check connection usage
CREATE OR REPLACE FUNCTION check_connection_usage() 
RETURNS TABLE (
    current_connections INTEGER,
    max_connections INTEGER,
    usage_percentage DECIMAL,
    status TEXT
) AS $$
DECLARE
    current_conn INTEGER;
    max_conn INTEGER := 60; -- Supabase free tier limit
    usage_pct DECIMAL;
    status_text TEXT;
BEGIN
    -- Get current connections
    SELECT COUNT(*) INTO current_conn
    FROM pg_stat_activity
    WHERE state = 'active';
    
    -- Calculate usage percentage
    usage_pct := ROUND((current_conn::DECIMAL / max_conn) * 100, 2);
    
    -- Determine status
    IF usage_pct >= 90 THEN
        status_text := 'CRITICAL';
    ELSIF usage_pct >= 75 THEN
        status_text := 'WARNING';
    ELSE
        status_text := 'OK';
    END IF;
    
    RETURN QUERY SELECT 
        current_conn,
        max_conn,
        usage_pct,
        status_text;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to check query performance
CREATE OR REPLACE FUNCTION check_query_performance() 
RETURNS TABLE (
    query_type TEXT,
    avg_execution_time_ms DECIMAL,
    total_executions BIGINT,
    slow_queries BIGINT,
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'SELECT' as query_type,
        ROUND(AVG(mean_exec_time), 2) as avg_execution_time_ms,
        SUM(calls) as total_executions,
        COUNT(CASE WHEN mean_exec_time > 1000 THEN 1 END) as slow_queries,
        CASE 
            WHEN AVG(mean_exec_time) > 1000 THEN 'CRITICAL'
            WHEN AVG(mean_exec_time) > 500 THEN 'WARNING'
            ELSE 'OK'
        END as status
    FROM pg_stat_statements
    WHERE query LIKE '%SELECT%'
    
    UNION ALL
    
    SELECT 
        'INSERT' as query_type,
        ROUND(AVG(mean_exec_time), 2) as avg_execution_time_ms,
        SUM(calls) as total_executions,
        COUNT(CASE WHEN mean_exec_time > 1000 THEN 1 END) as slow_queries,
        CASE 
            WHEN AVG(mean_exec_time) > 1000 THEN 'CRITICAL'
            WHEN AVG(mean_exec_time) > 500 THEN 'WARNING'
            ELSE 'OK'
        END as status
    FROM pg_stat_statements
    WHERE query LIKE '%INSERT%'
    
    UNION ALL
    
    SELECT 
        'UPDATE' as query_type,
        ROUND(AVG(mean_exec_time), 2) as avg_execution_time_ms,
        SUM(calls) as total_executions,
        COUNT(CASE WHEN mean_exec_time > 1000 THEN 1 END) as slow_queries,
        CASE 
            WHEN AVG(mean_exec_time) > 1000 THEN 'CRITICAL'
            WHEN AVG(mean_exec_time) > 500 THEN 'WARNING'
            ELSE 'OK'
        END as status
    FROM pg_stat_statements
    WHERE query LIKE '%UPDATE%';
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to check index usage
CREATE OR REPLACE FUNCTION check_index_usage() 
RETURNS TABLE (
    table_name TEXT,
    index_name TEXT,
    index_size_mb DECIMAL,
    usage_count BIGINT,
    efficiency_score DECIMAL,
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.tablename as table_name,
        i.indexname as index_name,
        ROUND(pg_relation_size(i.indexname) / (1024 * 1024), 2) as index_size_mb,
        COALESCE(s.idx_scan, 0) as usage_count,
        ROUND(
            CASE 
                WHEN s.idx_scan > 0 THEN (s.idx_scan::DECIMAL / (s.idx_scan + s.seq_scan)) * 100
                ELSE 0
            END, 2
        ) as efficiency_score,
        CASE 
            WHEN s.idx_scan = 0 THEN 'UNUSED'
            WHEN s.idx_scan < 10 THEN 'LOW_USAGE'
            WHEN s.idx_scan < 100 THEN 'MODERATE_USAGE'
            ELSE 'HIGH_USAGE'
        END as status
    FROM pg_indexes i
    JOIN pg_tables t ON i.tablename = t.tablename
    LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexrelname
    WHERE t.schemaname = 'public'
    ORDER BY pg_relation_size(i.indexname) DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to check free tier limits
CREATE OR REPLACE FUNCTION check_free_tier_limits() 
RETURNS TABLE (
    limit_name TEXT,
    current_usage DECIMAL,
    limit_value DECIMAL,
    usage_percentage DECIMAL,
    status TEXT,
    description TEXT
) AS $$
DECLARE
    db_size DECIMAL;
    conn_count INTEGER;
    max_conn INTEGER := 60;
    conn_pct DECIMAL;
    db_pct DECIMAL;
BEGIN
    -- Get database size
    SELECT 
        ROUND(SUM(pg_database_size(datname)) / (1024 * 1024), 2)
    INTO db_size
    FROM pg_database 
    WHERE datname = current_database();
    
    -- Get connection count
    SELECT COUNT(*) INTO conn_count
    FROM pg_stat_activity
    WHERE state = 'active';
    
    -- Calculate percentages
    db_pct := ROUND((db_size / 500) * 100, 2);
    conn_pct := ROUND((conn_count::DECIMAL / max_conn) * 100, 2);
    
    RETURN QUERY
    SELECT 
        'Database Storage' as limit_name,
        db_size as current_usage,
        500.0 as limit_value,
        db_pct as usage_percentage,
        CASE 
            WHEN db_pct >= 90 THEN 'CRITICAL'
            WHEN db_pct >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        '500MB storage limit for free tier' as description
    
    UNION ALL
    
    SELECT 
        'Active Connections' as limit_name,
        conn_count as current_usage,
        max_conn as limit_value,
        conn_pct as usage_percentage,
        CASE 
            WHEN conn_pct >= 90 THEN 'CRITICAL'
            WHEN conn_pct >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        '60 concurrent connections limit for free tier' as description
    
    UNION ALL
    
    SELECT 
        'Monthly Active Users' as limit_name,
        0 as current_usage, -- This would need to be tracked separately
        50000.0 as limit_value,
        0.0 as usage_percentage,
        'OK' as status,
        '50,000 monthly active users limit for free tier' as description;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to generate usage report
CREATE OR REPLACE FUNCTION generate_usage_report() 
RETURNS TABLE (
    report_section TEXT,
    metric_name TEXT,
    current_value TEXT,
    limit_value TEXT,
    usage_percentage DECIMAL,
    status TEXT,
    recommendation TEXT
) AS $$
BEGIN
    -- Database Storage Section
    RETURN QUERY
    SELECT 
        'Database Storage' as report_section,
        'Total Database Size' as metric_name,
        (SELECT total_size_mb::TEXT || ' MB' FROM check_database_storage_usage()) as current_value,
        '500 MB' as limit_value,
        (SELECT usage_percentage FROM check_database_storage_usage()) as usage_percentage,
        (SELECT status FROM check_database_storage_usage()) as status,
        CASE 
            WHEN (SELECT usage_percentage FROM check_database_storage_usage()) >= 90 THEN 'Consider upgrading to paid plan or optimizing data'
            WHEN (SELECT usage_percentage FROM check_database_storage_usage()) >= 75 THEN 'Monitor storage usage closely'
            ELSE 'Storage usage is within acceptable limits'
        END as recommendation
    
    UNION ALL
    
    -- Connection Usage Section
    SELECT 
        'Connection Usage' as report_section,
        'Active Connections' as metric_name,
        (SELECT current_connections::TEXT FROM check_connection_usage()) as current_value,
        '60' as limit_value,
        (SELECT usage_percentage FROM check_connection_usage()) as usage_percentage,
        (SELECT status FROM check_connection_usage()) as status,
        CASE 
            WHEN (SELECT usage_percentage FROM check_connection_usage()) >= 90 THEN 'Consider connection pooling or upgrading plan'
            WHEN (SELECT usage_percentage FROM check_connection_usage()) >= 75 THEN 'Monitor connection usage closely'
            ELSE 'Connection usage is within acceptable limits'
        END as recommendation
    
    UNION ALL
    
    -- Query Performance Section
    SELECT 
        'Query Performance' as report_section,
        'Average SELECT Time' as metric_name,
        (SELECT avg_execution_time_ms::TEXT || ' ms' FROM check_query_performance() WHERE query_type = 'SELECT') as current_value,
        '500 ms' as limit_value,
        CASE 
            WHEN (SELECT avg_execution_time_ms FROM check_query_performance() WHERE query_type = 'SELECT') > 500 THEN 100.0
            ELSE (SELECT avg_execution_time_ms FROM check_query_performance() WHERE query_type = 'SELECT') / 5.0
        END as usage_percentage,
        (SELECT status FROM check_query_performance() WHERE query_type = 'SELECT') as status,
        CASE 
            WHEN (SELECT avg_execution_time_ms FROM check_query_performance() WHERE query_type = 'SELECT') > 1000 THEN 'Optimize queries and add indexes'
            WHEN (SELECT avg_execution_time_ms FROM check_query_performance() WHERE query_type = 'SELECT') > 500 THEN 'Monitor query performance'
            ELSE 'Query performance is good'
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to log usage metrics
CREATE OR REPLACE FUNCTION log_usage_metrics() 
RETURNS VOID AS $$
DECLARE
    db_size DECIMAL;
    conn_count INTEGER;
    db_pct DECIMAL;
    conn_pct DECIMAL;
    db_status TEXT;
    conn_status TEXT;
BEGIN
    -- Get database size metrics
    SELECT total_size_mb, usage_percentage, status
    INTO db_size, db_pct, db_status
    FROM check_database_storage_usage();
    
    -- Get connection metrics
    SELECT current_connections, usage_percentage, status
    INTO conn_count, conn_pct, conn_status
    FROM check_connection_usage();
    
    -- Log database storage usage
    INSERT INTO usage_monitoring (metric_name, metric_value, metric_unit, limit_value, usage_percentage, status, notes)
    VALUES ('Database Storage', db_size, 'MB', 500, db_pct, db_status, 'Free tier storage limit monitoring');
    
    -- Log connection usage
    INSERT INTO usage_monitoring (metric_name, metric_value, metric_unit, limit_value, usage_percentage, status, notes)
    VALUES ('Active Connections', conn_count, 'connections', 60, conn_pct, conn_status, 'Free tier connection limit monitoring');
    
    -- Log timestamp
    INSERT INTO usage_monitoring (metric_name, metric_value, metric_unit, limit_value, usage_percentage, status, notes)
    VALUES ('Monitoring Timestamp', EXTRACT(EPOCH FROM NOW()), 'seconds', 0, 0, 'OK', 'Usage monitoring execution timestamp');
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to get usage trends
CREATE OR REPLACE FUNCTION get_usage_trends(days_back INTEGER DEFAULT 7) 
RETURNS TABLE (
    metric_name TEXT,
    date_recorded DATE,
    avg_usage_percentage DECIMAL,
    max_usage_percentage DECIMAL,
    min_usage_percentage DECIMAL,
    trend_direction TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        um.metric_name,
        um.recorded_at::DATE as date_recorded,
        ROUND(AVG(um.usage_percentage), 2) as avg_usage_percentage,
        ROUND(MAX(um.usage_percentage), 2) as max_usage_percentage,
        ROUND(MIN(um.usage_percentage), 2) as min_usage_percentage,
        CASE 
            WHEN AVG(um.usage_percentage) > LAG(AVG(um.usage_percentage)) OVER (PARTITION BY um.metric_name ORDER BY um.recorded_at::DATE) THEN 'INCREASING'
            WHEN AVG(um.usage_percentage) < LAG(AVG(um.usage_percentage)) OVER (PARTITION BY um.metric_name ORDER BY um.recorded_at::DATE) THEN 'DECREASING'
            ELSE 'STABLE'
        END as trend_direction
    FROM usage_monitoring um
    WHERE um.recorded_at >= NOW() - INTERVAL '1 day' * days_back
    AND um.metric_name IN ('Database Storage', 'Active Connections')
    GROUP BY um.metric_name, um.recorded_at::DATE
    ORDER BY um.metric_name, um.recorded_at::DATE;
END;
$$ LANGUAGE plpgsql;

-- 11. Create a function to check for optimization opportunities
CREATE OR REPLACE FUNCTION check_optimization_opportunities() 
RETURNS TABLE (
    optimization_type TEXT,
    description TEXT,
    potential_savings TEXT,
    priority TEXT,
    action_required TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check for unused indexes
    SELECT 
        'Index Optimization' as optimization_type,
        'Unused indexes consuming storage space' as description,
        (SELECT ROUND(SUM(pg_relation_size(indexname) / (1024 * 1024)), 2)::TEXT || ' MB' 
         FROM check_index_usage() 
         WHERE status = 'UNUSED') as potential_savings,
        'HIGH' as priority,
        'Consider dropping unused indexes to save storage space' as action_required
    WHERE EXISTS (SELECT 1 FROM check_index_usage() WHERE status = 'UNUSED')
    
    UNION ALL
    
    -- Check for large tables
    SELECT 
        'Table Optimization' as optimization_type,
        'Large tables consuming significant storage' as description,
        (SELECT ROUND(MAX(size_mb), 2)::TEXT || ' MB' 
         FROM check_table_sizes() 
         WHERE size_mb > 10) as potential_savings,
        'MEDIUM' as priority,
        'Consider data archiving or compression for large tables' as action_required
    WHERE EXISTS (SELECT 1 FROM check_table_sizes() WHERE size_mb > 10)
    
    UNION ALL
    
    -- Check for slow queries
    SELECT 
        'Query Optimization' as optimization_type,
        'Slow queries affecting performance' as description,
        (SELECT COUNT(*)::TEXT || ' slow queries' 
         FROM check_query_performance() 
         WHERE status IN ('WARNING', 'CRITICAL')) as potential_savings,
        'HIGH' as priority,
        'Optimize slow queries to improve performance' as action_required
    WHERE EXISTS (SELECT 1 FROM check_query_performance() WHERE status IN ('WARNING', 'CRITICAL'));
END;
$$ LANGUAGE plpgsql;

-- 12. Create a function to set up automated monitoring
CREATE OR REPLACE FUNCTION setup_automated_monitoring() 
RETURNS TEXT AS $$
BEGIN
    -- Create a function to be called by cron or scheduled job
    CREATE OR REPLACE FUNCTION automated_usage_monitoring() 
    RETURNS VOID AS $$
    BEGIN
        -- Log current usage metrics
        PERFORM log_usage_metrics();
        
        -- Check for critical usage levels
        IF EXISTS (
            SELECT 1 FROM check_free_tier_limits() 
            WHERE status = 'CRITICAL'
        ) THEN
            -- Log critical alert
            INSERT INTO usage_monitoring (metric_name, metric_value, metric_unit, limit_value, usage_percentage, status, notes)
            VALUES ('CRITICAL ALERT', 0, 'alert', 0, 0, 'CRITICAL', 'Critical usage level detected - immediate attention required');
        END IF;
        
        -- Check for warning levels
        IF EXISTS (
            SELECT 1 FROM check_free_tier_limits() 
            WHERE status = 'WARNING'
        ) THEN
            -- Log warning alert
            INSERT INTO usage_monitoring (metric_name, metric_value, metric_unit, limit_value, usage_percentage, status, notes)
            VALUES ('WARNING ALERT', 0, 'alert', 0, 0, 'WARNING', 'Warning usage level detected - monitor closely');
        END IF;
    END;
    $$ LANGUAGE plpgsql;
    
    RETURN 'Automated monitoring setup completed. Call automated_usage_monitoring() function regularly.';
END;
$$ LANGUAGE plpgsql;

-- 13. Create a function to get monitoring dashboard data
CREATE OR REPLACE FUNCTION get_monitoring_dashboard() 
RETURNS TABLE (
    section TEXT,
    metric TEXT,
    value TEXT,
    status TEXT,
    last_updated TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    -- Current usage metrics
    SELECT 
        'Current Usage' as section,
        'Database Storage' as metric,
        (SELECT total_size_mb::TEXT || ' MB (' || usage_percentage::TEXT || '%)' FROM check_database_storage_usage()) as value,
        (SELECT status FROM check_database_storage_usage()) as status,
        NOW() as last_updated
    
    UNION ALL
    
    SELECT 
        'Current Usage' as section,
        'Active Connections' as metric,
        (SELECT current_connections::TEXT || ' (' || usage_percentage::TEXT || '%)' FROM check_connection_usage()) as value,
        (SELECT status FROM check_connection_usage()) as status,
        NOW() as last_updated
    
    UNION ALL
    
    -- Recent alerts
    SELECT 
        'Recent Alerts' as section,
        um.metric_name as metric,
        um.status as value,
        um.status as status,
        um.recorded_at as last_updated
    FROM usage_monitoring um
    WHERE um.status IN ('WARNING', 'CRITICAL')
    AND um.recorded_at >= NOW() - INTERVAL '24 hours'
    ORDER BY um.recorded_at DESC
    LIMIT 5
    
    UNION ALL
    
    -- Optimization opportunities
    SELECT 
        'Optimization' as section,
        oo.optimization_type as metric,
        oo.potential_savings as value,
        oo.priority as status,
        NOW() as last_updated
    FROM check_optimization_opportunities() oo
    LIMIT 3;
END;
$$ LANGUAGE plpgsql;

-- 14. Create a function to export usage data
CREATE OR REPLACE FUNCTION export_usage_data(days_back INTEGER DEFAULT 30) 
RETURNS TABLE (
    export_date DATE,
    metric_name TEXT,
    daily_avg_usage DECIMAL,
    daily_max_usage DECIMAL,
    daily_min_usage DECIMAL,
    status_summary TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        um.recorded_at::DATE as export_date,
        um.metric_name,
        ROUND(AVG(um.usage_percentage), 2) as daily_avg_usage,
        ROUND(MAX(um.usage_percentage), 2) as daily_max_usage,
        ROUND(MIN(um.usage_percentage), 2) as daily_min_usage,
        CASE 
            WHEN COUNT(CASE WHEN um.status = 'CRITICAL' THEN 1 END) > 0 THEN 'CRITICAL'
            WHEN COUNT(CASE WHEN um.status = 'WARNING' THEN 1 END) > 0 THEN 'WARNING'
            ELSE 'OK'
        END as status_summary
    FROM usage_monitoring um
    WHERE um.recorded_at >= NOW() - INTERVAL '1 day' * days_back
    AND um.metric_name IN ('Database Storage', 'Active Connections')
    GROUP BY um.recorded_at::DATE, um.metric_name
    ORDER BY um.recorded_at::DATE DESC, um.metric_name;
END;
$$ LANGUAGE plpgsql;

-- 15. Create a function to validate monitoring setup
CREATE OR REPLACE FUNCTION validate_monitoring_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if usage_monitoring table exists
    SELECT 
        'Usage Monitoring Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'usage_monitoring') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing usage metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'usage_monitoring') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create usage_monitoring table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if monitoring functions exist
    SELECT 
        'Monitoring Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'check_database_storage_usage') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for checking usage metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'check_database_storage_usage') 
            THEN 'All monitoring functions are available' 
            ELSE 'Create monitoring functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if automated monitoring is set up
    SELECT 
        'Automated Monitoring' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'automated_usage_monitoring') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Automated monitoring function' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'automated_usage_monitoring') 
            THEN 'Automated monitoring is ready' 
            ELSE 'Set up automated monitoring' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_usage_monitoring_metric_name ON usage_monitoring(metric_name);
CREATE INDEX IF NOT EXISTS idx_usage_monitoring_recorded_at ON usage_monitoring(recorded_at);
CREATE INDEX IF NOT EXISTS idx_usage_monitoring_status ON usage_monitoring(status);

-- Create a view for easy monitoring access
CREATE OR REPLACE VIEW monitoring_dashboard AS
SELECT 
    'Database Storage' as metric_name,
    (SELECT total_size_mb::TEXT || ' MB' FROM check_database_storage_usage()) as current_value,
    '500 MB' as limit_value,
    (SELECT usage_percentage::TEXT || '%' FROM check_database_storage_usage()) as usage_percentage,
    (SELECT status FROM check_database_storage_usage()) as status
UNION ALL
SELECT 
    'Active Connections' as metric_name,
    (SELECT current_connections::TEXT FROM check_connection_usage()) as current_value,
    '60' as limit_value,
    (SELECT usage_percentage::TEXT || '%' FROM check_connection_usage()) as usage_percentage,
    (SELECT status FROM check_connection_usage()) as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON monitoring_dashboard TO authenticated;
GRANT SELECT, INSERT ON usage_monitoring TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Supabase usage monitoring setup completed successfully!';
    RAISE NOTICE 'Total functions created: 15';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 3';
    RAISE NOTICE 'All usage monitoring tools are now available.';
    RAISE NOTICE 'Use monitoring_dashboard view to access current usage metrics.';
    RAISE NOTICE 'Call setup_automated_monitoring() to enable automated monitoring.';
    RAISE NOTICE 'Call log_usage_metrics() to record current usage metrics.';
END $$;
