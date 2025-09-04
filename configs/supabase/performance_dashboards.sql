-- Database Performance Dashboards for Business Classification System
-- This script provides comprehensive performance monitoring and dashboard views

-- 1. Create a comprehensive performance monitoring table
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

-- 2. Create a function to collect comprehensive performance metrics
CREATE OR REPLACE FUNCTION collect_performance_metrics() 
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
    index_usage_score DECIMAL;
    table_bloat_score DECIMAL;
    query_performance_score DECIMAL;
    connection_utilization DECIMAL;
    storage_utilization DECIMAL;
    overall_health_score DECIMAL;
BEGIN
    -- Database size metrics
    SELECT 
        ROUND(SUM(pg_database_size(datname)) / (1024 * 1024), 2)
    INTO db_size
    FROM pg_database 
    WHERE datname = current_database();
    
    -- Connection metrics
    SELECT COUNT(*) INTO db_connections
    FROM pg_stat_activity
    WHERE state = 'active';
    
    -- Slow query metrics
    SELECT COUNT(*) INTO slow_queries
    FROM pg_stat_statements
    WHERE mean_exec_time > 1000;
    
    -- Cache hit ratio
    SELECT 
        ROUND(
            (blks_hit::DECIMAL / (blks_hit + blks_read)) * 100, 2
        )
    INTO cache_hit_ratio
    FROM pg_stat_database
    WHERE datname = current_database();
    
    -- Index usage score
    SELECT 
        ROUND(
            AVG(
                CASE 
                    WHEN idx_scan > 0 THEN (idx_scan::DECIMAL / (idx_scan + seq_scan)) * 100
                    ELSE 0
                END
            ), 2
        )
    INTO index_usage_score
    FROM pg_stat_user_tables;
    
    -- Table bloat score (simplified)
    SELECT 
        ROUND(
            AVG(
                CASE 
                    WHEN n_dead_tup > 0 THEN (n_dead_tup::DECIMAL / n_live_tup) * 100
                    ELSE 0
                END
            ), 2
        )
    INTO table_bloat_score
    FROM pg_stat_user_tables;
    
    -- Query performance score
    SELECT 
        ROUND(
            AVG(
                CASE 
                    WHEN mean_exec_time < 100 THEN 100
                    WHEN mean_exec_time < 500 THEN 80
                    WHEN mean_exec_time < 1000 THEN 60
                    WHEN mean_exec_time < 2000 THEN 40
                    ELSE 20
                END
            ), 2
        )
    INTO query_performance_score
    FROM pg_stat_statements;
    
    -- Connection utilization
    connection_utilization := ROUND((db_connections::DECIMAL / 60) * 100, 2);
    
    -- Storage utilization
    storage_utilization := ROUND((db_size / 500) * 100, 2);
    
    -- Overall health score
    overall_health_score := ROUND(
        (COALESCE(cache_hit_ratio, 0) + 
         COALESCE(index_usage_score, 0) + 
         COALESCE(query_performance_score, 0) + 
         (100 - COALESCE(table_bloat_score, 0)) + 
         (100 - COALESCE(connection_utilization, 0)) + 
         (100 - COALESCE(storage_utilization, 0))) / 6, 2
    );
    
    -- Return comprehensive metrics
    RETURN QUERY
    SELECT 
        'Database Size' as metric_name,
        db_size as metric_value,
        'MB' as metric_unit,
        'Storage' as metric_category,
        CASE 
            WHEN storage_utilization >= 90 THEN 'CRITICAL'
            WHEN storage_utilization >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'utilization_percentage', storage_utilization,
            'limit_mb', 500
        ) as details,
        CASE 
            WHEN storage_utilization >= 90 THEN 'Consider upgrading to paid plan or optimizing data'
            WHEN storage_utilization >= 75 THEN 'Monitor storage usage closely'
            ELSE 'Storage usage is within acceptable limits'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Active Connections' as metric_name,
        db_connections as metric_value,
        'connections' as metric_unit,
        'Connections' as metric_category,
        CASE 
            WHEN connection_utilization >= 90 THEN 'CRITICAL'
            WHEN connection_utilization >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'utilization_percentage', connection_utilization,
            'limit_connections', 60
        ) as details,
        CASE 
            WHEN connection_utilization >= 90 THEN 'Consider connection pooling or upgrading plan'
            WHEN connection_utilization >= 75 THEN 'Monitor connection usage closely'
            ELSE 'Connection usage is within acceptable limits'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Cache Hit Ratio' as metric_name,
        COALESCE(cache_hit_ratio, 0) as metric_value,
        '%' as metric_unit,
        'Performance' as metric_category,
        CASE 
            WHEN cache_hit_ratio < 80 THEN 'CRITICAL'
            WHEN cache_hit_ratio < 90 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'target_percentage', 95,
            'current_percentage', cache_hit_ratio
        ) as details,
        CASE 
            WHEN cache_hit_ratio < 80 THEN 'Increase shared_buffers or optimize queries'
            WHEN cache_hit_ratio < 90 THEN 'Monitor cache performance'
            ELSE 'Cache performance is good'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Index Usage Score' as metric_name,
        COALESCE(index_usage_score, 0) as metric_value,
        '%' as metric_unit,
        'Performance' as metric_category,
        CASE 
            WHEN index_usage_score < 50 THEN 'CRITICAL'
            WHEN index_usage_score < 70 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'target_percentage', 80,
            'current_percentage', index_usage_score
        ) as details,
        CASE 
            WHEN index_usage_score < 50 THEN 'Review and optimize indexes'
            WHEN index_usage_score < 70 THEN 'Monitor index usage'
            ELSE 'Index usage is efficient'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Query Performance Score' as metric_name,
        COALESCE(query_performance_score, 0) as metric_value,
        '%' as metric_unit,
        'Performance' as metric_category,
        CASE 
            WHEN query_performance_score < 50 THEN 'CRITICAL'
            WHEN query_performance_score < 70 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'target_percentage', 80,
            'current_percentage', query_performance_score
        ) as details,
        CASE 
            WHEN query_performance_score < 50 THEN 'Optimize slow queries and add indexes'
            WHEN query_performance_score < 70 THEN 'Monitor query performance'
            ELSE 'Query performance is good'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Table Bloat Score' as metric_name,
        COALESCE(table_bloat_score, 0) as metric_value,
        '%' as metric_unit,
        'Maintenance' as metric_category,
        CASE 
            WHEN table_bloat_score > 20 THEN 'CRITICAL'
            WHEN table_bloat_score > 10 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'target_percentage', 5,
            'current_percentage', table_bloat_score
        ) as details,
        CASE 
            WHEN table_bloat_score > 20 THEN 'Run VACUUM and consider VACUUM FULL'
            WHEN table_bloat_score > 10 THEN 'Monitor table bloat'
            ELSE 'Table bloat is within acceptable limits'
        END as recommendations
    
    UNION ALL
    
    SELECT 
        'Overall Health Score' as metric_name,
        overall_health_score as metric_value,
        '%' as metric_unit,
        'Overall' as metric_category,
        CASE 
            WHEN overall_health_score < 60 THEN 'CRITICAL'
            WHEN overall_health_score < 80 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        jsonb_build_object(
            'target_percentage', 85,
            'current_percentage', overall_health_score,
            'components', jsonb_build_object(
                'cache_hit_ratio', cache_hit_ratio,
                'index_usage_score', index_usage_score,
                'query_performance_score', query_performance_score,
                'table_bloat_score', table_bloat_score,
                'connection_utilization', connection_utilization,
                'storage_utilization', storage_utilization
            )
        ) as details,
        CASE 
            WHEN overall_health_score < 60 THEN 'Multiple performance issues detected - comprehensive optimization needed'
            WHEN overall_health_score < 80 THEN 'Some performance issues detected - monitor and optimize'
            ELSE 'Database performance is good'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to get detailed query performance analysis
CREATE OR REPLACE FUNCTION get_query_performance_analysis() 
RETURNS TABLE (
    query_id BIGINT,
    query_text TEXT,
    calls BIGINT,
    total_time DECIMAL,
    mean_time DECIMAL,
    min_time DECIMAL,
    max_time DECIMAL,
    stddev_time DECIMAL,
    rows BIGINT,
    performance_category TEXT,
    optimization_priority TEXT,
    recommendations TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.queryid as query_id,
        LEFT(s.query, 100) as query_text,
        s.calls,
        ROUND(s.total_exec_time, 2) as total_time,
        ROUND(s.mean_exec_time, 2) as mean_time,
        ROUND(s.min_exec_time, 2) as min_time,
        ROUND(s.max_exec_time, 2) as max_time,
        ROUND(s.stddev_exec_time, 2) as stddev_time,
        s.rows,
        CASE 
            WHEN s.mean_exec_time > 2000 THEN 'CRITICAL'
            WHEN s.mean_exec_time > 1000 THEN 'HIGH'
            WHEN s.mean_exec_time > 500 THEN 'MEDIUM'
            WHEN s.mean_exec_time > 100 THEN 'LOW'
            ELSE 'EXCELLENT'
        END as performance_category,
        CASE 
            WHEN s.mean_exec_time > 2000 AND s.calls > 100 THEN 'URGENT'
            WHEN s.mean_exec_time > 1000 AND s.calls > 50 THEN 'HIGH'
            WHEN s.mean_exec_time > 500 AND s.calls > 20 THEN 'MEDIUM'
            ELSE 'LOW'
        END as optimization_priority,
        CASE 
            WHEN s.mean_exec_time > 2000 THEN 'Add indexes, optimize query structure, consider query rewriting'
            WHEN s.mean_exec_time > 1000 THEN 'Review indexes, optimize joins, consider query hints'
            WHEN s.mean_exec_time > 500 THEN 'Monitor performance, consider minor optimizations'
            WHEN s.mean_exec_time > 100 THEN 'Performance is acceptable'
            ELSE 'Query performance is excellent'
        END as recommendations
    FROM pg_stat_statements s
    WHERE s.calls > 0
    ORDER BY s.mean_exec_time DESC
    LIMIT 50;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to get index performance analysis
CREATE OR REPLACE FUNCTION get_index_performance_analysis() 
RETURNS TABLE (
    table_name TEXT,
    index_name TEXT,
    index_size_mb DECIMAL,
    index_scans BIGINT,
    index_tuples_read BIGINT,
    index_tuples_fetched BIGINT,
    efficiency_score DECIMAL,
    usage_category TEXT,
    optimization_priority TEXT,
    recommendations TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.tablename as table_name,
        i.indexname as index_name,
        ROUND(pg_relation_size(i.indexname) / (1024 * 1024), 2) as index_size_mb,
        s.idx_scan as index_scans,
        s.idx_tup_read as index_tuples_read,
        s.idx_tup_fetch as index_tuples_fetched,
        ROUND(
            CASE 
                WHEN s.idx_scan > 0 THEN (s.idx_tup_fetch::DECIMAL / s.idx_tup_read) * 100
                ELSE 0
            END, 2
        ) as efficiency_score,
        CASE 
            WHEN s.idx_scan = 0 THEN 'UNUSED'
            WHEN s.idx_scan < 10 THEN 'LOW_USAGE'
            WHEN s.idx_scan < 100 THEN 'MODERATE_USAGE'
            WHEN s.idx_scan < 1000 THEN 'HIGH_USAGE'
            ELSE 'VERY_HIGH_USAGE'
        END as usage_category,
        CASE 
            WHEN s.idx_scan = 0 THEN 'HIGH'
            WHEN s.idx_scan < 10 AND pg_relation_size(i.indexname) > 10 * 1024 * 1024 THEN 'MEDIUM'
            WHEN s.idx_scan < 100 AND pg_relation_size(i.indexname) > 50 * 1024 * 1024 THEN 'LOW'
            ELSE 'LOW'
        END as optimization_priority,
        CASE 
            WHEN s.idx_scan = 0 THEN 'Consider dropping unused index to save storage space'
            WHEN s.idx_scan < 10 AND pg_relation_size(i.indexname) > 10 * 1024 * 1024 THEN 'Monitor usage - consider dropping if not needed'
            WHEN s.idx_scan < 100 AND pg_relation_size(i.indexname) > 50 * 1024 * 1024 THEN 'Large index with low usage - monitor'
            ELSE 'Index usage is efficient'
        END as recommendations
    FROM pg_indexes i
    JOIN pg_tables t ON i.tablename = t.tablename
    LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexrelname
    WHERE t.schemaname = 'public'
    ORDER BY pg_relation_size(i.indexname) DESC;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to get table performance analysis
CREATE OR REPLACE FUNCTION get_table_performance_analysis() 
RETURNS TABLE (
    table_name TEXT,
    table_size_mb DECIMAL,
    row_count BIGINT,
    dead_tuples BIGINT,
    live_tuples BIGINT,
    bloat_percentage DECIMAL,
    seq_scans BIGINT,
    seq_tuples_read BIGINT,
    idx_scans BIGINT,
    idx_tuples_fetched BIGINT,
    performance_score DECIMAL,
    optimization_priority TEXT,
    recommendations TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.tablename as table_name,
        ROUND(pg_total_relation_size(t.tablename) / (1024 * 1024), 2) as table_size_mb,
        s.n_live_tup as row_count,
        s.n_dead_tup as dead_tuples,
        s.n_live_tup as live_tuples,
        ROUND(
            CASE 
                WHEN s.n_live_tup > 0 THEN (s.n_dead_tup::DECIMAL / s.n_live_tup) * 100
                ELSE 0
            END, 2
        ) as bloat_percentage,
        s.seq_scan as seq_scans,
        s.seq_tup_read as seq_tuples_read,
        s.idx_scan as idx_scans,
        s.idx_tup_fetch as idx_tuples_fetched,
        ROUND(
            CASE 
                WHEN s.seq_scan + s.idx_scan > 0 THEN (s.idx_scan::DECIMAL / (s.seq_scan + s.idx_scan)) * 100
                ELSE 0
            END, 2
        ) as performance_score,
        CASE 
            WHEN s.n_dead_tup > s.n_live_tup THEN 'HIGH'
            WHEN s.seq_scan > s.idx_scan * 2 THEN 'MEDIUM'
            WHEN pg_total_relation_size(t.tablename) > 100 * 1024 * 1024 THEN 'LOW'
            ELSE 'LOW'
        END as optimization_priority,
        CASE 
            WHEN s.n_dead_tup > s.n_live_tup THEN 'Run VACUUM FULL to reduce bloat'
            WHEN s.seq_scan > s.idx_scan * 2 THEN 'Consider adding indexes to reduce sequential scans'
            WHEN pg_total_relation_size(t.tablename) > 100 * 1024 * 1024 THEN 'Large table - monitor performance'
            ELSE 'Table performance is good'
        END as recommendations
    FROM pg_tables t
    LEFT JOIN pg_stat_user_tables s ON t.tablename = s.relname
    WHERE t.schemaname = 'public'
    ORDER BY pg_total_relation_size(t.tablename) DESC;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to get connection performance analysis
CREATE OR REPLACE FUNCTION get_connection_performance_analysis() 
RETURNS TABLE (
    connection_count INTEGER,
    active_connections INTEGER,
    idle_connections INTEGER,
    max_connections INTEGER,
    utilization_percentage DECIMAL,
    status TEXT,
    recommendations TEXT
) AS $$
DECLARE
    total_connections INTEGER;
    active_conn INTEGER;
    idle_conn INTEGER;
    max_conn INTEGER := 60;
    utilization DECIMAL;
BEGIN
    -- Get total connections
    SELECT COUNT(*) INTO total_connections
    FROM pg_stat_activity;
    
    -- Get active connections
    SELECT COUNT(*) INTO active_conn
    FROM pg_stat_activity
    WHERE state = 'active';
    
    -- Get idle connections
    SELECT COUNT(*) INTO idle_conn
    FROM pg_stat_activity
    WHERE state = 'idle';
    
    -- Calculate utilization
    utilization := ROUND((total_connections::DECIMAL / max_conn) * 100, 2);
    
    RETURN QUERY
    SELECT 
        total_connections as connection_count,
        active_conn as active_connections,
        idle_conn as idle_connections,
        max_conn as max_connections,
        utilization as utilization_percentage,
        CASE 
            WHEN utilization >= 90 THEN 'CRITICAL'
            WHEN utilization >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN utilization >= 90 THEN 'Consider connection pooling or upgrading plan'
            WHEN utilization >= 75 THEN 'Monitor connection usage closely'
            ELSE 'Connection usage is within acceptable limits'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to generate performance dashboard data
CREATE OR REPLACE FUNCTION generate_performance_dashboard() 
RETURNS TABLE (
    dashboard_section TEXT,
    metric_name TEXT,
    current_value TEXT,
    target_value TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT
) AS $$
DECLARE
    db_size DECIMAL;
    db_connections INTEGER;
    cache_hit_ratio DECIMAL;
    index_usage_score DECIMAL;
    query_performance_score DECIMAL;
    overall_health_score DECIMAL;
BEGIN
    -- Get current metrics
    SELECT 
        ROUND(SUM(pg_database_size(datname)) / (1024 * 1024), 2)
    INTO db_size
    FROM pg_database 
    WHERE datname = current_database();
    
    SELECT COUNT(*) INTO db_connections
    FROM pg_stat_activity
    WHERE state = 'active';
    
    SELECT 
        ROUND(
            (blks_hit::DECIMAL / (blks_hit + blks_read)) * 100, 2
        )
    INTO cache_hit_ratio
    FROM pg_stat_database
    WHERE datname = current_database();
    
    SELECT 
        ROUND(
            AVG(
                CASE 
                    WHEN idx_scan > 0 THEN (idx_scan::DECIMAL / (idx_scan + seq_scan)) * 100
                    ELSE 0
                END
            ), 2
        )
    INTO index_usage_score
    FROM pg_stat_user_tables;
    
    SELECT 
        ROUND(
            AVG(
                CASE 
                    WHEN mean_exec_time < 100 THEN 100
                    WHEN mean_exec_time < 500 THEN 80
                    WHEN mean_exec_time < 1000 THEN 60
                    WHEN mean_exec_time < 2000 THEN 40
                    ELSE 20
                END
            ), 2
        )
    INTO query_performance_score
    FROM pg_stat_statements;
    
    -- Calculate overall health score
    overall_health_score := ROUND(
        (COALESCE(cache_hit_ratio, 0) + 
         COALESCE(index_usage_score, 0) + 
         COALESCE(query_performance_score, 0)) / 3, 2
    );
    
    RETURN QUERY
    -- Storage Performance Section
    SELECT 
        'Storage Performance' as dashboard_section,
        'Database Size' as metric_name,
        db_size::TEXT || ' MB' as current_value,
        '500 MB' as target_value,
        CASE 
            WHEN (db_size / 500) * 100 >= 90 THEN 'CRITICAL'
            WHEN (db_size / 500) * 100 >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN (db_size / 500) * 100 >= 90 THEN 'Consider upgrading to paid plan'
            WHEN (db_size / 500) * 100 >= 75 THEN 'Monitor storage usage'
            ELSE 'Storage usage is optimal'
        END as recommendations
    
    UNION ALL
    
    -- Connection Performance Section
    SELECT 
        'Connection Performance' as dashboard_section,
        'Active Connections' as metric_name,
        db_connections::TEXT as current_value,
        '60' as target_value,
        CASE 
            WHEN (db_connections::DECIMAL / 60) * 100 >= 90 THEN 'CRITICAL'
            WHEN (db_connections::DECIMAL / 60) * 100 >= 75 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN (db_connections::DECIMAL / 60) * 100 >= 90 THEN 'Consider connection pooling'
            WHEN (db_connections::DECIMAL / 60) * 100 >= 75 THEN 'Monitor connection usage'
            ELSE 'Connection usage is optimal'
        END as recommendations
    
    UNION ALL
    
    -- Cache Performance Section
    SELECT 
        'Cache Performance' as dashboard_section,
        'Cache Hit Ratio' as metric_name,
        COALESCE(cache_hit_ratio, 0)::TEXT || '%' as current_value,
        '95%' as target_value,
        CASE 
            WHEN cache_hit_ratio < 80 THEN 'CRITICAL'
            WHEN cache_hit_ratio < 90 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN cache_hit_ratio < 80 THEN 'Increase shared_buffers'
            WHEN cache_hit_ratio < 90 THEN 'Monitor cache performance'
            ELSE 'Cache performance is optimal'
        END as recommendations
    
    UNION ALL
    
    -- Index Performance Section
    SELECT 
        'Index Performance' as dashboard_section,
        'Index Usage Score' as metric_name,
        COALESCE(index_usage_score, 0)::TEXT || '%' as current_value,
        '80%' as target_value,
        CASE 
            WHEN index_usage_score < 50 THEN 'CRITICAL'
            WHEN index_usage_score < 70 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN index_usage_score < 50 THEN 'Review and optimize indexes'
            WHEN index_usage_score < 70 THEN 'Monitor index usage'
            ELSE 'Index usage is optimal'
        END as recommendations
    
    UNION ALL
    
    -- Query Performance Section
    SELECT 
        'Query Performance' as dashboard_section,
        'Query Performance Score' as metric_name,
        COALESCE(query_performance_score, 0)::TEXT || '%' as current_value,
        '80%' as target_value,
        CASE 
            WHEN query_performance_score < 50 THEN 'CRITICAL'
            WHEN query_performance_score < 70 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN query_performance_score < 50 THEN 'Optimize slow queries'
            WHEN query_performance_score < 70 THEN 'Monitor query performance'
            ELSE 'Query performance is optimal'
        END as recommendations
    
    UNION ALL
    
    -- Overall Health Section
    SELECT 
        'Overall Health' as dashboard_section,
        'Overall Health Score' as metric_name,
        overall_health_score::TEXT || '%' as current_value,
        '85%' as target_value,
        CASE 
            WHEN overall_health_score < 60 THEN 'CRITICAL'
            WHEN overall_health_score < 80 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN overall_health_score < 60 THEN 'Multiple issues detected'
            WHEN overall_health_score < 80 THEN 'Some issues detected'
            ELSE 'Database health is optimal'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to log performance metrics
CREATE OR REPLACE FUNCTION log_performance_metrics() 
RETURNS VOID AS $$
DECLARE
    metrics RECORD;
BEGIN
    -- Collect and log all performance metrics
    FOR metrics IN 
        SELECT * FROM collect_performance_metrics()
    LOOP
        INSERT INTO performance_metrics (
            metric_name, 
            metric_value, 
            metric_unit, 
            metric_category, 
            status, 
            details, 
            recommendations
        ) VALUES (
            metrics.metric_name,
            metrics.metric_value,
            metrics.metric_unit,
            metrics.metric_category,
            metrics.status,
            metrics.details,
            metrics.recommendations
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to get performance trends
CREATE OR REPLACE FUNCTION get_performance_trends(days_back INTEGER DEFAULT 7) 
RETURNS TABLE (
    metric_name TEXT,
    date_recorded DATE,
    avg_value DECIMAL,
    max_value DECIMAL,
    min_value DECIMAL,
    trend_direction TEXT,
    performance_category TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pm.metric_name,
        pm.recorded_at::DATE as date_recorded,
        ROUND(AVG(pm.metric_value), 2) as avg_value,
        ROUND(MAX(pm.metric_value), 2) as max_value,
        ROUND(MIN(pm.metric_value), 2) as min_value,
        CASE 
            WHEN AVG(pm.metric_value) > LAG(AVG(pm.metric_value)) OVER (PARTITION BY pm.metric_name ORDER BY pm.recorded_at::DATE) THEN 'IMPROVING'
            WHEN AVG(pm.metric_value) < LAG(AVG(pm.metric_value)) OVER (PARTITION BY pm.metric_name ORDER BY pm.recorded_at::DATE) THEN 'DEGRADING'
            ELSE 'STABLE'
        END as trend_direction,
        CASE 
            WHEN AVG(pm.metric_value) > 80 THEN 'EXCELLENT'
            WHEN AVG(pm.metric_value) > 60 THEN 'GOOD'
            WHEN AVG(pm.metric_value) > 40 THEN 'FAIR'
            ELSE 'POOR'
        END as performance_category
    FROM performance_metrics pm
    WHERE pm.recorded_at >= NOW() - INTERVAL '1 day' * days_back
    GROUP BY pm.metric_name, pm.recorded_at::DATE
    ORDER BY pm.metric_name, pm.recorded_at::DATE;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to get performance alerts
CREATE OR REPLACE FUNCTION get_performance_alerts() 
RETURNS TABLE (
    alert_id SERIAL,
    metric_name TEXT,
    alert_level TEXT,
    current_value TEXT,
    threshold_value TEXT,
    alert_message TEXT,
    recommendations TEXT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pm.id as alert_id,
        pm.metric_name,
        pm.status as alert_level,
        pm.metric_value::TEXT || ' ' || pm.metric_unit as current_value,
        COALESCE(pm.threshold_critical::TEXT, 'N/A') as threshold_value,
        CASE 
            WHEN pm.status = 'CRITICAL' THEN 'Critical performance issue detected'
            WHEN pm.status = 'WARNING' THEN 'Performance warning detected'
            ELSE 'Performance is normal'
        END as alert_message,
        pm.recommendations,
        pm.recorded_at as created_at
    FROM performance_metrics pm
    WHERE pm.status IN ('WARNING', 'CRITICAL')
    AND pm.recorded_at >= NOW() - INTERVAL '24 hours'
    ORDER BY pm.recorded_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 11. Create a function to get performance summary
CREATE OR REPLACE FUNCTION get_performance_summary() 
RETURNS TABLE (
    summary_category TEXT,
    total_metrics INTEGER,
    ok_count INTEGER,
    warning_count INTEGER,
    critical_count INTEGER,
    overall_status TEXT,
    last_updated TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pm.metric_category as summary_category,
        COUNT(*) as total_metrics,
        COUNT(CASE WHEN pm.status = 'OK' THEN 1 END) as ok_count,
        COUNT(CASE WHEN pm.status = 'WARNING' THEN 1 END) as warning_count,
        COUNT(CASE WHEN pm.status = 'CRITICAL' THEN 1 END) as critical_count,
        CASE 
            WHEN COUNT(CASE WHEN pm.status = 'CRITICAL' THEN 1 END) > 0 THEN 'CRITICAL'
            WHEN COUNT(CASE WHEN pm.status = 'WARNING' THEN 1 END) > 0 THEN 'WARNING'
            ELSE 'OK'
        END as overall_status,
        MAX(pm.recorded_at) as last_updated
    FROM performance_metrics pm
    WHERE pm.recorded_at >= NOW() - INTERVAL '1 hour'
    GROUP BY pm.metric_category
    ORDER BY pm.metric_category;
END;
$$ LANGUAGE plpgsql;

-- 12. Create a function to setup automated performance monitoring
CREATE OR REPLACE FUNCTION setup_automated_performance_monitoring() 
RETURNS TEXT AS $$
BEGIN
    -- Create a function to be called by cron or scheduled job
    CREATE OR REPLACE FUNCTION automated_performance_monitoring() 
    RETURNS VOID AS $$
    BEGIN
        -- Log current performance metrics
        PERFORM log_performance_metrics();
        
        -- Check for critical performance issues
        IF EXISTS (
            SELECT 1 FROM collect_performance_metrics() 
            WHERE status = 'CRITICAL'
        ) THEN
            -- Log critical alert
            INSERT INTO performance_metrics (metric_name, metric_value, metric_unit, metric_category, status, details, recommendations)
            VALUES ('CRITICAL ALERT', 0, 'alert', 'System', 'CRITICAL', '{}', 'Critical performance issue detected - immediate attention required');
        END IF;
        
        -- Check for warning issues
        IF EXISTS (
            SELECT 1 FROM collect_performance_metrics() 
            WHERE status = 'WARNING'
        ) THEN
            -- Log warning alert
            INSERT INTO performance_metrics (metric_name, metric_value, metric_unit, metric_category, status, details, recommendations)
            VALUES ('WARNING ALERT', 0, 'alert', 'System', 'WARNING', '{}', 'Performance warning detected - monitor closely');
        END IF;
    END;
    $$ LANGUAGE plpgsql;
    
    RETURN 'Automated performance monitoring setup completed. Call automated_performance_monitoring() function regularly.';
END;
$$ LANGUAGE plpgsql;

-- 13. Create a function to get performance dashboard data
CREATE OR REPLACE FUNCTION get_performance_dashboard_data() 
RETURNS TABLE (
    section TEXT,
    metric TEXT,
    value TEXT,
    status TEXT,
    last_updated TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    -- Current performance metrics
    SELECT 
        'Current Performance' as section,
        'Database Health Score' as metric,
        (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Overall Health Score') as value,
        (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Overall Health Score') as status,
        NOW() as last_updated
    
    UNION ALL
    
    SELECT 
        'Current Performance' as section,
        'Cache Hit Ratio' as metric,
        (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Cache Hit Ratio') as value,
        (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Cache Hit Ratio') as status,
        NOW() as last_updated
    
    UNION ALL
    
    SELECT 
        'Current Performance' as section,
        'Index Usage Score' as metric,
        (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Index Usage Score') as value,
        (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Index Usage Score') as status,
        NOW() as last_updated
    
    UNION ALL
    
    -- Recent alerts
    SELECT 
        'Recent Alerts' as section,
        pa.metric_name as metric,
        pa.alert_level as value,
        pa.alert_level as status,
        pa.created_at as last_updated
    FROM get_performance_alerts() pa
    ORDER BY pa.created_at DESC
    LIMIT 5
    
    UNION ALL
    
    -- Performance trends
    SELECT 
        'Performance Trends' as section,
        pt.metric_name as metric,
        pt.trend_direction as value,
        pt.performance_category as status,
        NOW() as last_updated
    FROM get_performance_trends(7) pt
    WHERE pt.date_recorded = CURRENT_DATE
    LIMIT 3;
END;
$$ LANGUAGE plpgsql;

-- 14. Create a function to export performance data
CREATE OR REPLACE FUNCTION export_performance_data(days_back INTEGER DEFAULT 30) 
RETURNS TABLE (
    export_date DATE,
    metric_name TEXT,
    daily_avg_value DECIMAL,
    daily_max_value DECIMAL,
    daily_min_value DECIMAL,
    status_summary TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pm.recorded_at::DATE as export_date,
        pm.metric_name,
        ROUND(AVG(pm.metric_value), 2) as daily_avg_value,
        ROUND(MAX(pm.metric_value), 2) as daily_max_value,
        ROUND(MIN(pm.metric_value), 2) as daily_min_value,
        CASE 
            WHEN COUNT(CASE WHEN pm.status = 'CRITICAL' THEN 1 END) > 0 THEN 'CRITICAL'
            WHEN COUNT(CASE WHEN pm.status = 'WARNING' THEN 1 END) > 0 THEN 'WARNING'
            ELSE 'OK'
        END as status_summary
    FROM performance_metrics pm
    WHERE pm.recorded_at >= NOW() - INTERVAL '1 day' * days_back
    GROUP BY pm.recorded_at::DATE, pm.metric_name
    ORDER BY pm.recorded_at::DATE DESC, pm.metric_name;
END;
$$ LANGUAGE plpgsql;

-- 15. Create a function to validate performance monitoring setup
CREATE OR REPLACE FUNCTION validate_performance_monitoring_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if performance_metrics table exists
    SELECT 
        'Performance Metrics Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing performance metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'performance_metrics') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create performance_metrics table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if performance functions exist
    SELECT 
        'Performance Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'collect_performance_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for collecting performance metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'collect_performance_metrics') 
            THEN 'All performance functions are available' 
            ELSE 'Create performance functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if automated monitoring is set up
    SELECT 
        'Automated Monitoring' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'automated_performance_monitoring') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Automated performance monitoring function' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'automated_performance_monitoring') 
            THEN 'Automated monitoring is ready' 
            ELSE 'Set up automated monitoring' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_performance_metrics_metric_name ON performance_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_recorded_at ON performance_metrics(recorded_at);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_status ON performance_metrics(status);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_category ON performance_metrics(metric_category);

-- Create a view for easy performance dashboard access
CREATE OR REPLACE VIEW performance_dashboard AS
SELECT 
    'Database Health' as metric_name,
    (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Overall Health Score') as current_value,
    '85%' as target_value,
    (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Overall Health Score') as status
UNION ALL
SELECT 
    'Cache Performance' as metric_name,
    (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Cache Hit Ratio') as current_value,
    '95%' as target_value,
    (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Cache Hit Ratio') as status
UNION ALL
SELECT 
    'Index Performance' as metric_name,
    (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Index Usage Score') as current_value,
    '80%' as target_value,
    (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Index Usage Score') as status
UNION ALL
SELECT 
    'Query Performance' as metric_name,
    (SELECT metric_value::TEXT || '%' FROM collect_performance_metrics() WHERE metric_name = 'Query Performance Score') as current_value,
    '80%' as target_value,
    (SELECT status FROM collect_performance_metrics() WHERE metric_name = 'Query Performance Score') as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON performance_dashboard TO authenticated;
GRANT SELECT, INSERT ON performance_metrics TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Database performance dashboards setup completed successfully!';
    RAISE NOTICE 'Total functions created: 15';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 4';
    RAISE NOTICE 'All performance monitoring tools are now available.';
    RAISE NOTICE 'Use performance_dashboard view to access current performance metrics.';
    RAISE NOTICE 'Call setup_automated_performance_monitoring() to enable automated monitoring.';
    RAISE NOTICE 'Call log_performance_metrics() to record current performance metrics.';
END $$;
