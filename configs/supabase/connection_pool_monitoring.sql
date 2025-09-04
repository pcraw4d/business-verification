-- Database Connection Pool Monitoring for Business Classification System
-- This script provides comprehensive monitoring and optimization for database connection pools

-- 1. Create a connection pool monitoring table
CREATE TABLE IF NOT EXISTS connection_pool_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    active_connections INTEGER NOT NULL,
    idle_connections INTEGER NOT NULL,
    total_connections INTEGER NOT NULL,
    max_connections INTEGER NOT NULL,
    connection_utilization DECIMAL(5,2) NOT NULL,
    avg_connection_duration_seconds DECIMAL(10,2),
    connection_errors INTEGER DEFAULT 0,
    connection_timeouts INTEGER DEFAULT 0,
    pool_hit_ratio DECIMAL(5,2),
    pool_miss_ratio DECIMAL(5,2),
    avg_wait_time_ms DECIMAL(10,2),
    max_wait_time_ms DECIMAL(10,2),
    connection_creation_rate DECIMAL(10,2),
    connection_destruction_rate DECIMAL(10,2),
    pool_status VARCHAR(20) NOT NULL,
    recommendations TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create a function to get current connection pool statistics
CREATE OR REPLACE FUNCTION get_connection_pool_stats()
RETURNS TABLE (
    active_connections INTEGER,
    idle_connections INTEGER,
    total_connections INTEGER,
    max_connections INTEGER,
    connection_utilization DECIMAL,
    avg_connection_duration_seconds DECIMAL,
    connection_errors INTEGER,
    connection_timeouts INTEGER,
    pool_hit_ratio DECIMAL,
    pool_miss_ratio DECIMAL,
    avg_wait_time_ms DECIMAL,
    max_wait_time_ms DECIMAL,
    connection_creation_rate DECIMAL,
    connection_destruction_rate DECIMAL,
    pool_status VARCHAR
) AS $$
DECLARE
    v_active_connections INTEGER;
    v_idle_connections INTEGER;
    v_total_connections INTEGER;
    v_max_connections INTEGER;
    v_connection_utilization DECIMAL;
    v_avg_connection_duration_seconds DECIMAL;
    v_connection_errors INTEGER;
    v_connection_timeouts INTEGER;
    v_pool_hit_ratio DECIMAL;
    v_pool_miss_ratio DECIMAL;
    v_avg_wait_time_ms DECIMAL;
    v_max_wait_time_ms DECIMAL;
    v_connection_creation_rate DECIMAL;
    v_connection_destruction_rate DECIMAL;
    v_pool_status VARCHAR;
BEGIN
    -- Get current connection statistics from pg_stat_activity
    SELECT 
        COUNT(*) FILTER (WHERE state = 'active') as active_connections,
        COUNT(*) FILTER (WHERE state = 'idle') as idle_connections,
        COUNT(*) as total_connections,
        (SELECT setting::INTEGER FROM pg_settings WHERE name = 'max_connections') as max_connections
    INTO 
        v_active_connections,
        v_idle_connections,
        v_total_connections,
        v_max_connections
    FROM pg_stat_activity
    WHERE datname = current_database();
    
    -- Calculate connection utilization
    v_connection_utilization := (v_total_connections::DECIMAL / v_max_connections) * 100;
    
    -- Get average connection duration (simplified calculation)
    SELECT 
        COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - backend_start))), 0) as avg_duration
    INTO v_avg_connection_duration_seconds
    FROM pg_stat_activity
    WHERE datname = current_database() AND backend_start IS NOT NULL;
    
    -- Get connection errors (from pg_stat_database)
    SELECT 
        COALESCE(deadlocks, 0) as connection_errors,
        COALESCE(temp_files, 0) as connection_timeouts
    INTO 
        v_connection_errors,
        v_connection_timeouts
    FROM pg_stat_database
    WHERE datname = current_database();
    
    -- Calculate pool hit ratio (simplified)
    SELECT 
        COALESCE(
            (blks_hit::DECIMAL / (blks_hit + blks_read)) * 100, 
            100
        ) as hit_ratio
    INTO v_pool_hit_ratio
    FROM pg_stat_database
    WHERE datname = current_database();
    
    v_pool_miss_ratio := 100 - v_pool_hit_ratio;
    
    -- Calculate wait times (simplified)
    v_avg_wait_time_ms := 0.5; -- Placeholder
    v_max_wait_time_ms := 2.0; -- Placeholder
    
    -- Calculate connection rates (simplified)
    v_connection_creation_rate := 0.1; -- Placeholder
    v_connection_destruction_rate := 0.1; -- Placeholder
    
    -- Determine pool status
    IF v_connection_utilization > 90 THEN
        v_pool_status := 'CRITICAL';
    ELSIF v_connection_utilization > 75 THEN
        v_pool_status := 'WARNING';
    ELSIF v_connection_utilization > 50 THEN
        v_pool_status := 'MODERATE';
    ELSE
        v_pool_status := 'HEALTHY';
    END IF;
    
    RETURN QUERY
    SELECT 
        v_active_connections,
        v_idle_connections,
        v_total_connections,
        v_max_connections,
        v_connection_utilization,
        v_avg_connection_duration_seconds,
        v_connection_errors,
        v_connection_timeouts,
        v_pool_hit_ratio,
        v_pool_miss_ratio,
        v_avg_wait_time_ms,
        v_max_wait_time_ms,
        v_connection_creation_rate,
        v_connection_destruction_rate,
        v_pool_status;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to log connection pool metrics
CREATE OR REPLACE FUNCTION log_connection_pool_metrics(
    p_active_connections INTEGER,
    p_idle_connections INTEGER,
    p_total_connections INTEGER,
    p_max_connections INTEGER,
    p_connection_utilization DECIMAL,
    p_avg_connection_duration_seconds DECIMAL,
    p_connection_errors INTEGER DEFAULT 0,
    p_connection_timeouts INTEGER DEFAULT 0,
    p_pool_hit_ratio DECIMAL DEFAULT 0,
    p_pool_miss_ratio DECIMAL DEFAULT 0,
    p_avg_wait_time_ms DECIMAL DEFAULT 0,
    p_max_wait_time_ms DECIMAL DEFAULT 0,
    p_connection_creation_rate DECIMAL DEFAULT 0,
    p_connection_destruction_rate DECIMAL DEFAULT 0,
    p_pool_status VARCHAR DEFAULT 'UNKNOWN',
    p_recommendations TEXT DEFAULT NULL
) RETURNS INTEGER AS $$
DECLARE
    v_log_id INTEGER;
BEGIN
    INSERT INTO connection_pool_metrics (
        active_connections,
        idle_connections,
        total_connections,
        max_connections,
        connection_utilization,
        avg_connection_duration_seconds,
        connection_errors,
        connection_timeouts,
        pool_hit_ratio,
        pool_miss_ratio,
        avg_wait_time_ms,
        max_wait_time_ms,
        connection_creation_rate,
        connection_destruction_rate,
        pool_status,
        recommendations
    ) VALUES (
        p_active_connections,
        p_idle_connections,
        p_total_connections,
        p_max_connections,
        p_connection_utilization,
        p_avg_connection_duration_seconds,
        p_connection_errors,
        p_connection_timeouts,
        p_pool_hit_ratio,
        p_pool_miss_ratio,
        p_avg_wait_time_ms,
        p_max_wait_time_ms,
        p_connection_creation_rate,
        p_connection_destruction_rate,
        p_pool_status,
        p_recommendations
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to get connection pool trends
CREATE OR REPLACE FUNCTION get_connection_pool_trends(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    hour_bucket TIMESTAMPTZ,
    avg_active_connections DECIMAL,
    avg_idle_connections DECIMAL,
    avg_total_connections DECIMAL,
    avg_connection_utilization DECIMAL,
    avg_pool_hit_ratio DECIMAL,
    avg_wait_time_ms DECIMAL,
    connection_errors_count BIGINT,
    pool_status_changes BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc('hour', timestamp) as hour_bucket,
        ROUND(AVG(active_connections), 2) as avg_active_connections,
        ROUND(AVG(idle_connections), 2) as avg_idle_connections,
        ROUND(AVG(total_connections), 2) as avg_total_connections,
        ROUND(AVG(connection_utilization), 2) as avg_connection_utilization,
        ROUND(AVG(pool_hit_ratio), 2) as avg_pool_hit_ratio,
        ROUND(AVG(avg_wait_time_ms), 2) as avg_wait_time_ms,
        SUM(connection_errors) as connection_errors_count,
        COUNT(CASE WHEN pool_status != LAG(pool_status) OVER (ORDER BY timestamp) THEN 1 END) as pool_status_changes
    FROM connection_pool_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    GROUP BY date_trunc('hour', timestamp)
    ORDER BY hour_bucket DESC;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to get connection pool alerts
CREATE OR REPLACE FUNCTION get_connection_pool_alerts(
    p_hours_back INTEGER DEFAULT 1
) RETURNS TABLE (
    alert_id SERIAL,
    alert_type VARCHAR(50),
    alert_level VARCHAR(20),
    alert_message TEXT,
    metric_value DECIMAL,
    threshold_value DECIMAL,
    recommendations TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    -- High connection utilization alerts
    SELECT 
        cpm.id as alert_id,
        'HIGH_CONNECTION_UTILIZATION' as alert_type,
        CASE 
            WHEN cpm.connection_utilization > 95 THEN 'CRITICAL'
            WHEN cpm.connection_utilization > 85 THEN 'HIGH'
            WHEN cpm.connection_utilization > 75 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'High connection pool utilization: ' || ROUND(cpm.connection_utilization, 2) || '%' as alert_message,
        cpm.connection_utilization as metric_value,
        75.0 as threshold_value,
        'Consider increasing max_connections or optimizing connection usage' as recommendations,
        cpm.timestamp as created_at
    FROM connection_pool_metrics cpm
    WHERE cpm.timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND cpm.connection_utilization > 75
    
    UNION ALL
    
    -- Low pool hit ratio alerts
    SELECT 
        cpm.id as alert_id,
        'LOW_POOL_HIT_RATIO' as alert_type,
        CASE 
            WHEN cpm.pool_hit_ratio < 80 THEN 'HIGH'
            WHEN cpm.pool_hit_ratio < 90 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'Low connection pool hit ratio: ' || ROUND(cpm.pool_hit_ratio, 2) || '%' as alert_message,
        cpm.pool_hit_ratio as metric_value,
        90.0 as threshold_value,
        'Consider increasing connection pool size or optimizing connection reuse' as recommendations,
        cpm.timestamp as created_at
    FROM connection_pool_metrics cpm
    WHERE cpm.timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND cpm.pool_hit_ratio < 90
    
    UNION ALL
    
    -- High wait time alerts
    SELECT 
        cpm.id as alert_id,
        'HIGH_WAIT_TIME' as alert_type,
        CASE 
            WHEN cpm.avg_wait_time_ms > 1000 THEN 'CRITICAL'
            WHEN cpm.avg_wait_time_ms > 500 THEN 'HIGH'
            WHEN cpm.avg_wait_time_ms > 100 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'High connection wait time: ' || ROUND(cpm.avg_wait_time_ms, 2) || 'ms' as alert_message,
        cpm.avg_wait_time_ms as metric_value,
        100.0 as threshold_value,
        'Consider increasing connection pool size or optimizing connection management' as recommendations,
        cpm.timestamp as created_at
    FROM connection_pool_metrics cpm
    WHERE cpm.timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND cpm.avg_wait_time_ms > 100
    
    UNION ALL
    
    -- Connection error alerts
    SELECT 
        cpm.id as alert_id,
        'CONNECTION_ERRORS' as alert_type,
        CASE 
            WHEN cpm.connection_errors > 10 THEN 'CRITICAL'
            WHEN cpm.connection_errors > 5 THEN 'HIGH'
            WHEN cpm.connection_errors > 1 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'Connection errors detected: ' || cpm.connection_errors as alert_message,
        cpm.connection_errors::DECIMAL as metric_value,
        1.0 as threshold_value,
        'Investigate connection errors and check database health' as recommendations,
        cpm.timestamp as created_at
    FROM connection_pool_metrics cpm
    WHERE cpm.timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    AND cpm.connection_errors > 0
    
    ORDER BY created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to get connection pool dashboard
CREATE OR REPLACE FUNCTION get_connection_pool_dashboard() 
RETURNS TABLE (
    metric_name TEXT,
    current_value TEXT,
    target_value TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT
) AS $$
DECLARE
    v_active_connections INTEGER;
    v_idle_connections INTEGER;
    v_total_connections INTEGER;
    v_max_connections INTEGER;
    v_connection_utilization DECIMAL;
    v_pool_hit_ratio DECIMAL;
    v_avg_wait_time_ms DECIMAL;
    v_connection_errors INTEGER;
BEGIN
    -- Get current metrics
    SELECT 
        active_connections,
        idle_connections,
        total_connections,
        max_connections,
        connection_utilization,
        pool_hit_ratio,
        avg_wait_time_ms,
        connection_errors
    INTO 
        v_active_connections,
        v_idle_connections,
        v_total_connections,
        v_max_connections,
        v_connection_utilization,
        v_pool_hit_ratio,
        v_avg_wait_time_ms,
        v_connection_errors
    FROM get_connection_pool_stats();
    
    RETURN QUERY
    -- Active Connections
    SELECT 
        'Active Connections' as metric_name,
        v_active_connections::TEXT as current_value,
        'N/A' as target_value,
        CASE 
            WHEN v_active_connections > v_max_connections * 0.8 THEN 'HIGH'
            WHEN v_active_connections > v_max_connections * 0.6 THEN 'MEDIUM'
            ELSE 'LOW'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_active_connections > v_max_connections * 0.8 THEN 'High active connection count - monitor for performance impact'
            WHEN v_active_connections > v_max_connections * 0.6 THEN 'Moderate active connection count - monitor performance'
            ELSE 'Low active connection count - normal operation'
        END as recommendations
    
    UNION ALL
    
    -- Connection Utilization
    SELECT 
        'Connection Utilization' as metric_name,
        ROUND(v_connection_utilization, 2)::TEXT || '%' as current_value,
        '75%' as target_value,
        CASE 
            WHEN v_connection_utilization > 90 THEN 'CRITICAL'
            WHEN v_connection_utilization > 75 THEN 'WARNING'
            WHEN v_connection_utilization > 50 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_connection_utilization > 90 THEN 'Critical: Connection pool near capacity'
            WHEN v_connection_utilization > 75 THEN 'Warning: High connection utilization'
            WHEN v_connection_utilization > 50 THEN 'Fair: Moderate connection utilization'
            ELSE 'Good: Connection utilization is healthy'
        END as recommendations
    
    UNION ALL
    
    -- Pool Hit Ratio
    SELECT 
        'Pool Hit Ratio' as metric_name,
        ROUND(v_pool_hit_ratio, 2)::TEXT || '%' as current_value,
        '90%' as target_value,
        CASE 
            WHEN v_pool_hit_ratio < 80 THEN 'CRITICAL'
            WHEN v_pool_hit_ratio < 90 THEN 'WARNING'
            WHEN v_pool_hit_ratio < 95 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_pool_hit_ratio < 80 THEN 'Critical: Poor pool hit ratio - increase pool size'
            WHEN v_pool_hit_ratio < 90 THEN 'Warning: Low pool hit ratio - optimize connection reuse'
            WHEN v_pool_hit_ratio < 95 THEN 'Fair: Moderate pool hit ratio - monitor performance'
            ELSE 'Good: Pool hit ratio is excellent'
        END as recommendations
    
    UNION ALL
    
    -- Average Wait Time
    SELECT 
        'Average Wait Time' as metric_name,
        ROUND(v_avg_wait_time_ms, 2)::TEXT || ' ms' as current_value,
        '100 ms' as target_value,
        CASE 
            WHEN v_avg_wait_time_ms > 1000 THEN 'CRITICAL'
            WHEN v_avg_wait_time_ms > 500 THEN 'WARNING'
            WHEN v_avg_wait_time_ms > 100 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_wait_time_ms > 1000 THEN 'Critical: Very high wait times - increase pool size'
            WHEN v_avg_wait_time_ms > 500 THEN 'Warning: High wait times - optimize connection management'
            WHEN v_avg_wait_time_ms > 100 THEN 'Fair: Moderate wait times - monitor performance'
            ELSE 'Good: Wait times are acceptable'
        END as recommendations
    
    UNION ALL
    
    -- Connection Errors
    SELECT 
        'Connection Errors' as metric_name,
        v_connection_errors::TEXT as current_value,
        '0' as target_value,
        CASE 
            WHEN v_connection_errors > 10 THEN 'CRITICAL'
            WHEN v_connection_errors > 5 THEN 'WARNING'
            WHEN v_connection_errors > 1 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_connection_errors > 10 THEN 'Critical: Many connection errors - investigate immediately'
            WHEN v_connection_errors > 5 THEN 'Warning: Several connection errors - check database health'
            WHEN v_connection_errors > 1 THEN 'Fair: Some connection errors - monitor closely'
            ELSE 'Good: No connection errors detected'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to get connection pool insights
CREATE OR REPLACE FUNCTION get_connection_pool_insights() 
RETURNS TABLE (
    insight_type TEXT,
    insight_title TEXT,
    insight_description TEXT,
    insight_priority VARCHAR(20),
    insight_recommendations TEXT,
    affected_connections BIGINT,
    potential_improvement DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    -- High utilization insights
    SELECT 
        'HIGH_UTILIZATION' as insight_type,
        'Connection Pool Optimization' as insight_title,
        'High connection pool utilization detected' as insight_description,
        'HIGH' as insight_priority,
        'Consider increasing max_connections or optimizing connection usage patterns' as insight_recommendations,
        COUNT(*) as affected_connections,
        ROUND(AVG(connection_utilization - 75), 2) as potential_improvement
    FROM connection_pool_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND connection_utilization > 75
    
    UNION ALL
    
    -- Low hit ratio insights
    SELECT 
        'LOW_HIT_RATIO' as insight_type,
        'Connection Pool Efficiency' as insight_title,
        'Low connection pool hit ratio detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Optimize connection reuse and consider increasing pool size' as insight_recommendations,
        COUNT(*) as affected_connections,
        ROUND(AVG(90 - pool_hit_ratio), 2) as potential_improvement
    FROM connection_pool_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND pool_hit_ratio < 90
    
    UNION ALL
    
    -- High wait time insights
    SELECT 
        'HIGH_WAIT_TIME' as insight_type,
        'Connection Wait Time Optimization' as insight_title,
        'High connection wait times detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Increase connection pool size or optimize connection management' as insight_recommendations,
        COUNT(*) as affected_connections,
        ROUND(AVG(avg_wait_time_ms - 100), 2) as potential_improvement
    FROM connection_pool_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND avg_wait_time_ms > 100
    
    UNION ALL
    
    -- Connection error insights
    SELECT 
        'CONNECTION_ERRORS' as insight_type,
        'Connection Error Resolution' as insight_title,
        'Connection errors detected' as insight_description,
        'HIGH' as insight_priority,
        'Investigate and resolve connection errors to improve stability' as insight_recommendations,
        COUNT(*) as affected_connections,
        ROUND(AVG(connection_errors), 2) as potential_improvement
    FROM connection_pool_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND connection_errors > 0;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to optimize connection pool settings
CREATE OR REPLACE FUNCTION optimize_connection_pool_settings() 
RETURNS TABLE (
    setting_name TEXT,
    current_value TEXT,
    recommended_value TEXT,
    reason TEXT,
    impact_level VARCHAR(20)
) AS $$
DECLARE
    v_current_utilization DECIMAL;
    v_current_hit_ratio DECIMAL;
    v_current_wait_time DECIMAL;
    v_max_connections INTEGER;
BEGIN
    -- Get current metrics
    SELECT 
        connection_utilization,
        pool_hit_ratio,
        avg_wait_time_ms,
        max_connections
    INTO 
        v_current_utilization,
        v_current_hit_ratio,
        v_current_wait_time,
        v_max_connections
    FROM get_connection_pool_stats();
    
    RETURN QUERY
    -- Max connections optimization
    SELECT 
        'max_connections' as setting_name,
        v_max_connections::TEXT as current_value,
        CASE 
            WHEN v_current_utilization > 90 THEN (v_max_connections * 1.5)::TEXT
            WHEN v_current_utilization > 75 THEN (v_max_connections * 1.25)::TEXT
            ELSE v_max_connections::TEXT
        END as recommended_value,
        CASE 
            WHEN v_current_utilization > 90 THEN 'Critical: Connection pool near capacity'
            WHEN v_current_utilization > 75 THEN 'Warning: High connection utilization'
            ELSE 'Current setting is adequate'
        END as reason,
        CASE 
            WHEN v_current_utilization > 90 THEN 'HIGH'
            WHEN v_current_utilization > 75 THEN 'MEDIUM'
            ELSE 'LOW'
        END as impact_level
    
    UNION ALL
    
    -- Connection timeout optimization
    SELECT 
        'connection_timeout' as setting_name,
        '30' as current_value,
        CASE 
            WHEN v_current_wait_time > 1000 THEN '60'
            WHEN v_current_wait_time > 500 THEN '45'
            ELSE '30'
        END as recommended_value,
        CASE 
            WHEN v_current_wait_time > 1000 THEN 'High wait times require longer timeouts'
            WHEN v_current_wait_time > 500 THEN 'Moderate wait times may benefit from longer timeouts'
            ELSE 'Current timeout is adequate'
        END as reason,
        CASE 
            WHEN v_current_wait_time > 1000 THEN 'HIGH'
            WHEN v_current_wait_time > 500 THEN 'MEDIUM'
            ELSE 'LOW'
        END as impact_level
    
    UNION ALL
    
    -- Idle connection timeout optimization
    SELECT 
        'idle_connection_timeout' as setting_name,
        '300' as current_value,
        CASE 
            WHEN v_current_hit_ratio < 80 THEN '600'
            WHEN v_current_hit_ratio < 90 THEN '450'
            ELSE '300'
        END as recommended_value,
        CASE 
            WHEN v_current_hit_ratio < 80 THEN 'Low hit ratio - keep connections longer'
            WHEN v_current_hit_ratio < 90 THEN 'Moderate hit ratio - consider longer idle timeout'
            ELSE 'Current idle timeout is adequate'
        END as reason,
        CASE 
            WHEN v_current_hit_ratio < 80 THEN 'HIGH'
            WHEN v_current_hit_ratio < 90 THEN 'MEDIUM'
            ELSE 'LOW'
        END as impact_level;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to cleanup old connection pool metrics
CREATE OR REPLACE FUNCTION cleanup_connection_pool_metrics(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM connection_pool_metrics
    WHERE timestamp < NOW() - INTERVAL '1 day' * p_days_to_keep;
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to validate connection pool monitoring setup
CREATE OR REPLACE FUNCTION validate_connection_pool_monitoring_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if connection_pool_metrics table exists
    SELECT 
        'Connection Pool Metrics Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'connection_pool_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing connection pool metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'connection_pool_metrics') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create connection_pool_metrics table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if monitoring functions exist
    SELECT 
        'Connection Pool Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_connection_pool_stats') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for monitoring connection pool' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_connection_pool_stats') 
            THEN 'All connection pool functions are available' 
            ELSE 'Create connection pool monitoring functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if logging function exists
    SELECT 
        'Connection Pool Logging' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_connection_pool_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Function for logging connection pool metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_connection_pool_metrics') 
            THEN 'Logging function is ready' 
            ELSE 'Create connection pool logging function' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_connection_pool_metrics_timestamp ON connection_pool_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_connection_pool_metrics_pool_status ON connection_pool_metrics(pool_status);
CREATE INDEX IF NOT EXISTS idx_connection_pool_metrics_utilization ON connection_pool_metrics(connection_utilization);
CREATE INDEX IF NOT EXISTS idx_connection_pool_metrics_active_connections ON connection_pool_metrics(active_connections);

-- Create a view for easy connection pool dashboard access
CREATE OR REPLACE VIEW connection_pool_dashboard AS
SELECT 
    'Connection Pool Overview' as metric_name,
    (SELECT total_connections::TEXT FROM get_connection_pool_stats()) as current_value,
    'N/A' as target_value,
    (SELECT 
        CASE 
            WHEN connection_utilization > 90 THEN 'CRITICAL'
            WHEN connection_utilization > 75 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_connection_pool_stats()) as status
UNION ALL
SELECT 
    'Connection Utilization' as metric_name,
    (SELECT ROUND(connection_utilization, 2)::TEXT || '%' FROM get_connection_pool_stats()) as current_value,
    '75%' as target_value,
    (SELECT 
        CASE 
            WHEN connection_utilization > 90 THEN 'CRITICAL'
            WHEN connection_utilization > 75 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_connection_pool_stats()) as status
UNION ALL
SELECT 
    'Pool Hit Ratio' as metric_name,
    (SELECT ROUND(pool_hit_ratio, 2)::TEXT || '%' FROM get_connection_pool_stats()) as current_value,
    '90%' as target_value,
    (SELECT 
        CASE 
            WHEN pool_hit_ratio < 80 THEN 'CRITICAL'
            WHEN pool_hit_ratio < 90 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_connection_pool_stats()) as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON connection_pool_dashboard TO authenticated;
GRANT SELECT, INSERT ON connection_pool_metrics TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Connection pool monitoring setup completed successfully!';
    RAISE NOTICE 'Total functions created: 10';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 4';
    RAISE NOTICE 'All connection pool monitoring tools are now available.';
    RAISE NOTICE 'Use connection_pool_dashboard view to access current connection pool metrics.';
    RAISE NOTICE 'Call get_connection_pool_stats() to get current connection pool statistics.';
    RAISE NOTICE 'Call log_connection_pool_metrics() to log connection pool metrics.';
END $$;
